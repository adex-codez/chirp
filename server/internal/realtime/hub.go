// Package realtime owns live WebSocket connections and fanout.
// Authentication happens before Open is reachable, so a Session always
// carries a resolved User and access-token expiry.
package realtime

import (
	"errors"
	"sync"
	"time"

	"backend/internal/event"
	"github.com/gorilla/websocket"
)

const (
	DefaultHeartbeat = 25 * time.Second
	defaultRead      = 60 * time.Second
	defaultWrite     = 10 * time.Second
	defaultQueue     = 16
	defaultMaxFrame  = 4096
)

// Application close codes. 4000-4999 is the private-use range.
const (
	CloseTokenExpired = 4001
	CloseSessionEnd   = 4003
	CloseSlowConsumer = 4009
)

var (
	ErrIdentity = errors.New("realtime identity is incomplete")
	ErrClosed   = errors.New("realtime hub is closed")
)

type Config struct {
	Heartbeat    time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	QueueDepth   int
	MaxFrame     int64
}

func (c Config) withDefaults() Config {
	if c.Heartbeat <= 0 {
		c.Heartbeat = DefaultHeartbeat
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = defaultRead
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = defaultWrite
	}
	if c.QueueDepth <= 0 {
		c.QueueDepth = defaultQueue
	}
	if c.MaxFrame <= 0 {
		c.MaxFrame = defaultMaxFrame
	}
	return c
}

type Identity struct {
	UserID    string
	ExpiresAt time.Time
}

// Delivery fans one block of events out to every live session of one User.
type Delivery struct {
	UserID string
	Events []event.Event
}

type Hub struct {
	mu       sync.RWMutex
	sessions map[string]map[*Session]struct{}
	cfg      Config
	closed   bool
	shutdown sync.Once
}

func NewHub(cfg Config) *Hub {
	return &Hub{
		sessions: make(map[string]map[*Session]struct{}),
		cfg:      cfg.withDefaults(),
	}
}

// Open registers a live connection. It returns once both pumps are running;
// the handler can then return, because the connection is hijacked.
func (h *Hub) Open(conn *websocket.Conn, id Identity) (*Session, error) {
	if id.UserID == "" || id.ExpiresAt.IsZero() {
		_ = conn.Close()
		return nil, ErrIdentity
	}

	sessionID, err := newSessionID()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	session := &Session{
		id:        sessionID,
		userID:    id.UserID,
		expiresAt: id.ExpiresAt,
		conn:      conn,
		cfg:       h.cfg,
		send:      make(chan []event.Event, h.cfg.QueueDepth),
		done:      make(chan struct{}),
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		_ = conn.Close()
		return nil, ErrClosed
	}
	if h.sessions[id.UserID] == nil {
		h.sessions[id.UserID] = make(map[*Session]struct{})
	}
	h.sessions[id.UserID][session] = struct{}{}
	h.mu.Unlock()

	go session.writeLoop()
	go session.expiryLoop()
	go session.readLoop(h)
	return session, nil
}

// Publish returns how many live sessions accepted the block.
func (h *Hub) Publish(d Delivery) int {
	if len(d.Events) == 0 {
		return 0
	}
	reached := 0
	for _, session := range h.snapshot(d.UserID) {
		if session.enqueue(d.Events) {
			reached++
		}
	}
	return reached
}

// CloseUser ends every live session of one User.
func (h *Hub) CloseUser(userID string, reason event.EndReason) {
	for _, session := range h.snapshot(userID) {
		session.end(event.SessionEnded{Reason: reason}, CloseSessionEnd, "session ended")
	}
}

// snapshot copies a User's live sessions so the caller can write without the lock.
func (h *Hub) snapshot(userID string) []*Session {
	h.mu.RLock()
	defer h.mu.RUnlock()
	sessions := make([]*Session, 0, len(h.sessions[userID]))
	for session := range h.sessions[userID] {
		sessions = append(sessions, session)
	}
	return sessions
}

// Shutdown ends every live connection. It is idempotent and safe to call
// from the signal path while the HTTP server drains.
func (h *Hub) Shutdown() {
	h.shutdown.Do(func() {
		h.mu.Lock()
		h.closed = true
		sessions := make([]*Session, 0, len(h.sessions))
		for _, set := range h.sessions {
			for session := range set {
				sessions = append(sessions, session)
			}
		}
		h.sessions = make(map[string]map[*Session]struct{})
		h.mu.Unlock()

		var wg sync.WaitGroup
		for _, session := range sessions {
			wg.Add(1)
			go func(session *Session) {
				defer wg.Done()
				session.end(event.SessionEnded{Reason: event.ReasonServerShutdown}, websocket.CloseGoingAway, "server shutdown")
			}(session)
		}
		wg.Wait()
	})
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, set := range h.sessions {
		count += len(set)
	}
	return count
}

func (h *Hub) remove(session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.sessions[session.userID]
	delete(set, session)
	if len(set) == 0 {
		delete(h.sessions, session.userID)
	}
}
