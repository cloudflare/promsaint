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
}
