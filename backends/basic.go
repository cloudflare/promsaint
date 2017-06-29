package backends

import (
	"encoding/json"
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/cloudflare/promsaint/models"
	"github.com/cloudflare/promsaint/utils"
	log "github.com/Sirupsen/logrus"
	prometheus "github.com/prometheus/common/model"
)

var pruneTimeout = flag.Duration("pruneage", time.Second*60, "How old of alerts to keep around")
var forgetTimeout = flag.Duration("forgetage", time.Second*60, "How long to persist an alert after nagios forgets about it before removing it")

type BasicBackend struct {
	data       map[string]models.InternalAlert
	lock       *sync.Mutex
	lastUpdate time.Time
}

func NewBasicBackend() *BasicBackend {
	return &BasicBackend{
		data:       map[string]models.InternalAlert{},
		lock:       &sync.Mutex{},
		lastUpdate: time.Now().UTC(),
	}
}

// Add a Nagios Alert
func (set *BasicBackend) Import(alert *models.Alert) {
	defer utils.TimeIt(time.Now(), Operations.WithLabelValues("import"))

	// Bump lastUpdate
	set.lastUpdate = time.Now().UTC()

	key := utils.Key(alert)
	log.Debugf("Doing Import of %s: %+v", key, alert)

	// Lock the dataset for insert delete
	// At some point we will probably want to lock items but this will work for now AFAIK
	set.lock.Lock()
	defer set.lock.Unlock()

	pAlert, ok := set.data[key]
	// Alert Exists. Merge the summary
	if !ok {
		log.Debugln("No existing alert creating new entry")
		pAlert = models.InternalAlert{}
	}

	utils.Merge(&pAlert, alert)

	set.data[key] = pAlert
	set.collectMetrics()
}

// Build list of alerts to send to alertmanager
// Pass the send function here so that we can manage locks from this
func (set *BasicBackend) Export(notifier models.NotificationSender) {
	defer utils.TimeIt(time.Now(), Operations.WithLabelValues("export"))
	set.lock.Lock()
	defer set.lock.Unlock()

	log.Debugln("Running Export")
	output := make([]prometheus.Alert, 0)
	for _, value := range set.data {
		output = append(output, value.PrometheusAlert)
	}
	notifier.Send(output)
	set.collectMetrics()
}

func (set *BasicBackend) Prune() {
	defer utils.TimeIt(time.Now(), Operations.WithLabelValues("prune"))

	set.lock.Lock()
	defer set.lock.Unlock()

	log.Debugln("Running Prune")
	for key, value := range set.data {
		if !value.PrometheusAlert.EndsAt.IsZero() && time.Now().UTC().Sub(value.PrometheusAlert.EndsAt) > *pruneTimeout {
			DatabasePruned.WithLabelValues("resolved").Inc()
			delete(set.data, key)
		} else if time.Now().UTC().Sub(value.Metadata.LastUpdate) > *forgetTimeout {
			DatabasePruned.WithLabelValues("forgotten").Inc()
			delete(set.data, key)
		}
	}
	set.collectMetrics()
}

func (set *BasicBackend) collectMetrics() {
	total, firing, resolved := 0, 0, 0
	for _, value := range set.data {
		if value.PrometheusAlert.EndsAt.IsZero() {
			firing = firing + 1
		} else {
			resolved = resolved + 1
		}
		total = total + 1
	}
	DatabaseTotal.Set(float64(total))
	DatabaseFiring.Set(float64(firing))
	DatabaseResolved.Set(float64(resolved))
}

func (set *BasicBackend) CalculateLastAlert() float64 {
	return time.Since(set.lastUpdate).Seconds()
}

func (set *BasicBackend) DebugHandler(w http.ResponseWriter, r *http.Request) {
	defer utils.TimeIt(time.Now(), Operations.WithLabelValues("debug"))

	set.lock.Lock()
	defer set.lock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(set.data)
	if err != nil {
		log.Error(err)
	}
}
