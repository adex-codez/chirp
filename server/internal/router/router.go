package router

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

func New(db repository.HealthRepository, cfg config.ServerConfig) (*gin.Engine, error) {
	gin.SetMode(cfg.Mode)

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	router.Use(middleware.Recovery(), middleware.Logger())

	healthService := service.NewHealthService(db)
	h := handler.NewHealthHandler(healthService)
	router.GET("/health", h.Live)
	router.GET("/ready", h.Ready)

	return router, nil
}
