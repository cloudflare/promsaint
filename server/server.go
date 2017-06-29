package server

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudflare/promsaint/backends"
	"github.com/cloudflare/promsaint/forwarders"
	"github.com/cloudflare/promsaint/models"
	"github.com/cloudflare/promsaint/receivers"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	pruneInterval   = flag.Duration("pruneinterval", time.Minute*2, "How often to prune the database")
	publishInterval = flag.Duration("publishinterval", time.Second*60, "How often to publish the database")
	publishMinimum  = flag.Duration("publishminimum", time.Second*5, "Minimum time to wait between alertmanager pushes")
	listenAddr      = flag.String("listen", ":8080", "HTTP endpoint")
)

type PromsaintServer struct {
	alertChan        chan models.Alert
	lastNotification time.Time
	skippedExport    bool
	publishTicker    *time.Ticker
	pruneTicker      *time.Ticker
	backend          backends.Backend
	forwarder        forwarders.Forwarder
	receiver         receivers.Receiver
}

func NewPromsaint() *PromsaintServer {
	server := &PromsaintServer{
		alertChan:     make(chan models.Alert, 100),
		pruneTicker:   time.NewTicker(*pruneInterval),
		publishTicker: time.NewTicker(time.Millisecond * 500),
		backend:       backends.NewBasicBackend(),
		forwarder:     forwarders.NewSimpleForwarder(),
	}
	server.receiver = receivers.NewJsonReceiver(server.alertChan)
	return server
}

func (server *PromsaintServer) pruneLoop() {
	log.Info("Starting Pruner")
	for range server.pruneTicker.C {
		log.Info("Running Prune Operation")
		server.backend.Prune()
	}
}

func (server *PromsaintServer) forwardLoop() {
	log.Info("Starting forwarder")
	for {
		select {
		case alert := <-server.alertChan:
			server.backend.Import(&alert)
			if time.Since(server.lastNotification) > *publishMinimum {
				server.backend.Export(server.forwarder)
				log.Info("Running event based publish")
				server.lastNotification = time.Now().UTC()
			} else {
				server.skippedExport = true
			}
		case <-server.publishTicker.C:
			if (server.skippedExport &&
				time.Since(server.lastNotification) > *publishMinimum) ||
				time.Since(server.lastNotification) > *publishInterval {
				log.Info("Running time based publish")
				server.backend.Export(server.forwarder)
				server.lastNotification = time.Now().UTC()
				server.skippedExport = false
			}
		}
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

// This is where we setup metrics that need to be updated everytime the metrics endpoint is called
func (server *PromsaintServer) registerDynamicMetrics() {
	LastAlert := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "last_alert_seconds",
			Help: "Number of seconds since the last alert was sent from nagios",
		},
		server.backend.CalculateLastAlert,
	)
	prometheus.MustRegister(LastAlert)
}

func (server *PromsaintServer) Start() {
	log.Info("Starting Up")

	// Register the dynamic metrics
	server.registerDynamicMetrics()

	// Start the forwarder loop
	go server.forwardLoop()

	// Start the pruner
	go server.pruneLoop()

	// Setup Http server
	// Interestingly / matches everything... So lets give this a name
	http.HandleFunc("/json", server.receiver.Handler)
	http.HandleFunc("/debug", server.backend.DebugHandler)
	http.HandleFunc("/status", statusHandler)
	http.Handle("/metrics", prometheus.Handler())

	log.WithFields(log.Fields{
		"addr": *listenAddr,
	}).Warn("HTTP Endpoint starting")
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
