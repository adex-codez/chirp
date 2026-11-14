package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/logging"
	"backend/internal/realtime"
	"backend/internal/router"
	"backend/internal/server"
)

func main() {
	// Load .env first so LOG_LEVEL and LOG_FORMAT come from the same file as
	// the rest of the configuration.
	_ = config.LoadDotEnv()

	logger, err := logging.New(logging.ConfigFromEnv())
	if err != nil {
		// Report through a bootstrap logger: the configured one is unusable.
		slog.New(slog.NewJSONHandler(os.Stdout, nil)).Error("invalid log configuration", "error", err)
		os.Exit(1)
	}
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if cfg.Server.StartupTimeout <= 0 {
		cfg.Server.StartupTimeout = config.DefaultStartupTimeout
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		cfg.Server.ShutdownTimeout = config.DefaultShutdownTimeout
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.StartupTimeout)
	defer cancel()
	db, err := database.NewPostgres(startupCtx, cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	hub := realtime.NewHub(realtime.Config{})
	routes, err := router.New(db, cfg.Server, cfg.Auth, cfg.Mail, cfg.R2, cfg.Realtime, hub)
	if err != nil {
		slog.Error("failed to create router", "error", err)
		os.Exit(1)
	}
	httpServer := server.New(cfg.Server.Address(), routes, cfg.Server.ShutdownTimeout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Hijacked WebSocket connections are outside http.Server.Shutdown, so the
	// hub closes them when the process starts draining.
	go func() {
		<-ctx.Done()
		hub.Shutdown()
	}()

	slog.Info("starting server", "addr", cfg.Server.Address(), "mode", cfg.Server.Mode)
	runErr := httpServer.Run(ctx)
	hub.Shutdown()
	if runErr != nil {
		slog.Error("server stopped with error", "error", runErr)
		os.Exit(1)
	}
}
