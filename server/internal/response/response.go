package response

import (
	"log/slog"
	"net/http"

	"backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool              `json:"success"`
	Data    any               `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
	Code    string            `json:"code,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

// Fail writes the failure that err maps to. Every 5xx is logged once with its
// cause, since the client only reads a generic message.
func Fail(c *gin.Context, err error) {
	problem := resolve(From(err))
	if problem.Status >= http.StatusInternalServerError {
		slog.ErrorContext(c.Request.Context(), "request failed",
			"request_id", requestid.FromContext(c.Request.Context()),
			"status", problem.Status,
			"code", problem.Code,
			"error", err,
		)
	}
	write(c, problem)
}

// FailCode writes the failure for a code with no error behind it, and leaves
// logging to the caller.
func FailCode(c *gin.Context, code string) {
	write(c, resolve(Problem{Code: code}))
}

// Validation reports a 400 with a message per offending field.
func Validation(c *gin.Context, fields map[string]string) {
	write(c, resolve(Problem{Code: CodeValidationFailed, Fields: fields}))
}

func write(c *gin.Context, problem Problem) {
	c.JSON(problem.Status, Response{
		Success: false,
		Error:   problem.Message,
		Code:    problem.Code,
		Fields:  problem.Fields,
	})
}
