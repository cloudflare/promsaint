package forwarders

import (
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	prometheus "github.com/prometheus/common/model"
)

var (
	Forwards = prometheus_client.NewSummary(
		prometheus_client.SummaryOpts{
			Name: "forward_operation_duration_microseconds",
			Help: "Total duration and counts of forwarder methods",
		})
	ForwardErrors = prometheus_client.NewCounter(
		prometheus_client.CounterOpts{
			Name: "forward_operation_errors",
			Help: "Errors encountered while forwarding",
		})
)

func init() {
	prometheus_client.MustRegister(Forwards)
	prometheus_client.MustRegister(ForwardErrors)
}

type Forwarder interface {
	Send([]prometheus.Alert)
}
