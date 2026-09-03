package service

import (
	"context"
	"errors"
	"testing"
)

type fakeHealthRepository struct {
	err error
}

func (r fakeHealthRepository) Ping(context.Context) error {
	return r.err
}

func TestHealthServiceLive(t *testing.T) {
	svc := NewHealthService(fakeHealthRepository{})

	status := svc.Live()

	if status.Status != "ok" {
		t.Fatalf("status = %q, want %q", status.Status, "ok")
	}
	if status.Database != "" {
		t.Fatalf("database = %q, want empty", status.Database)
	}
}

func TestHealthServiceReady(t *testing.T) {
	svc := NewHealthService(fakeHealthRepository{})

	status, err := svc.Ready(context.Background())

	if err != nil {
		t.Fatalf("Ready() error = %v", err)
	}
	if status.Status != "ready" {
		t.Fatalf("status = %q, want %q", status.Status, "ready")
	}
	if status.Database != "ok" {
		t.Fatalf("database = %q, want %q", status.Database, "ok")
	}
}

func TestHealthServiceReadyReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	svc := NewHealthService(fakeHealthRepository{err: wantErr})

	_, err := svc.Ready(context.Background())

	if !errors.Is(err, wantErr) {
		t.Fatalf("Ready() error = %v, want %v", err, wantErr)
	}
}
