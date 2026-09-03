package service

import (
	"context"

	"backend/internal/repository"
)

type HealthStatus struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
}

// HealthService contains the application behavior for liveness and readiness.
type HealthService struct {
	repository repository.HealthRepository
}

func NewHealthService(repo repository.HealthRepository) *HealthService {
	return &HealthService{repository: repo}
}

func (s *HealthService) Live() HealthStatus {
	return HealthStatus{Status: "ok"}
}

func (s *HealthService) Ready(ctx context.Context) (HealthStatus, error) {
	if err := s.repository.Ping(ctx); err != nil {
		return HealthStatus{}, err
	}

	return HealthStatus{Status: "ready", Database: "ok"}, nil
}
