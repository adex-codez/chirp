package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &buf
}

func logRequest(t *testing.T, path string, status int) *bytes.Buffer {
	t.Helper()

	logs := captureLogs(t)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID(), Recovery(), Logger())
	engine.GET(path, func(c *gin.Context) {
		c.Status(status)
	})

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))

	return logs
}

func decodeLine(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()

	line := strings.TrimSpace(logs.String())
	if line == "" {
		t.Fatal("no log line emitted")
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("decoding log line %q: %v", line, err)
	}

	return record
}

func TestLoggerRoutesStatusToLevel(t *testing.T) {
	tests := []struct {
		name   string
		status int
		level  string
	}{
		{name: "success logs at info", status: http.StatusOK, level: "INFO"},
		{name: "client error logs at warn", status: http.StatusBadRequest, level: "WARN"},
		{name: "server error logs at error", status: http.StatusInternalServerError, level: "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := decodeLine(t, logRequest(t, "/thing", tt.status))

			if record["level"] != tt.level {
				t.Fatalf("level = %v, want %v", record["level"], tt.level)
			}
		})
	}
}

func TestLoggerEmitsDebuggableFields(t *testing.T) {
	record := decodeLine(t, logRequest(t, "/thing", http.StatusInternalServerError))

	for _, field := range []string{"request_id", "method", "path", "status", "duration_ms", "client_ip"} {
		if _, ok := record[field]; !ok {
			t.Fatalf("log line missing %q field: %v", field, record)
		}
	}

	if record["path"] != "/thing" {
		t.Fatalf("path = %v, want /thing", record["path"])
	}
	if record["method"] != http.MethodGet {
		t.Fatalf("method = %v, want %s", record["method"], http.MethodGet)
	}
	if requestID, _ := record["request_id"].(string); requestID == "" {
		t.Fatal("request_id is empty, the log line cannot be correlated")
	}
}

func TestLoggerSkipsSuccessfulProbes(t *testing.T) {
	logs := logRequest(t, "/health", http.StatusOK)

	if line := strings.TrimSpace(logs.String()); line != "" {
		t.Fatalf("expected no log for a healthy probe, got %q", line)
	}
}

func TestLoggerKeepsFailingProbes(t *testing.T) {
	record := decodeLine(t, logRequest(t, "/ready", http.StatusServiceUnavailable))

	if record["level"] != "ERROR" {
		t.Fatalf("level = %v, want ERROR", record["level"])
	}
}

func TestRecoveryLogsRequestContextAndStack(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logs := captureLogs(t)

	engine := gin.New()
	engine.Use(RequestID(), Recovery(), Logger())
	engine.GET("/boom", func(*gin.Context) {
		panic("something broke")
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	record := decodeLine(t, logs)
	if record["msg"] != "panic recovered" {
		t.Fatalf("msg = %v, want panic recovered", record["msg"])
	}

	// A panic unwinds the chain, so the recovery line is the only one emitted.
	if record["path"] != "/boom" {
		t.Fatalf("path = %v, want /boom", record["path"])
	}
	if _, ok := record["stack"].(string); !ok {
		t.Fatalf("panic log line missing stack: %v", record)
	}
}
