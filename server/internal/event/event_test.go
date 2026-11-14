package event_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"backend/internal/event"
)

func TestNewEvent(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	got, err := event.New("user:user-1", event.TypeConnectionReady, event.ConnectionReady{
		ConnectionID:     "conn-1",
		Username:         "ada",
		HeartbeatSeconds: 25,
		AccessExpiresAt:  now.Add(15 * time.Minute),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != event.Version {
		t.Fatalf("version = %d, want %d", got.Version, event.Version)
	}
	if !strings.HasPrefix(got.ID, "evt_") {
		t.Fatalf("id = %q, want evt_ prefix", got.ID)
	}
	if got.Type != event.TypeConnectionReady {
		t.Fatalf("type = %q", got.Type)
	}
	if got.Topic != "user:user-1" {
		t.Fatalf("topic = %q", got.Topic)
	}
	if !got.Time.Equal(now) {
		t.Fatalf("time = %s", got.Time)
	}

	var payload event.ConnectionReady
	if err := json.Unmarshal(got.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ConnectionID != "conn-1" || payload.Username != "ada" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestEncodeUsesWireKeys(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	envelope, err := event.New("user:user-1", event.TypeSessionEnded, event.SessionEnded{
		Reason: event.ReasonServerShutdown,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := envelope.Encode()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"v":1`, `"id":"evt_`, `"type":"session.ended"`, `"topic":"user:user-1"`, `"ts":"2026-11-14T12:00:00Z"`, `"reason":"server_shutdown"`} {
		if !strings.Contains(string(encoded), key) {
			t.Fatalf("encoded envelope missing %s: %s", key, encoded)
		}
	}
}

func TestNewRejectsUnencodablePayload(t *testing.T) {
	_, err := event.New("user:user-1", "test", func() {}, time.Now())
	if err == nil {
		t.Fatal("expected an encode error")
	}
}
