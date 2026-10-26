package middleware

import (
	"log/slog"
	"runtime/debug"

	"backend/internal/response"
	"github.com/gin-gonic/gin"
)

// Recovery turns a panic into a 500 instead of a dropped connection. The log
// line carries the same request fields as the request logger, because a panic
// unwinds the handler chain and the "request completed" line is never reached.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				attrs := append(requestLogAttrs(c),
					"error", recovered,
					"stack", string(debug.Stack()),
				)
				slog.ErrorContext(c.Request.Context(), "panic recovered", attrs...)

				c.Abort()
				response.FailCode(c, response.CodeInternalError)
			}
		}()

		c.Next()
	}
}
