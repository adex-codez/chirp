package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Mail     MailConfig
	R2       *R2Config
	Realtime RealtimeConfig
}

// RealtimeConfig carries the browser origins allowed to open a WebSocket.
// Handshakes carrying an Authorization header are native and skip this list.
type RealtimeConfig struct {
	AllowedOrigins []string
}

type R2Config struct {
	AccountID       string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	PublicBaseURL   url.URL
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
	JWTSecret       string
	DevExposeCodes  bool
	AppleAudience   string
	GoogleAudiences []string
}

const (
	DefaultStartupTimeout  = 10 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
)

// LoadDotEnv reads .env into the process environment.
func LoadDotEnv() error {
	return godotenv.Load()
}

func Load() (Config, error) {
	// Missing .env is fine: production takes configuration from the environment.
	_ = LoadDotEnv()

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
	r2Cfg, err := loadR2Config()
	if err != nil {
		return Config{}, err
	}
	realtimeCfg, err := loadRealtimeConfig(mode)
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
		R2:       r2Cfg,
		Realtime: realtimeCfg,
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

var (
	r2AccountPattern = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
	r2BucketPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)
)

func loadRealtimeConfig(mode string) (RealtimeConfig, error) {
	origins := splitList(os.Getenv("WS_ALLOWED_ORIGINS"))
	if len(origins) == 0 && mode != "release" {
		origins = []string{
			"http://localhost:8081",
			"http://127.0.0.1:8081",
			"http://localhost:19006",
			"http://127.0.0.1:19006",
		}
	}
	for _, origin := range origins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Host == "" ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") ||
			(parsed.Path != "" && parsed.Path != "/") ||
			parsed.RawQuery != "" || parsed.Fragment != "" {
			return RealtimeConfig{}, fmt.Errorf("invalid WS_ALLOWED_ORIGINS entry %q", origin)
		}
	}
	return RealtimeConfig{AllowedOrigins: origins}, nil
}

func loadR2Config() (*R2Config, error) {
	names := []string{"R2_ACCOUNT_ID", "R2_BUCKET", "R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_PUBLIC_BASE_URL"}
	values := make([]string, len(names))
	configured := false
	for i, name := range names {
		values[i] = strings.TrimSpace(os.Getenv(name))
		configured = configured || values[i] != ""
	}
	if !configured {
		return nil, nil
	}
	for i, value := range values {
		if value == "" {
			return nil, fmt.Errorf("%s is required when R2 is configured", names[i])
		}
	}
	if !r2AccountPattern.MatchString(values[0]) {
		return nil, fmt.Errorf("R2_ACCOUNT_ID must contain 32 hexadecimal characters")
	}
	if !r2BucketPattern.MatchString(values[1]) {
		return nil, fmt.Errorf("R2_BUCKET must contain 3-63 lowercase letters, digits, or hyphens, starting and ending with a letter or digit")
	}
	for _, name := range names[2:4] {
		if strings.ContainsFunc(os.Getenv(name), unicode.IsControl) {
			return nil, fmt.Errorf("%s must not contain control characters", name)
		}
	}
	publicURL, err := url.Parse(values[4])
	if err != nil || publicURL.Scheme != "https" || publicURL.Hostname() == "" || publicURL.User != nil || publicURL.Opaque != "" || publicURL.Path != "" && publicURL.Path != "/" || publicURL.RawPath != "" || publicURL.RawQuery != "" || publicURL.ForceQuery || publicURL.Fragment != "" || strings.ContainsAny(values[4], "\\%?#") {
		return nil, fmt.Errorf("R2_PUBLIC_BASE_URL must be an HTTPS origin without credentials, a path prefix, a query, or a fragment")
	}
	publicURL.Host = strings.ToLower(publicURL.Host)
	publicURL.Path = ""
	return &R2Config{
		AccountID:       strings.ToLower(values[0]),
		Bucket:          values[1],
		AccessKeyID:     values[2],
		SecretAccessKey: values[3],
		PublicBaseURL:   *publicURL,
	}, nil
}
