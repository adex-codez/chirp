// Package event defines the versioned envelope for server-pushed events.
// It knows nothing about sockets, HTTP, or Gin, so a future webhook
// dispatcher can publish the same payloads.
package event

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Version is the envelope schema version. Bump it only for a breaking change.
const Version = 1

const (
	TypeConnectionReady = "connection.ready"
	TypeSessionEnded    = "session.ended"
)

type EndReason string

const (
	ReasonExpired        EndReason = "expired"
	ReasonRevoked        EndReason = "revoked"
	ReasonSignedOut      EndReason = "signed_out"
	ReasonServerShutdown EndReason = "server_shutdown"
)

// Event is one server-to-client frame.
type Event struct {
	Version int             `json:"v"`
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Topic   string          `json:"topic"`
	Time    time.Time       `json:"ts"`
	Data    json.RawMessage `json:"data"`
}

// ConnectionReady is sent once per accepted socket.
type ConnectionReady struct {
	ConnectionID     string    `json:"connection_id"`
	Username         string    `json:"username"`
	HeartbeatSeconds int       `json:"heartbeat_seconds"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
}

// SessionEnded precedes a server-initiated close.
type SessionEnded struct {
	Reason EndReason `json:"reason"`
}

// New builds an envelope with a fresh stable ID. Payload types are concrete,
// so a caller cannot publish an event whose data shape is undeclared.
func New[T any](topic, eventType string, payload T, now time.Time) (Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("encode %s payload: %w", eventType, err)
	}
	id, err := newID()
	if err != nil {
		return Event{}, err
	}
	return Event{
		Version: Version,
		ID:      id,
		Type:    eventType,
		Topic:   topic,
		Time:    now.UTC(),
		Data:    data,
	}, nil
}

func (e Event) Encode() ([]byte, error) {
	return json.Marshal(e)
}

// UserTopic addresses one User's connections. Conversation topics will sit
// beside this when rooms exist.
func UserTopic(userID string) string {
	return "user:" + userID
}

func newID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate event id: %w", err)
	}
	return "evt_" + hex.EncodeToString(raw[:]), nil
}
