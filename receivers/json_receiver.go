package receivers

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/cloudflare/promsaint/models"
	"github.com/cloudflare/promsaint/utils"
)

// JsonReceiver implements the Receiver interface by providing an HTTP POST
// handler that receives content formatted for Json's API.
type JsonReceiver struct {
	alertReceived chan<- models.Alert
}

// NewJsonReceiver initializes an empty `JsonReceiver` and returns a pointer to it.
//
// It expects a response channel to be passed in for parsed `Alert` data structures
// to be placed onto.
func NewJsonReceiver(alertReceived chan<- models.Alert) *JsonReceiver {
	return &JsonReceiver{alertReceived: alertReceived}
}

func (h *JsonReceiver) Handler(w http.ResponseWriter, r *http.Request) {
	defer utils.TimeIt(time.Now(), Receives)
	msg := &models.Alert{}
	err := json.NewDecoder(r.Body).Decode(msg)
	if err != nil {
		log.Warnf("WARN: JSON failed to decode: %s", err)
		ReceiveErrors.Inc()
		return
	}

	h.alertReceived <- *msg
}
