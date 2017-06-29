package backends

import (
	"testing"

	"github.com/cloudflare/promsaint/models"
	prometheus "github.com/prometheus/common/model"
)

var testingAlert = models.Alert{
	Service:          "Syslog Dropped Logs",
	Message:          "syslogng.dst.network/type=dropped = 0 :: syslogng.dst.network/type=dropped = 0 :: syslogng.dst.network/type=dropped = 0 :: syslogng.dst.network/type=dropped = 0",
	NotificationType: "PROBLEM/CRITICAL",
	Host:             "TestInstance1",
}

func TestBasicImport(t *testing.T) {
	backend := NewBasicBackend()
	backend.Import(&testingAlert)
}

type dummy struct {
	pipe chan prometheus.Alert
}

func (d *dummy) Send(alerts []prometheus.Alert) {
	for _, alert := range alerts {
		d.pipe <- alert
	}
}

func TestBasicExport(t *testing.T) {

	dummyInstance := &dummy{pipe: make(chan prometheus.Alert, 1)}
	backend := NewBasicBackend()
	backend.Import(&testingAlert)
	backend.Export(dummyInstance)

	alert := <-dummyInstance.pipe

	t.Logf("%v", alert)
}

func TestBasicPrune(t *testing.T) {
	backend := NewBasicBackend()
	backend.Import(&testingAlert)
	backend.Prune()

}
