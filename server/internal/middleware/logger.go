package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

// probePaths are polled by orchestrators several times a minute. Their
// successes are noise, but their failures are not: a failing /ready means the
// instance should be pulled out of rotation, so only 2xx and 3xx probes are
// dropped.
var probePaths = map[string]struct{}{
	"/health": {},
	"/ready":  {},
}

// Logger emits one line per request that is not a successful probe. The
// severity is derived from the response status so that alerting on level
// catches 4xx and 5xx responses, which would otherwise be indistinguishable
// from successful traffic.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		status := c.Writer.Status()
		if _, isProbe := probePaths[c.Request.URL.Path]; isProbe && status < http.StatusBadRequest {
			return
		}

		attrs := append(requestLogAttrs(c),
			"status", status,
			"duration_ms", float64(time.Since(startedAt).Microseconds())/1000,
		)

		switch {
		case status >= http.StatusInternalServerError:
			slog.ErrorContext(c.Request.Context(), "request completed", attrs...)
		case status >= http.StatusBadRequest:
			slog.WarnContext(c.Request.Context(), "request completed", attrs...)
		default:
			slog.InfoContext(c.Request.Context(), "request completed", attrs...)
		}
	}
}

func requestLogAttrs(c *gin.Context) []any {
	return []any{
		"request_id", requestid.FromContext(c.Request.Context()),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"client_ip", c.ClientIP(),
	}
}
