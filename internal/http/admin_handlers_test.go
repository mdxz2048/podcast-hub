package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	cfg := config.Config{
		AppEnv:            "development",
		FrontendOrigin:    "http://127.0.0.1:5173",
		SessionCookieName: "podcast_hub_session",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	s := NewServer(cfg, nil, nil, HealthDependencies{})
	return s
}

func TestRequireAuthReturns401WhenNoSession(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(stdhttp.MethodGet, "/admin/me", nil)
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAdminReturns403ForUserRole(t *testing.T) {
	server := testServer(t)
	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		if token != "user-token" {
			return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
		}
		return auth.Session{}, auth.User{ID: "u1", Email: "user@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive}, nil
	}

	req := httptest.NewRequest(stdhttp.MethodGet, "/admin/me", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminMeReturnsCurrentAdmin(t *testing.T) {
	server := testServer(t)
	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		if token != "admin-token" {
			return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
		}
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}

	req := httptest.NewRequest(stdhttp.MethodGet, "/admin/me", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "admin-token"})
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["admin"]["role"] != "admin" {
		t.Fatalf("expected role admin, got %s", body["admin"]["role"])
	}
}

func TestAdminSystemStatusSuccessForAdmin(t *testing.T) {
	server := testServer(t)
	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		if token != "admin-token" {
			return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
		}
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}
	req := httptest.NewRequest(stdhttp.MethodGet, "/admin/system/status", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "admin-token"})
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSuspendedAdminCannotAccessAdminAPI(t *testing.T) {
	server := testServer(t)
	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		if token != "admin-token" {
			return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
		}
		return auth.Session{}, auth.User{}, auth.ErrAccountUnavailable
	}
	req := httptest.NewRequest(stdhttp.MethodGet, "/admin/me", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "admin-token"})
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
