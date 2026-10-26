package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
)

const (
	serviceName = "chirp-server"
)

// Config controls how the application emits structured log lines.
type Config struct {
	// Level is the minimum severity to emit: debug, info, warn or error.
	Level string
	// Format selects the encoder: json for shipping to a log collector,
	// text for reading over a terminal.
	Format string
	// Output receives the encoded lines. Defaults to stdout.
	Output io.Writer
}

// ConfigFromEnv reads the logging configuration from the process environment.
// Unset values are left empty so New can apply its defaults.
func ConfigFromEnv() Config {
	return Config{
		Level:  os.Getenv("LOG_LEVEL"),
		Format: os.Getenv("LOG_FORMAT"),
	}
}

// New builds a logger that stamps every line with the service name, build
// version and environment, so lines can be attributed after a rollout.
func New(cfg Config) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	options := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch format := strings.ToLower(strings.TrimSpace(cfg.Format)); format {
	case "", "json":
		handler = slog.NewJSONHandler(output, options)
	case "text":
		handler = slog.NewTextHandler(output, options)
	default:
		return nil, fmt.Errorf("invalid LOG_FORMAT %q, want json or text", cfg.Format)
	}

	return slog.New(handler).
		With("service", serviceName).
		With("version", Version()).
		With("env", Environment()), nil
}

// Version reports the build version recorded by the Go toolchain, falling back
// to "dev" for `go run` and other builds without module version information.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "dev"
	}

	return info.Main.Version
}

// Environment reports the deployment environment, defaulting to development.
func Environment() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}

	return "development"
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL %q, want debug, info, warn or error", level)
	}
}
