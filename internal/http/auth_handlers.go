package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/security"
)

func (s *Server) handleRegisterRequestCode(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		TurnstileToken  string `json:"turnstile_token"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	if err := s.verifyTurnstile(r, req.TurnstileToken); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	emailHint, expiresIn, resendAfter, err := s.auth.RequestRegistrationCode(r.Context(), auth.RegisterCodeInput{
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
		IPSummary:       clientIP(r),
		UserAgent:       userAgentSummary(r),
	})
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"status":               "verification_required",
		"email_hint":           emailHint,
		"expires_in_seconds":   expiresIn,
		"resend_after_seconds": resendAfter,
	})
}

func (s *Server) handleRegisterVerifyCode(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, sessionToken, err := s.auth.VerifyRegistrationCode(r.Context(), auth.VerifyCodeInput{
		Email:     req.Email,
		Code:      req.Code,
		IPSummary: clientIP(r),
		UserAgent: userAgentSummary(r),
	})
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	s.setSessionCookie(w, sessionToken)
	csrfToken, _ := security.NewOpaqueToken()
	s.setCSRFCookie(w, csrfToken)
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"status": "authenticated",
		"user": map[string]any{
			"id":     user.ID,
			"email":  user.Email,
			"role":   user.Role,
			"status": user.Status,
		},
	})
}

func (s *Server) handleLogin(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	if strings.TrimSpace(req.TurnstileToken) != "" {
		if err := s.verifyTurnstile(r, req.TurnstileToken); err != nil {
			s.writeAuthError(w, r, err)
			return
		}
	}
	user, sessionToken, err := s.auth.Login(r.Context(), auth.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IPSummary: clientIP(r),
		UserAgent: userAgentSummary(r),
	})
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	s.setSessionCookie(w, sessionToken)
	csrfToken, _ := security.NewOpaqueToken()
	s.setCSRFCookie(w, csrfToken)
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"status": "authenticated",
		"user": map[string]any{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"role":         user.Role,
			"status":       user.Status,
		},
	})
}

func (s *Server) handleLogout(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	sessionCookie, err := r.Cookie(s.cfg.SessionCookieName)
	if err != nil {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return
	}
	if err := s.auth.Logout(r.Context(), sessionCookie.Value, clientIP(r), userAgentSummary(r)); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	s.clearSessionCookie(w)
	writeJSON(w, stdhttp.StatusOK, map[string]any{"status": "logged_out"})
}

func (s *Server) handlePasswordResetRequest(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req struct {
		Email          string `json:"email"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	if err := s.verifyTurnstile(r, req.TurnstileToken); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	resendAfter, err := s.auth.RequestPasswordReset(r.Context(), auth.PasswordResetRequestInput{
		Email:     req.Email,
		IPSummary: clientIP(r),
		UserAgent: userAgentSummary(r),
	})
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"status":               "reset_instructions_sent_if_account_exists",
		"resend_after_seconds": resendAfter,
	})
}

func (s *Server) handlePasswordResetVerify(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req struct {
		Email           string `json:"email"`
		Proof           string `json:"proof"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	if err := s.auth.VerifyPasswordReset(r.Context(), auth.PasswordResetVerifyInput{
		Email:           req.Email,
		Proof:           req.Proof,
		NewPassword:     req.NewPassword,
		ConfirmPassword: req.ConfirmPassword,
		IPSummary:       clientIP(r),
		UserAgent:       userAgentSummary(r),
	}); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	s.clearSessionCookie(w)
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"status":           "password_reset",
		"sessions_revoked": true,
	})
}

func (s *Server) handleMe(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sessionCookie, err := r.Cookie(s.cfg.SessionCookieName)
	if err != nil {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return
	}
	_, user, err := s.auth.ResolveSession(r.Context(), sessionCookie.Value)
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"user": map[string]any{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"role":         user.Role,
			"status":       user.Status,
		},
	})
}

func (s *Server) verifyTurnstile(r *stdhttp.Request, token string) error {
	if err := s.turnstile.Verify(r.Context(), strings.TrimSpace(token), clientIP(r)); err != nil {
		if errors.Is(err, security.ErrTurnstileRequired) {
			return auth.ErrTurnstileRequired
		}
		return auth.ErrTurnstileFailed
	}
	return nil
}
