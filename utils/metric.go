package utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TimeIt(start time.Time, metric prometheus.Summary) {
	elapsed := time.Since(start)
	// Create Microseconds
	metric.Observe(float64(elapsed.Nanoseconds() / 1e3))
}
