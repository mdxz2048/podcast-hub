package http

import (
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/security"
)

type Server struct {
	cfg       config.Config
	auth      *auth.Service
	turnstile security.TurnstileVerifier
}

func NewServer(cfg config.Config, authService *auth.Service, turnstile security.TurnstileVerifier) *Server {
	return &Server{
		cfg:       cfg,
		auth:      authService,
		turnstile: turnstile,
	}
}

func (s *Server) Router() stdhttp.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(s.corsMiddleware)
	router.Use(s.securityHeadersMiddleware)
	router.Use(s.csrfOriginMiddleware)

	router.Options("/*", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register/request-code", s.handleRegisterRequestCode)
		r.Post("/register/verify-code", s.handleRegisterVerifyCode)
		r.Post("/login", s.handleLogin)
		r.Post("/logout", s.handleLogout)
		r.Post("/password-reset/request", s.handlePasswordResetRequest)
		r.Post("/password-reset/verify", s.handlePasswordResetVerify)
		r.Get("/me", s.handleMe)
	})
	return router
}

type errorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

func writeJSON(w stdhttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w stdhttp.ResponseWriter, r *stdhttp.Request, status int, code string, message string) {
	response := errorResponse{}
	response.Error.Code = code
	response.Error.Message = message
	response.Error.RequestID = middleware.GetReqID(r.Context())
	writeJSON(w, status, response)
}

func mapAuthError(err error) (int, string, string) {
	switch {
	case errors.Is(err, auth.ErrInvalidEmail):
		return stdhttp.StatusBadRequest, "invalid_email", "邮箱格式无效。"
	case errors.Is(err, auth.ErrWeakPassword):
		return stdhttp.StatusBadRequest, "weak_password", "密码强度不足。"
	case errors.Is(err, auth.ErrPasswordMismatch):
		return stdhttp.StatusBadRequest, "password_mismatch", "两次输入的密码不一致。"
	case errors.Is(err, auth.ErrRateLimited):
		return stdhttp.StatusTooManyRequests, "rate_limited", "请求过于频繁，请稍后再试。"
	case errors.Is(err, auth.ErrTurnstileRequired):
		return stdhttp.StatusBadRequest, "turnstile_required", "请先完成人机验证。"
	case errors.Is(err, auth.ErrTurnstileFailed):
		return stdhttp.StatusBadRequest, "turnstile_failed", "人机验证失败，请重试。"
	case errors.Is(err, auth.ErrInvalidOrExpiredCode):
		return stdhttp.StatusBadRequest, "invalid_or_expired_code", "验证码错误或已过期。"
	case errors.Is(err, auth.ErrVerificationCodeExpired):
		return stdhttp.StatusBadRequest, "verification_code_expired", "验证码已过期，请重新获取。"
	case errors.Is(err, auth.ErrTooManyAttempts):
		return stdhttp.StatusTooManyRequests, "too_many_attempts", "尝试次数过多，请稍后再试。"
	case errors.Is(err, auth.ErrInvalidCredentials):
		return stdhttp.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误。"
	case errors.Is(err, auth.ErrAccountUnavailable):
		return stdhttp.StatusForbidden, "account_unavailable", "账号当前不可用。"
	case errors.Is(err, auth.ErrNotAuthenticated):
		return stdhttp.StatusUnauthorized, "not_authenticated", "当前未登录。"
	case errors.Is(err, auth.ErrInvalidOrExpiredProof):
		return stdhttp.StatusBadRequest, "invalid_or_expired_proof", "重置凭证错误或已过期。"
	case errors.Is(err, auth.ErrInvalidResetProof):
		return stdhttp.StatusBadRequest, "invalid_reset_proof", "重置验证码错误。"
	case errors.Is(err, auth.ErrResetProofExpired):
		return stdhttp.StatusBadRequest, "reset_proof_expired", "重置验证码已过期。"
	case errors.Is(err, auth.ErrTemporarilyUnavailable):
		return stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "服务暂时不可用，请稍后重试。"
	case errors.Is(err, auth.ErrForbiddenCSRF):
		return stdhttp.StatusForbidden, "csrf_invalid", "请求未通过安全校验。"
	case errors.Is(err, auth.ErrUnsupportedContentType):
		return stdhttp.StatusUnsupportedMediaType, "unsupported_content_type", "请求内容类型必须为 application/json。"
	case errors.Is(err, auth.ErrInvalidJSONBody):
		return stdhttp.StatusBadRequest, "invalid_json_body", "请求体格式无效。"
	default:
		return stdhttp.StatusInternalServerError, "internal_error", "请求暂时无法完成。"
	}
}

func (s *Server) writeAuthError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	status, code, message := mapAuthError(err)
	writeError(w, r, status, code, message)
}

func (s *Server) parseJSONBody(r *stdhttp.Request, target any) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return auth.ErrUnsupportedContentType
	}
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: %s", auth.ErrInvalidJSONBody, err.Error())
	}
	return nil
}

func (s *Server) setSessionCookie(w stdhttp.ResponseWriter, token string) {
	cookie := &stdhttp.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		Expires:  time.Now().Add(s.cfg.SessionTTL),
	}
	if strings.TrimSpace(s.cfg.SessionCookieDomain) != "" {
		cookie.Domain = s.cfg.SessionCookieDomain
	}
	stdhttp.SetCookie(w, cookie)
}

func (s *Server) clearSessionCookie(w stdhttp.ResponseWriter) {
	stdhttp.SetCookie(w, &stdhttp.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (s *Server) setCSRFCookie(w stdhttp.ResponseWriter, token string) {
	stdhttp.SetCookie(w, &stdhttp.Cookie{
		Name:     "podcast_hub_csrf",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		Expires:  time.Now().Add(s.cfg.SessionTTL),
	})
}
