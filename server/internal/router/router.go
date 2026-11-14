package router

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/mail"
	"backend/internal/middleware"
	"backend/internal/profilepicture"
	"backend/internal/realtime"
	"backend/internal/repository"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, cfg config.ServerConfig, authCfg config.AuthConfig, mailCfg config.MailConfig, r2Cfg *config.R2Config, realtimeCfg config.RealtimeConfig, hub *realtime.Hub) (*gin.Engine, error) {
	gin.SetMode(cfg.Mode)

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	// RequestID runs outermost so the recovery and request loggers can read the
	// correlation ID assigned to the request they are handling.
	router.Use(middleware.RequestID(), middleware.Recovery(), middleware.Logger())

	healthService := service.NewHealthService(repository.NewHealthRepository(pool))
	health := handler.NewHealthHandler(healthService)
	router.GET("/health", health.Live)
	router.GET("/ready", health.Ready)

	authRepo := repository.NewAuthRepository(pool)
	var pictures *profilepicture.Client
	if r2Cfg != nil {
		pictures = profilepicture.NewClient(*r2Cfg)
	}
	authApp := service.NewAuthService(
		authRepo,
		authRepo,
		authRepo,
		authRepo,
		authCfg.JWTSecret,
		authCfg.DevExposeCodes,
		authCfg.AppleAudience,
		authCfg.GoogleAudiences,
		mail.NewClient(mailCfg.SendlibAPIKey, mailCfg.SendlibBaseURL, mailCfg.SendlibFrom),
		pictures,
	)
	auth := handler.NewAuthHandler(authApp)
	authGroup := router.Group("/auth")
	authGroup.POST("/signup", auth.SignUp)
	authGroup.POST("/verify", auth.Verify)
	authGroup.POST("/verify/resend", auth.Resend)
	authGroup.POST("/login", auth.SignIn)
	authGroup.POST("/social", auth.SocialSignIn)
	authGroup.POST("/username", auth.SetUsername)
	authGroup.POST("/username/change", auth.ChangeUsername)
	authGroup.POST("/profile-picture", auth.UploadProfilePicture)
	authGroup.POST("/password/forgot", auth.ForgotPassword)
	authGroup.POST("/password/reset", auth.ResetPassword)
	authGroup.POST("/password", auth.AddPassword)
	authGroup.POST("/refresh", auth.Refresh)
	authGroup.POST("/logout", auth.SignOut)
	authGroup.POST("/logout-all", auth.SignOutAll)
	authGroup.GET("/me", auth.Profile)

	contacts := handler.NewContactHandler(
		authApp,
		service.NewContactsService(repository.NewContactRepository(pool)),
	)
	contactGroup := router.Group("/contacts")
	contactGroup.GET("", contacts.List)
	contactGroup.POST("", contacts.Add)

	websocket := handler.NewWebSocketHandler(authApp, hub, realtimeCfg.AllowedOrigins)
	router.GET("/ws", websocket.Connect)

	return router, nil
}
