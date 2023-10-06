package messages_test

import (
	"testing"

	"github.com/burke/zeus/go/messages"
)

func TestCreateCommandAndArgumentsMessage(t *testing.T) {
	message := messages.CreateCommandAndArgumentsMessage([]string{"arg1", "arg2"}, 100)
	if message != "T:1:100:arg1" {
		t.Fatal(message)
	}
}

func TestCreateSpawnSlaveMessage(t *testing.T) {
	message := messages.CreateSpawnSlaveMessage("decimate")
	if message != "S:decimate" {
		t.Fatal(message)
	}
}

func TestCreateSpawnCommandMessage(t *testing.T) {
	message := messages.CreateSpawnCommandMessage("decimate")
	if message != "C:decimate" {
		t.Fatal(message)
	}
}
