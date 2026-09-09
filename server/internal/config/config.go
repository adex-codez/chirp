package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Mail     MailConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	Mode            string
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	URL string
}

// MailConfig carries the Sendlib transactional mail sender settings.
// Empty APIKey disables sending (dev/test); codes stay in API responses.
type MailConfig struct {
	SendlibAPIKey  string
	SendlibBaseURL string
	SendlibFrom    string
}

// AuthConfig carries the signing secret for access tokens, the audiences
// accepted for Social sign-in tokens, and the dev code exposure flag.
// DevExposeCodes returns verification codes in API responses outside release
// builds, alongside the transactional mail send.
type AuthConfig struct {
	JWTSecret        string
	DevExposeCodes   bool
	AppleAudience    string
	GoogleAudiences  []string
}

const (
	DefaultStartupTimeout  = 10 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
)

func Load() (Config, error) {
	_ = godotenv.Load()

	port := 8080
	if value := os.Getenv("PORT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 65535 {
			return Config{}, fmt.Errorf("invalid PORT %q", value)
		}
		port = parsed
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = "debug"
	}
	if mode != "debug" && mode != "release" && mode != "test" {
		return Config{}, fmt.Errorf("invalid GIN_MODE %q", mode)
	}

	authCfg, err := loadAuthConfig(mode)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Server: ServerConfig{
			Host:            host,
			Port:            port,
			Mode:            mode,
			StartupTimeout:  DefaultStartupTimeout,
			ShutdownTimeout: DefaultShutdownTimeout,
		},
		Database: DatabaseConfig{URL: databaseURL},
		Auth:     authCfg,
		Mail:     loadMailConfig(),
	}, nil
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func splitList(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func loadMailConfig() MailConfig {
	baseURL := os.Getenv("SENDLIB_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sendlib.samueltuoyo.com/api/send"
	}
	return MailConfig{
		SendlibAPIKey:  strings.TrimSpace(os.Getenv("SENDLIB_API_KEY")),
		SendlibBaseURL: strings.TrimSpace(baseURL),
		SendlibFrom:    strings.TrimSpace(os.Getenv("SENDLIB_FROM")),
	}
}

func loadAuthConfig(mode string) (AuthConfig, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if mode == "release" {
			// Fail fast rather than sign tokens with a known secret.
			return AuthConfig{}, fmt.Errorf("JWT_SECRET is required in release mode")
		}
		secret = "dev-only-change-me"
	}
	return AuthConfig{
		JWTSecret:       secret,
		DevExposeCodes:  mode != "release",
		AppleAudience:   os.Getenv("APPLE_AUDIENCE"),
		GoogleAudiences: splitList(os.Getenv("GOOGLE_AUDIENCES")),
	}, nil
}
