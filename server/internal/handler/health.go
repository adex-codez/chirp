package handler

import (
	"context"
	"net/http"

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
	status, err := h.service.Ready(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "service unavailable")
		return
	}

	response.Success(c, status)
}
