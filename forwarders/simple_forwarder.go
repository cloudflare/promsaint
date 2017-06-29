package forwarders

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudflare/promsaint/utils"
	prometheus "github.com/prometheus/common/model"
)

var (
	alertManager = flag.String("alertmanager", "http://localhost:9093", "Alertmanager host")
	regex2xx     = regexp.MustCompile(`^2..`)
)

type SimpleForwarder struct{}

func NewSimpleForwarder() *SimpleForwarder {
	return &SimpleForwarder{}
}

func (forwarder *SimpleForwarder) Send(alerts []prometheus.Alert) {
	defer utils.TimeIt(time.Now(), Forwards)

	u, err := url.Parse(*alertManager)
	if err != nil {
		log.Error(err)
		ForwardErrors.Inc()
		return
	}
	u.Path = path.Join(u.Path, "/api/v1/alerts")
	log.Debugf("Sending to %s", u.String())

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(alerts)
	if err != nil {
		log.Error(err)
		ForwardErrors.Inc()
		return
	}

	req, err := http.NewRequest("POST", u.String(), b)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		ForwardErrors.Inc()
		return
	}

	defer resp.Body.Close()

	status := fmt.Sprintf("%d", resp.StatusCode)
	log.Debugf("Status: %s", status)
	if !regex2xx.Match([]byte(status)) {
		log.Errorf("Alertmanager responded with non 2xx error: %s", resp.Status)
		ForwardErrors.Inc()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Debugf("Alertmanager response:%s", string(body))
	}

}
