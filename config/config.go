package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv         string
	HTTPAddr       string
	FrontendOrigin string

	DatabaseURL string
	RedisURL    string

	SessionCookieName   string
	SessionTTL          time.Duration
	SessionCookieSecure bool
	SessionCookieDomain string
	SessionPepper       string
	AuthCodePepper      string

	TurnstileMode      string
	TurnstileSiteKey   string
	TurnstileSecretKey string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	CSRFHeaderName string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		HTTPAddr:            getEnv("HTTP_ADDR", ":8080"),
		FrontendOrigin:      getEnv("FRONTEND_ORIGIN", "http://127.0.0.1:5173"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		SessionCookieName:   getEnv("SESSION_COOKIE_NAME", "podcast_hub_session"),
		SessionTTL:          durationFromSeconds("SESSION_TTL_SECONDS", 60*60*24*14),
		SessionCookieSecure: parseBool(getEnv("SESSION_COOKIE_SECURE", "false")),
		SessionCookieDomain: os.Getenv("SESSION_COOKIE_DOMAIN"),
		SessionPepper:       os.Getenv("SESSION_PEPPER"),
		AuthCodePepper:      os.Getenv("AUTH_CODE_PEPPER"),
		TurnstileMode:       getEnv("TURNSTILE_MODE", "mock"),
		TurnstileSiteKey:    os.Getenv("VITE_TURNSTILE_SITE_KEY"),
		TurnstileSecretKey:  os.Getenv("TURNSTILE_SECRET_KEY"),
		SMTPHost:            getEnv("SMTP_HOST", "127.0.0.1"),
		SMTPPort:            parseInt(getEnv("SMTP_PORT", "1025"), 1025),
		SMTPUsername:        os.Getenv("SMTP_USERNAME"),
		SMTPPassword:        os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:            getEnv("SMTP_FROM", "no-reply@example.invalid"),
		CSRFHeaderName:      getEnv("CSRF_HEADER_NAME", "X-CSRF-Token"),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func (c Config) Validate() error {
	var missing []string
	if strings.TrimSpace(c.DatabaseURL) == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if strings.TrimSpace(c.SessionPepper) == "" {
		missing = append(missing, "SESSION_PEPPER")
	}
	if strings.TrimSpace(c.AuthCodePepper) == "" {
		missing = append(missing, "AUTH_CODE_PEPPER")
	}
	if strings.TrimSpace(c.FrontendOrigin) == "" {
		missing = append(missing, "FRONTEND_ORIGIN")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	switch c.TurnstileMode {
	case "mock", "cloudflare":
	default:
		return errors.New("TURNSTILE_MODE must be one of: mock, cloudflare")
	}
	if c.IsProduction() {
		if c.TurnstileMode != "cloudflare" {
			return errors.New("production must use TURNSTILE_MODE=cloudflare")
		}
		if strings.TrimSpace(c.TurnstileSecretKey) == "" {
			return errors.New("TURNSTILE_SECRET_KEY is required in production")
		}
		if strings.TrimSpace(c.RedisURL) == "" {
			return errors.New("REDIS_URL is required in production")
		}
		if !c.SessionCookieSecure {
			return errors.New("SESSION_COOKIE_SECURE must be true in production")
		}
	}
	if c.TurnstileMode == "cloudflare" && strings.TrimSpace(c.TurnstileSecretKey) == "" {
		return errors.New("TURNSTILE_SECRET_KEY is required when TURNSTILE_MODE=cloudflare")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseInt(raw string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return n
}

func durationFromSeconds(key string, fallbackSeconds int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(n) * time.Second
}
