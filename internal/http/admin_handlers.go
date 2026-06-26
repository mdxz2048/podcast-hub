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
	})
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
