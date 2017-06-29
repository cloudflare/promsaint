package backends

import (
	"net/http"

	"github.com/cloudflare/promsaint/models"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	DatabaseTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_total",
			Help: "Total number of items in the database",
		})
	DatabaseFiring = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_firing",
			Help: "Total number of firing alerts in the database",
		})
	DatabaseResolved = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_resolved",
			Help: "Total number of resolved alerts in the database",
		})
	DatabasePruned = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_pruned",
			Help: "Total number of alerts pruned",
		},
		[]string{"reason"},
	)
	Operations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "database_operation_duration_microseconds",
			Help: "Total duration and counts of database methods",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(DatabaseTotal)
	prometheus.MustRegister(DatabaseFiring)
	prometheus.MustRegister(DatabaseResolved)
	prometheus.MustRegister(DatabasePruned)
	prometheus.MustRegister(Operations)
}

type Backend interface {
	Import(*models.Alert)
	Export(models.NotificationSender)
	Prune()
	CalculateLastAlert() float64
	DebugHandler(w http.ResponseWriter, r *http.Request)
}
