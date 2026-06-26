package config

import "testing"

func TestLoad_ProductionMissingSecurityConfigFails(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("FRONTEND_ORIGIN", "http://127.0.0.1:5173")
	t.Setenv("SESSION_PEPPER", "pepper")
	t.Setenv("AUTH_CODE_PEPPER", "pepper2")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("TURNSTILE_MODE", "cloudflare")
	t.Setenv("TURNSTILE_SECRET_KEY", "")
	t.Setenv("REDIS_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected config load to fail when production security vars are missing")
	}
}
