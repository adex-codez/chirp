package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/internal/event"
	"backend/internal/realtime"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type fakeRealtimeSessions struct {
	session service.RealtimeSession
	err     error
}

func (f fakeRealtimeSessions) RealtimeSession(context.Context, string) (service.RealtimeSession, error) {
	return f.session, f.err
}

func TestWebSocketToken(t *testing.T) {
	cases := []struct {
		name          string
		authorization string
		protocols     string
		want          string
		ok            bool
	}{
		{"authorization header", "Bearer header-token", "", "header-token", true},
		{"bearer subprotocol", "", "chirp.v1, bearer.sub-token", "sub-token", true},
		{"header wins over subprotocol", "Bearer header-token", "chirp.v1, bearer.sub-token", "header-token", true},
		{"empty bearer entry", "", "chirp.v1, bearer.", "", false},
		{"no token", "", "chirp.v1", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tc.authorization != "" {
				request.Header.Set("Authorization", tc.authorization)
			}
			if tc.protocols != "" {
				request.Header.Set("Sec-WebSocket-Protocol", tc.protocols)
			}

			got, ok := websocketToken(request)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("websocketToken() = %q, %v; want %q, %v", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestOriginAllowed(t *testing.T) {
	cases := []struct {
		name          string
		origin        string
		authorization string
		allowed       []string
		want          bool
	}{
		{"no origin", "", "", nil, true},
		{"configured origin", "http://localhost:8081", "", []string{"http://localhost:8081"}, true},
		{"same host origin", "https://api.chirp.test", "", nil, true},
		{"foreign origin", "https://evil.test", "", []string{"http://localhost:8081"}, false},
		{"native handshake skips origin filter", "https://evil.test", "Bearer token", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "https://api.chirp.test/ws", nil)
			if tc.origin != "" {
				request.Header.Set("Origin", tc.origin)
			}
			if tc.authorization != "" {
				request.Header.Set("Authorization", tc.authorization)
			}

			if got := originAllowed(request, tc.allowed); got != tc.want {
				t.Fatalf("originAllowed() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConnectSendsReadyEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := realtime.NewHub(realtime.Config{})
	t.Cleanup(hub.Shutdown)

	session := service.RealtimeSession{
		User:      service.PublicUser{ID: "user-1", Username: "ada"},
		ExpiresAt: time.Now().Add(time.Hour),
	}
	handler := NewWebSocketHandler(fakeRealtimeSessions{session: session}, hub, nil)
	router := gin.New()
	router.GET("/ws", handler.Connect)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	socketURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, http.Header{
		"Authorization": []string{"Bearer test-token"},
	})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, frame, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var envelope event.Event
	if err := json.Unmarshal(frame, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Type != event.TypeConnectionReady {
		t.Fatalf("type = %q, want %q", envelope.Type, event.TypeConnectionReady)
	}
	if envelope.Topic != event.UserTopic("user-1") {
		t.Fatalf("topic = %q", envelope.Topic)
	}

	var ready event.ConnectionReady
	if err := json.Unmarshal(envelope.Data, &ready); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if ready.ConnectionID == "" {
		t.Fatal("ready event has no connection id")
	}
	if ready.Username != "ada" {
		t.Fatalf("username = %q", ready.Username)
	}
}

func TestConnectRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := realtime.NewHub(realtime.Config{})
	t.Cleanup(hub.Shutdown)

	handler := NewWebSocketHandler(fakeRealtimeSessions{}, hub, nil)
	router := gin.New()
	router.GET("/ws", handler.Connect)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ws", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
