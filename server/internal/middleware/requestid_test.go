package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

func TestRequestIDGeneratesAndEchoesID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var fromContext string

	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/thing", func(c *gin.Context) {
		fromContext = requestid.FromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if fromContext == "" {
		t.Fatal("handler did not see a generated request ID")
	}
	if got := recorder.Header().Get(requestid.Header); got != fromContext {
		t.Fatalf("response header = %q, want %q", got, fromContext)
	}
}

func TestRequestIDReusesIncomingID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var captured string

	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/thing", func(c *gin.Context) {
		captured = requestid.FromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/thing", nil)
	request.Header.Set(requestid.Header, "client-supplied-id")

	engine.ServeHTTP(httptest.NewRecorder(), request)

	if captured != "client-supplied-id" {
		t.Fatalf("request ID = %q, want %q", captured, "client-supplied-id")
	}
}

func TestRequestIDIsUniquePerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	seen := make(map[string]struct{})
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/thing", func(c *gin.Context) {
		seen[requestid.FromContext(c.Request.Context())] = struct{}{}
		c.Status(http.StatusNoContent)
	})

	for range 10 {
		engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/thing", nil))
	}

	if len(seen) != 10 {
		t.Fatalf("generated %d distinct request IDs, want 10", len(seen))
	}
}
