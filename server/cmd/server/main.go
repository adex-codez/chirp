package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/router"
	"backend/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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

	routes, err := router.New(db, cfg.Server, cfg.Auth, cfg.Mail)
	if err != nil {
		slog.Error("failed to create router", "error", err)
		os.Exit(1)
	}
	httpServer := server.New(cfg.Server.Address(), routes, cfg.Server.ShutdownTimeout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting server", "addr", cfg.Server.Address(), "mode", cfg.Server.Mode)
	if err := httpServer.Run(ctx); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
