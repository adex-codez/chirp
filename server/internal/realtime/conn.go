package realtime

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"backend/internal/event"
	"github.com/gorilla/websocket"
)

// Session is one live socket. Only writeLoop writes data frames; close frames
// and the final session.ended frame take writeMu so they cannot interleave.
type Session struct {
	id        string
	userID    string
	expiresAt time.Time
	conn      *websocket.Conn
	cfg       Config
	send      chan []event.Event
	done      chan struct{}
	once      sync.Once
	writeMu   sync.Mutex
}

func newSessionID() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return "conn_" + hex.EncodeToString(raw[:]), nil
}

func (s *Session) ID() string {
	return s.id
}

// Send queues events on this connection only.
func (s *Session) Send(events ...event.Event) bool {
	return s.enqueue(events)
}

func (s *Session) enqueue(events []event.Event) bool {
	select {
	case <-s.done:
		return false
	default:
	}
	select {
	case s.send <- events:
		return true
	case <-s.done:
		return false
	default:
		s.Close(CloseSlowConsumer, "outbound queue full")
		return false
	}
}

// Close sends a close frame, stops both pumps, and closes the transport.
func (s *Session) Close(code int, reason string) {
	s.once.Do(func() {
		s.writeMu.Lock()
		_ = s.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason),
			time.Now().Add(s.cfg.WriteTimeout),
		)
		s.writeMu.Unlock()
		close(s.done)
		_ = s.conn.Close()
	})
}

// end delivers a final event synchronously, then closes. A client sees the
// reason before the close frame, so it can distinguish expiry from shutdown.
func (s *Session) end(payload event.SessionEnded, code int, reason string) {
	s.once.Do(func() {
		s.writeMu.Lock()
		final, err := event.New(event.UserTopic(s.userID), event.TypeSessionEnded, payload, time.Now())
		if err == nil {
			if encoded, encodeErr := final.Encode(); encodeErr == nil {
				_ = s.conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
				_ = s.conn.WriteMessage(websocket.TextMessage, encoded)
			}
		}
		_ = s.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason),
			time.Now().Add(s.cfg.WriteTimeout),
		)
		s.writeMu.Unlock()
		close(s.done)
		_ = s.conn.Close()
	})
}

func (s *Session) readLoop(h *Hub) {
	defer func() {
		h.remove(s)
		slog.Info("websocket closed", "connection_id", s.id, "user_id", s.userID)
	}()
	s.conn.SetReadLimit(s.cfg.MaxFrame)
	_ = s.conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))
	s.conn.SetPongHandler(func(string) error {
		return s.conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))
	})

	for {
		// v1 defines no client commands. Inbound frames are read only to keep
		// the pong handler and close handshake alive.
		if _, _, err := s.conn.ReadMessage(); err != nil {
			s.Close(websocket.CloseNormalClosure, "")
			return
		}
	}
}

func (s *Session) writeLoop() {
	ticker := time.NewTicker(s.cfg.Heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case block := <-s.send:
			select {
			case <-s.done:
				return
			default:
			}
			if !s.writeBlock(block) {
				return
			}
		case <-ticker.C:
			s.writeMu.Lock()
			_ = s.conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
			err := s.conn.WriteMessage(websocket.PingMessage, nil)
			s.writeMu.Unlock()
			if err != nil {
				s.Close(websocket.CloseInternalServerErr, "heartbeat failed")
				return
			}
		}
	}
}

func (s *Session) writeBlock(block []event.Event) bool {
	for _, envelope := range block {
		encoded, err := envelope.Encode()
		if err != nil {
			s.Close(websocket.CloseInternalServerErr, "encode failed")
			return false
		}
		s.writeMu.Lock()
		_ = s.conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
		err = s.conn.WriteMessage(websocket.TextMessage, encoded)
		s.writeMu.Unlock()
		if err != nil {
			s.Close(websocket.CloseInternalServerErr, "write failed")
			return false
		}
	}
	return true
}

func (s *Session) expiryLoop() {
	timer := time.NewTimer(time.Until(s.expiresAt))
	defer timer.Stop()

	select {
	case <-s.done:
	case <-timer.C:
		s.end(event.SessionEnded{Reason: event.ReasonExpired}, CloseTokenExpired, "access token expired")
	}
}
