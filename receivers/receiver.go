package receivers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Receives = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "receive_operation_duration_microseconds",
			Help: "Total duration and counts of receiver methods",
		})
	ReceiveErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receive_operation_errors",
			Help: "Errors encountered while receiving",
		})
)

func init() {
	prometheus.MustRegister(Receives)
	prometheus.MustRegister(ReceiveErrors)
}

// Receiver is an interface whose implementers must provide a generic `Receive()`
// function that sets up a networked listening interface to receive alerts
// from Nagios.
type Receiver interface {
	Handler(http.ResponseWriter, *http.Request)
}
