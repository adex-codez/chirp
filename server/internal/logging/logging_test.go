package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewStampsStaticAttributes(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{Output: &buf})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("hello")

	line := buf.String()
	if !strings.Contains(line, `"service":"chirp-server"`) {
		t.Fatalf("log line missing service attribute: %s", line)
	}
	if !strings.Contains(line, `"version":`) {
		t.Fatalf("log line missing version attribute: %s", line)
	}
	if !strings.Contains(line, `"env":`) {
		t.Fatalf("log line missing env attribute: %s", line)
	}
}

func TestNewDefaultsToInfoLevel(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{Output: &buf})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Debug("suppressed")
	if buf.Len() != 0 {
		t.Fatalf("debug line emitted at the default level: %s", buf.String())
	}

	logger.Info("emitted")
	if !strings.Contains(buf.String(), "emitted") {
		t.Fatalf("info line missing: %s", buf.String())
	}
}

func TestNewHonoursConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{Level: "warn", Output: &buf})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("suppressed")
	if buf.Len() != 0 {
		t.Fatalf("info line emitted at warn level: %s", buf.String())
	}

	logger.Warn("emitted")
	if !strings.Contains(buf.String(), "emitted") {
		t.Fatalf("warn line missing: %s", buf.String())
	}
}

func TestNewWritesTextFormat(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("hello")

	if !strings.Contains(buf.String(), "level=INFO") {
		t.Fatalf("expected text encoded line, got %s", buf.String())
	}
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	if _, err := New(Config{Level: "verbose"}); err == nil {
		t.Fatal("New() error = nil, want error for an unknown level")
	}
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	if _, err := New(Config{Format: "xml"}); err == nil {
		t.Fatal("New() error = nil, want error for an unknown format")
	}
}
