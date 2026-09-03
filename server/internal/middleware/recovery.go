package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"backend/internal/response"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.ErrorContext(c.Request.Context(), "panic recovered",
					"error", recovered,
					"stack", string(debug.Stack()),
				)
				c.Abort()
				response.Error(c, http.StatusInternalServerError, "internal server error")
			}
		}()

		c.Next()
	}
}
