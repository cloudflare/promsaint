package utils

import (
	"crypto/sha1"
	"flag"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudflare/promsaint/models"
	prometheus "github.com/prometheus/common/model"
)

var (
	generatorUrl = flag.String("generator_url", "https://nagios.example.com/nagios/cgi-bin/status.cgi", "Breakdown Source for nagios")
)

func Key(alert *models.Alert) string {
	sum := sha1.Sum(append([]byte(alert.Service), []byte(alert.Host)...))
	str := string(sum[:])
	log.Debugf("AlertKey: %s", str)
	return str
}

/* https://github.com/prometheus/prometheus/blob/master/vendor/github.com/prometheus/common/model/alert.go#L29
 * // Alert is a generic representation of an alert in the Prometheus eco-system.
 * type Alert struct {
 * 	// Label value pairs for purpose of aggregation, matching, and disposition
 * 	// dispatching. This must minimally include an "alertname" label.
 * 	Labels LabelSet `json:"labels"`
 *
 * 	// Extra key/value information which does not define alert identity.
 * 	Annotations LabelSet `json:"annotations"`
 *
 * 	// The known time range for this alert. Both ends are optional.
 * 	StartsAt     time.Time `json:"startsAt,omitempty"`
 * 	EndsAt       time.Time `json:"endsAt,omitempty"`
 * 	GeneratorURL string    `json:"generatorURL"`
 * }
 *
 * type LabelSet map[LabelName]LabelValue
 * type LabelName string
 * type LabelValue string
 */

// Merge a new hipchat alert into an existing prometheus alert (or an empty prometheus struct if the alert doesn't already exist)
func Merge(pAlert *models.InternalAlert, alert *models.Alert) {
	var alertname string
	if alert.Type == "host" {
		alertname = "Host Down"
	} else {
		alertname = alert.Service
	}

	log.Debugf("NOTIFY: %s -> %s", string(pAlert.PrometheusAlert.Labels["notify"]), alert.Notify)
	notifyMap := map[string]bool{}
	if v := pAlert.PrometheusAlert.Labels["notify"]; v != "" {
		for _, value := range strings.Split(string(v), " ") {
			notifyMap[value] = true
		}
	}

	if alert.Notify != "" {
		notifyMap[alert.Notify] = true
	}

	var notifySlice []string
	for key, _ := range notifyMap {
		notifySlice = append(notifySlice, key)
	}

	labels := prometheus.LabelSet{
		"alertname": prometheus.LabelValue(alertname),
		"instance":  prometheus.LabelValue(alert.Host),
		"notify":    prometheus.LabelValue(strings.Join(notifySlice, " ")),
	}

	annotations := prometheus.LabelSet{}
	if alert.Message != "" {
		annotations["summary"] = prometheus.LabelValue(alert.Message)
	}

	if alert.State != "" {
		annotations["state"] = prometheus.LabelValue(alert.State)
	}

	if alert.NotificationType != "" {
		annotations["type"] = prometheus.LabelValue(alert.NotificationType)
	}

	if alert.Note != "" {
		annotations["link"] = prometheus.LabelValue(alert.Note)
	}

	pAlert.PrometheusAlert.Labels = labels
	pAlert.PrometheusAlert.Annotations = annotations

	u, err := url.Parse(*generatorUrl)
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	if alert.Type == "host" {
		q.Set("host", alert.Host)
	} else {
		q.Set("servicegroup", alert.Service)
	}
	q.Set("style", "detail")
	q.Set("limit", "1000")
	u.RawQuery = q.Encode()

	pAlert.PrometheusAlert.GeneratorURL = u.String()

	pAlert.Metadata.LastUpdate = time.Now().UTC()

	if pAlert.PrometheusAlert.StartsAt.IsZero() {
		// Set the start at time to 1s ago to avoid conflicts with recoveries on an empty database
		pAlert.PrometheusAlert.StartsAt = time.Now().UTC().Add(-time.Second)
	}

	if alert.NotificationType == "RECOVERY" {
		pAlert.PrometheusAlert.EndsAt = time.Now().UTC()
	} else {
		// Odd case that we have a recovered alert that fires before we do a prune
		pAlert.PrometheusAlert.EndsAt = time.Time{}
	}
}
