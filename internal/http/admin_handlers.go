package http

import (
	"context"
	"net"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/auth"
)

func (s *Server) handleAdminMe(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := authUserFromContext(r.Context())
	if !ok {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"admin": map[string]any{
			"id":     user.ID,
			"email":  user.Email,
			"role":   user.Role,
			"status": user.Status,
		},
	})
}

func (s *Server) handleAdminSystemStatus(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.cfg.IsProduction() {
		writeJSON(w, stdhttp.StatusOK, map[string]any{
			"api":    "ok",
			"detail": "restricted",
			"runner": s.runnerStatus(),
		})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"api": "ok",
		"dependencies": map[string]any{
			"database": s.pingDatabase(ctx),
			"redis":    s.pingRedis(ctx),
			"mailpit":  s.pingSMTP(ctx),
		},
		"runner": s.runnerStatus(),
	})
}

func (s *Server) handleHealthz(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleReadyz(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	db := s.pingDatabase(ctx)
	redis := s.pingRedis(ctx)
	if db == "ok" && (redis == "ok" || redis == "unknown") {
		writeJSON(w, stdhttp.StatusOK, map[string]any{"status": "ready"})
		return
	}
	writeJSON(w, stdhttp.StatusServiceUnavailable, map[string]any{"status": "not_ready"})
}

func (s *Server) runnerStatus() map[string]any {
	mode := s.cfg.RunnerMode
	if mode == "" {
		mode = "disabled"
	}
	status := map[string]any{
		"mode":         mode,
		"can_run_jobs": mode == "docker_trusted_admin",
	}
	if mode == "disabled" {
		status["code"] = "runner_disabled"
		status["reason"] = "RUNNER_MODE=disabled; queued Import Jobs will not execute until a separate Runner is started."
	} else {
		status["code"] = "runner_enabled"
		status["reason"] = "RUNNER_MODE=docker_trusted_admin; only trusted-admin fixture execution is supported."
	}
	return status
}

func (s *Server) pingDatabase(ctx context.Context) string {
	if s.health.DB == nil {
		return "unknown"
	}
	if err := s.health.DB.Ping(ctx); err != nil {
		return "unavailable"
	}
	return "ok"
}

func (s *Server) pingRedis(ctx context.Context) string {
	if s.health.Redis == nil {
		return "unknown"
	}
	if err := s.health.Redis.Ping(ctx).Err(); err != nil {
		return "unavailable"
	}
	return "ok"
}

func (s *Server) pingSMTP(ctx context.Context) string {
	if s.health.SMTPHost == "" || s.health.SMTPPort <= 0 {
		return "unknown"
	}
	address := net.JoinHostPort(s.health.SMTPHost, intToString(s.health.SMTPPort))
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return "unavailable"
	}
	_ = conn.Close()
	return "ok"
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
