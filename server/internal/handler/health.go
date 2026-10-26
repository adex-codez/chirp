package handler

import (
	"context"
	"log/slog"

	"backend/internal/requestid"
	"backend/internal/response"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type healthService interface {
	Live() service.HealthStatus
	Ready(context.Context) (service.HealthStatus, error)
}

type HealthHandler struct {
	service healthService
}

func NewHealthHandler(svc healthService) *HealthHandler {
	return &HealthHandler{service: svc}
}

func (h *HealthHandler) Live(c *gin.Context) {
	response.Success(c, h.service.Live())
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.service.Ready(ctx)
	if err != nil {
		// Log the cause before responding: the client only ever sees a generic
		// message, so without this line a database outage is indistinguishable
		// from a timeout or a rejected connection.
		slog.ErrorContext(ctx, "readiness check failed",
			"request_id", requestid.FromContext(ctx),
			"error", err,
		)

		response.FailCode(c, response.CodeServiceUnavailable)
		return
	}

	response.Success(c, status)
}
