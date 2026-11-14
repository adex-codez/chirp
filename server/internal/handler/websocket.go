package handler

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"backend/internal/event"
	"backend/internal/realtime"
	"backend/internal/response"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type realtimeSessions interface {
	RealtimeSession(ctx context.Context, accessToken string) (service.RealtimeSession, error)
}

type WebSocketHandler struct {
	sessions realtimeSessions
	hub      *realtime.Hub
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(sessions realtimeSessions, hub *realtime.Hub, allowedOrigins []string) *WebSocketHandler {
	return &WebSocketHandler{
		sessions: sessions,
		hub:      hub,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			Subprotocols:     []string{"chirp.v1"},
			CheckOrigin: func(r *http.Request) bool {
				return originAllowed(r, allowedOrigins)
			},
		},
	}
}

// Connect authenticates the access token before the upgrade, then hands the
// socket to the hub. Returning is safe: gorilla hijacks the connection.
func (h *WebSocketHandler) Connect(c *gin.Context) {
	token, ok := websocketToken(c.Request)
	if !ok {
		response.FailCode(c, response.CodeInvalidSession)
		return
	}
	session, err := h.sessions.RealtimeSession(c.Request.Context(), token)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "websocket handshake rejected", "error", err)
		response.Fail(c, err)
		return
	}
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	live, err := h.hub.Open(conn, realtime.Identity{
		UserID:    session.User.ID,
		ExpiresAt: session.ExpiresAt,
	})
	if err != nil {
		return
	}
	slog.InfoContext(c.Request.Context(), "websocket connected",
		"connection_id", live.ID(),
		"user_id", session.User.ID,
	)
	ready, err := event.New(event.UserTopic(session.User.ID), event.TypeConnectionReady, event.ConnectionReady{
		ConnectionID:     live.ID(),
		Username:         session.User.Username,
		HeartbeatSeconds: int(realtime.DefaultHeartbeat / time.Second),
		AccessExpiresAt:  session.ExpiresAt,
	}, time.Now())
	if err != nil {
		live.Close(websocket.CloseInternalServerErr, "encode failed")
		return
	}
	live.Send(ready)
}

// websocketToken accepts the handshake Authorization header, which React
// Native supports, and a bearer subprotocol entry for browsers, which cannot
// set handshake headers. Neither path puts the token in the URL.
func websocketToken(r *http.Request) (string, bool) {
	if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
		return token, true
	}
	for _, offered := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		token, ok := strings.CutPrefix(strings.TrimSpace(offered), "bearer.")
		if ok && token != "" {
			return token, true
		}
	}
	return "", false
}

func originAllowed(r *http.Request, allowed []string) bool {
	// A native handshake proves itself with the Authorization header, which a
	// browser cannot set cross-site. Native runtimes may emit an Origin the
	// server has no business policing, so only the browser path is filtered.
	if _, ok := bearerToken(r.Header.Get("Authorization")); ok {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if slices.Contains(allowed, origin) {
		return true
	}
	parsed, err := url.Parse(origin)
	return err == nil && parsed.Host != "" && parsed.Host == r.Host
}
