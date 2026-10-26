package response

import (
	"log/slog"
	"net/http"

	"backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Response{Success: false, Error: message})
}

// Fail writes a 5xx response whose cause must not reach the client. The cause
// is logged once with the request ID so operators can trace the failure.
func Fail(c *gin.Context, status int, message string, cause error) {
	slog.ErrorContext(c.Request.Context(), "request failed",
		"request_id", requestid.FromContext(c.Request.Context()),
		"status", status,
		"error", cause,
	)
	Error(c, status, message)
}
