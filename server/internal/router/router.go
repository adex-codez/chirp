package router

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, cfg config.ServerConfig, authCfg config.AuthConfig) (*gin.Engine, error) {
	gin.SetMode(cfg.Mode)

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	router.Use(middleware.Recovery(), middleware.Logger())

	healthService := service.NewHealthService(repository.NewHealthRepository(pool))
	health := handler.NewHealthHandler(healthService)
	router.GET("/health", health.Live)
	router.GET("/ready", health.Ready)

	authService := service.NewAuthService(repository.NewAuthRepository(pool), authCfg.JWTSecret, authCfg.DevExposeCodes)
	auth := handler.NewAuthHandler(authService)
	authGroup := router.Group("/auth")
	authGroup.POST("/signup", auth.SignUp)
	authGroup.POST("/verify", auth.Verify)
	authGroup.POST("/verify/resend", auth.Resend)
	authGroup.POST("/login", auth.SignIn)
	authGroup.GET("/me", auth.Profile)

	return router, nil
}
