package realtime_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/internal/event"
	"backend/internal/realtime"
	"github.com/gorilla/websocket"
)

func TestPublishesReadyAndEvents(t *testing.T) {
	hub := realtime.NewHub(realtime.Config{})
	conn := dial(t, hub, time.Now().Add(time.Hour))

	ready := readEvent(t, conn)
	if ready.Type != event.TypeConnectionReady {
		t.Fatalf("first type = %q", ready.Type)
	}
	var connection event.ConnectionReady
	if err := json.Unmarshal(ready.Data, &connection); err != nil {
		t.Fatal(err)
	}
	if connection.ConnectionID == "" {
		t.Fatal("ready event has no connection id")
	}

	sent, err := event.New("user:user-1", "test.ping", struct {
		Value string `json:"value"`
	}{Value: "ok"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if reached := hub.Publish(realtime.Delivery{UserID: "user-1", Events: []event.Event{sent}}); reached != 1 {
		t.Fatalf("reached = %d, want 1", reached)
	}

	got := readEvent(t, conn)
	if got.ID != sent.ID || got.Type != "test.ping" {
		t.Fatalf("event = %+v", got)
	}
}

func TestFansOutToEveryUserSession(t *testing.T) {
	hub := realtime.NewHub(realtime.Config{})
	first := dial(t, hub, time.Now().Add(time.Hour))
	second := dial(t, hub, time.Now().Add(time.Hour))
	readEvent(t, first)
	readEvent(t, second)

	sent, err := event.New("user:user-1", "test.ping", struct{}{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if reached := hub.Publish(realtime.Delivery{UserID: "user-1", Events: []event.Event{sent}}); reached != 2 {
		t.Fatalf("reached = %d, want 2", reached)
	}
	if got := readEvent(t, first); got.ID != sent.ID {
		t.Fatalf("first id = %q", got.ID)
	}
	if got := readEvent(t, second); got.ID != sent.ID {
		t.Fatalf("second id = %q", got.ID)
	}
}

func TestConnectionClosesWhenAccessTokenExpires(t *testing.T) {
	hub := realtime.NewHub(realtime.Config{})
	conn := dial(t, hub, time.Now().Add(200*time.Millisecond))
	readEvent(t, conn)

	ended := readEvent(t, conn)
	if ended.Type != event.TypeSessionEnded {
		t.Fatalf("type = %q", ended.Type)
	}
	var payload event.SessionEnded
	if err := json.Unmarshal(ended.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Reason != event.ReasonExpired {
		t.Fatalf("reason = %q", payload.Reason)
	}
	assertCloseCode(t, conn, realtime.CloseTokenExpired)
}

func TestShutdownEndsEveryConnection(t *testing.T) {
	hub := realtime.NewHub(realtime.Config{})
	conn := dial(t, hub, time.Now().Add(time.Hour))
	readEvent(t, conn)

	hub.Shutdown()
	if count := hub.Count(); count != 0 {
		t.Fatalf("count = %d after shutdown", count)
	}

	ended := readEvent(t, conn)
	var payload event.SessionEnded
	if err := json.Unmarshal(ended.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Reason != event.ReasonServerShutdown {
		t.Fatalf("reason = %q", payload.Reason)
	}
	assertCloseCode(t, conn, websocket.CloseGoingAway)
}

func dial(t *testing.T, hub *realtime.Hub, expiresAt time.Time) *websocket.Conn {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		session, err := hub.Open(raw, realtime.Identity{UserID: "user-1", ExpiresAt: expiresAt})
		if err != nil {
			t.Errorf("open: %v", err)
			return
		}
		ready, err := event.New(event.UserTopic("user-1"), event.TypeConnectionReady, event.ConnectionReady{
			ConnectionID: session.ID(),
		}, time.Now())
		if err != nil {
			t.Errorf("ready event: %v", err)
			return
		}
		session.Send(ready)
	}))
	t.Cleanup(server.Close)

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func readEvent(t *testing.T, conn *websocket.Conn) event.Event {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var envelope event.Event
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return envelope
}

func assertCloseCode(t *testing.T, conn *websocket.Conn, want int) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err := conn.ReadMessage()
	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) {
		t.Fatalf("read after close = %v, want CloseError", err)
	}
	if closeErr.Code != want {
		t.Fatalf("close code = %d, want %d", closeErr.Code, want)
	}
}
