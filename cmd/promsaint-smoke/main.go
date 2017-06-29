package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	_ "github.com/cloudflare/promsaint/logging"
	prometheus "github.com/prometheus/common/model"
	"github.com/satori/go.uuid"
)

var (
	promsaint     = flag.String("promsaint", "http://localhost:8080", "Promsaint instance to smoketest")
	promsaintCli  = flag.String("cli", "/usr/local/bin/promsaint-cli", "Promsaint-cli path")
	alertmanager  = flag.String("alertmanager", "http://localhost:9093", "The alertmanager to query for the test")
	retryDuration = flag.Duration("retry.duration", time.Second*15, "Time to wait between lookups to alertmanager")
	retryCount    = flag.Int("retry.count", 5, "Number of times to check alertmanager for the alert")
)

type PrometheusResponse struct {
	Status string             `json:"status"`
	Data   []prometheus.Alert `json:"data"`
}

func main() {
	flag.Parse()

	out, err := exec.Command(*promsaintCli, "-version").Output()
	if err != nil {
		log.Fatal(err)
	}

	versionVec := strings.Split(fmt.Sprintf("%s", out), "\n")
	version, buildTime := versionVec[0], versionVec[1]

	log.WithFields(log.Fields{
		"version": version,
		"built":   buildTime,
	}).Info("Found promsaint-cli")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	runId := uuid.NewV4().String()

	arguments := []string{
		"-servicealert",
		"-service", "Promsaint_Smoke_Test",
		"-host", hostname,
		"-msg", runId,
		"-note", "This is a smoke test of the promsaint tool. No action is required",
		"-notify", "blackhole",
		"-promsaint", *promsaint,
	}

	criticalflag := []string{
		"-ntype", "PROBLEM",
		"-state", "CRITICAL",
	}

	recoverflag := []string{
		"-ntype", "RECOVERY",
		"-state", "OK",
	}

	out, err = exec.Command(*promsaintCli, append(arguments, criticalflag...)...).Output()
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"sent":    "CRITICAL",
		"service": "Promsaint_Smoke_Test",
		"runid":   runId,
		"out":     fmt.Sprintf("%s", out),
	}).Info("Dispatched Testing Alert")

	retryTicker := time.NewTicker(*retryDuration)
	retryWaiter := sync.WaitGroup{}
	found := false

	retryWaiter.Add(1)
	go func() {
		defer retryWaiter.Done()
		for i := 0; i < *retryCount; i++ {
			if i != 0 {
				<-retryTicker.C
			}
			alerts := &PrometheusResponse{}
			resp, err := http.Get(fmt.Sprintf("%s/%s", *alertmanager, "api/v1/alerts"))
			if err != nil {
				log.Fatal(err)
			}
			err = json.NewDecoder(resp.Body).Decode(alerts)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()

			for _, alert := range alerts.Data {
				if alert.Labels["alertname"] == "Promsaint_Smoke_Test" && alert.Annotations["summary"] == prometheus.LabelValue(runId) {
					found = true
					return
				}
			}
			log.WithFields(log.Fields{
				"retry": i + 1,
			}).Warn("Unable to find testing alert")
		}
	}()

	retryWaiter.Wait()

	out, err = exec.Command(*promsaintCli, append(arguments, recoverflag...)...).Output()
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"sent":    "RECOVERY",
		"service": "Promsaint_Smoke_Test",
	}).Info("Cleared Testing Alert")
	log.Debug(out)

	if !found {
		log.Fatalf("Unable to find testing alert after %d tries. Exiting.", retryCount)
	} else {
		fmt.Println("[PASS]")
	}
}
