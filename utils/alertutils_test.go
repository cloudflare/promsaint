package utils

import (
	"testing"

	"github.com/cloudflare/promsaint/models"
)

var testAlert = &models.Alert{
	Type:             "service",
	Host:             "foo",
	Service:          "bar",
	Notify:           "hipchat-bar",
	NotificationType: "PROBLEM",
	State:            "CRITICAL",
	Message:          "It's dead jim",
	Note:             "http://reddit.com/",
}

var testAlert2 = &models.Alert{
	Type:             "service",
	Host:             "bar",
	Service:          "foo",
	Notify:           "hipchat-bar",
	NotificationType: "PROBLEM",
	State:            "CRITICAL",
	Message:          "It's dead jim",
	Note:             "http://reddit.com/",
}

var testAlert3 = &models.Alert{
	Type:             "service",
	Host:             "bar",
	Service:          "foo",
	Notify:           "hipchat-cafe",
	NotificationType: "PROBLEM",
	State:            "CRITICAL",
	Message:          "It's dead jim",
	Note:             "http://reddit.com/",
}

func TestKey(t *testing.T) {
	key := Key(testAlert)
	key2 := Key(testAlert)

	key3 := Key(testAlert2)

	if key != key2 {
		t.Fatal("Hashes for same object mismatch")
	}

	if key == key3 {
		t.Fatal("Hashes match for different object")
	}

	if got := Key(testAlert3); got != key3 {
		t.Fatalf("Hashes should have matched. got %#v, want %#v", got, key3)
	}
}

func TestMerge(t *testing.T) {
	pAlert := models.InternalAlert{}
	Merge(&pAlert, testAlert2)
	if got := pAlert.PrometheusAlert.Labels["notify"]; string(got) != testAlert.Notify {
		t.Errorf("Merge failed: got %#v, want %#v", got, testAlert.Notify)
	}

	// now test merging a new field:
	Merge(&pAlert, testAlert3)
	want := "hipchat-bar hipchat-cafe"
	if got := pAlert.PrometheusAlert.Labels["notify"]; string(got) != want {
		t.Errorf("Merge failed: got %#v, want %#v", got, want)
	}
}
