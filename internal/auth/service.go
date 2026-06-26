package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/security"
	"github.com/mdxz2048/podcast-hub/internal/users"
)

type Service struct {
	store            Store
	mailer           Mailer
	turnstile        TurnstileVerifier
	limiter          RateLimiter
	passwordHasher   security.Argon2idHasher
	sessionPepper    string
	authCodePepper   string
	sessionTTL       time.Duration
	codeTTL          time.Duration
	resetProofTTL    time.Duration
	codeMaxAttempts  int
	resetMaxAttempts int
}

type Options struct {
	SessionPepper    string
	AuthCodePepper   string
	SessionTTL       time.Duration
	CodeTTL          time.Duration
	ResetProofTTL    time.Duration
	CodeMaxAttempts  int
	ResetMaxAttempts int
}

func NewService(store Store, mailer Mailer, turnstile TurnstileVerifier, limiter RateLimiter, opt Options) *Service {
	return &Service{
		store:            store,
		mailer:           mailer,
		turnstile:        turnstile,
		limiter:          limiter,
		passwordHasher:   security.DefaultArgon2idHasher(),
		sessionPepper:    opt.SessionPepper,
		authCodePepper:   opt.AuthCodePepper,
		sessionTTL:       opt.SessionTTL,
		codeTTL:          opt.CodeTTL,
		resetProofTTL:    opt.ResetProofTTL,
		codeMaxAttempts:  opt.CodeMaxAttempts,
		resetMaxAttempts: opt.ResetMaxAttempts,
	}
}

func (s *Service) RequestRegistrationCode(ctx context.Context, in RegisterCodeInput) (string, int, int, error) {
	email, err := users.NormalizeEmail(in.Email)
	if err != nil {
		return "", 0, 0, ErrInvalidEmail
	}
	if strings.TrimSpace(in.Password) != strings.TrimSpace(in.ConfirmPassword) {
		return "", 0, 0, ErrPasswordMismatch
	}
	if err := s.limitAuth(ctx, "register_request", email, in.IPSummary); err != nil {
		return "", 0, 0, err
	}
	existingUser, found, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return "", 0, 0, fmt.Errorf("find existing user: %w", err)
	}
	if found && (existingUser.Status == StatusSuspended || existingUser.Status == StatusDeleted || existingUser.Role == RoleAdmin) {
		_ = s.store.InsertAuditLog(ctx, AuditEvent{
			TargetUserID: ptr(existingUser.ID),
			EventType:    "auth.register_code_blocked",
			Result:       "account_unavailable",
			IPSummary:    in.IPSummary,
			UserAgent:    truncate(in.UserAgent, 200),
		})
		return "", 0, 0, ErrAccountUnavailable
	}
	passwordHash, err := s.passwordHasher.HashPassword(in.Password)
	if err != nil {
		return "", 0, 0, ErrWeakPassword
	}
	rawCode, codeHash := s.newOneTimeCode()
	user, err := s.store.CreateOrUpdatePendingRegistration(ctx, email, passwordHash, codeHash, time.Now().Add(s.codeTTL), s.codeMaxAttempts)
	if err != nil {
		return "", 0, 0, fmt.Errorf("create pending registration: %w", err)
	}
	if err := s.mailer.SendRegistrationCode(ctx, user.Email, rawCode, s.codeTTL); err != nil {
		return "", 0, 0, fmt.Errorf("send registration code: %w", err)
	}
	_ = s.store.InsertAuditLog(ctx, AuditEvent{
		TargetUserID: ptr(user.ID),
		EventType:    "auth.register_code_requested",
		Result:       "success",
		IPSummary:    in.IPSummary,
		UserAgent:    truncate(in.UserAgent, 200),
		Metadata:     `{"email_hint":"` + users.EmailHint(email) + `"}`,
	})
	return users.EmailHint(email), int(s.codeTTL.Seconds()), 60, nil
}

func (s *Service) VerifyRegistrationCode(ctx context.Context, in VerifyCodeInput) (User, string, error) {
	email, err := users.NormalizeEmail(in.Email)
	if err != nil {
		return User{}, "", ErrInvalidEmail
	}
	if err := s.limitAuth(ctx, "register_verify", email, in.IPSummary); err != nil {
		return User{}, "", err
	}
	codeHash := security.HashWithPepper(strings.TrimSpace(in.Code), s.authCodePepper)
	user, err := s.store.VerifyRegistrationCode(ctx, email, codeHash, time.Now())
	if err != nil {
		if errors.Is(err, ErrInvalidOrExpiredCode) || errors.Is(err, ErrVerificationCodeExpired) || errors.Is(err, ErrTooManyAttempts) {
			return User{}, "", err
		}
		return User{}, "", fmt.Errorf("verify code: %w", err)
	}
	token, err := s.createSession(ctx, user, in.IPSummary, in.UserAgent, "auth.session_created")
	if err != nil {
		return User{}, "", err
	}
	return user, token, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (User, string, error) {
	email, err := users.NormalizeEmail(in.Email)
	if err != nil {
		return User{}, "", ErrInvalidCredentials
	}
	if err := s.limitAuth(ctx, "login", email, in.IPSummary); err != nil {
		return User{}, "", err
	}
	user, found, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return User{}, "", fmt.Errorf("find user: %w", err)
	}
	if !found {
		_ = s.store.InsertAuditLog(ctx, AuditEvent{
			EventType: "auth.login_failed", Result: "invalid_credentials", IPSummary: in.IPSummary, UserAgent: truncate(in.UserAgent, 200),
		})
		return User{}, "", ErrInvalidCredentials
	}
	if user.Status != StatusActive {
		_ = s.store.InsertAuditLog(ctx, AuditEvent{
			TargetUserID: ptr(user.ID), EventType: "auth.login_failed", Result: "account_unavailable", IPSummary: in.IPSummary, UserAgent: truncate(in.UserAgent, 200),
		})
		return User{}, "", ErrAccountUnavailable
	}
	hash, err := s.store.GetCredentialHash(ctx, user.ID)
	if err != nil {
		return User{}, "", fmt.Errorf("get credential: %w", err)
	}
	ok, err := s.passwordHasher.VerifyPassword(in.Password, hash)
	if err != nil || !ok {
		_ = s.store.IncrementFailedLogin(ctx, user.ID, time.Now())
		_ = s.store.InsertAuditLog(ctx, AuditEvent{
			TargetUserID: ptr(user.ID), EventType: "auth.login_failed", Result: "invalid_credentials", IPSummary: in.IPSummary, UserAgent: truncate(in.UserAgent, 200),
		})
		return User{}, "", ErrInvalidCredentials
	}
	_ = s.store.ResetFailedLogin(ctx, user.ID, time.Now())
	token, err := s.createSession(ctx, user, in.IPSummary, in.UserAgent, "auth.login_succeeded")
	if err != nil {
		return User{}, "", err
	}
	return user, token, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, in PasswordResetRequestInput) (int, error) {
	email, err := users.NormalizeEmail(in.Email)
	if err != nil {
		return 0, nil
	}
	if err := s.limitAuth(ctx, "password_reset_request", email, in.IPSummary); err != nil {
		return 0, err
	}
	user, found, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("find user for reset: %w", err)
	}
	if found && user.Status == StatusActive {
		rawProof, proofHash := s.newOneTimeCode()
		if err := s.store.IssuePasswordReset(ctx, user.ID, email, proofHash, time.Now().Add(s.resetProofTTL), s.resetMaxAttempts); err != nil {
			return 0, fmt.Errorf("issue reset proof: %w", err)
		}
		if err := s.mailer.SendPasswordResetProof(ctx, email, rawProof, s.resetProofTTL); err != nil {
			return 0, fmt.Errorf("send reset proof: %w", err)
		}
	}
	_ = s.store.InsertAuditLog(ctx, AuditEvent{
		TargetUserID: maybeUserID(found, user.ID), EventType: "auth.password_reset_requested", Result: "accepted", IPSummary: in.IPSummary, UserAgent: truncate(in.UserAgent, 200),
	})
	return 60, nil
}

func (s *Service) VerifyPasswordReset(ctx context.Context, in PasswordResetVerifyInput) error {
	email, err := users.NormalizeEmail(in.Email)
	if err != nil {
		return ErrInvalidOrExpiredProof
	}
	if strings.TrimSpace(in.NewPassword) != strings.TrimSpace(in.ConfirmPassword) {
		return ErrPasswordMismatch
	}
	if err := s.limitAuth(ctx, "password_reset_verify", email, in.IPSummary); err != nil {
		return err
	}
	proofHash := security.HashWithPepper(strings.TrimSpace(in.Proof), s.authCodePepper)
	user, err := s.store.VerifyPasswordReset(ctx, email, proofHash, time.Now())
	if err != nil {
		if errors.Is(err, ErrInvalidOrExpiredProof) || errors.Is(err, ErrInvalidResetProof) || errors.Is(err, ErrResetProofExpired) || errors.Is(err, ErrTooManyAttempts) {
			return err
		}
		return fmt.Errorf("verify reset proof: %w", err)
	}
	newHash, err := s.passwordHasher.HashPassword(in.NewPassword)
	if err != nil {
		return ErrWeakPassword
	}
	now := time.Now()
	if err := s.store.UpdatePassword(ctx, user.ID, newHash, now); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := s.store.RevokeAllSessionsByUserID(ctx, user.ID, "password_reset", now); err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	_ = s.mailer.SendPasswordResetNotice(ctx, email)
	_ = s.store.InsertAuditLog(ctx, AuditEvent{
		TargetUserID: ptr(user.ID), EventType: "auth.password_reset_succeeded", Result: "success", IPSummary: in.IPSummary, UserAgent: truncate(in.UserAgent, 200),
	})
	return nil
}

func (s *Service) ResolveSession(ctx context.Context, token string) (Session, User, error) {
	if strings.TrimSpace(token) == "" {
		return Session{}, User{}, ErrNotAuthenticated
	}
	sessionHash := security.HashWithPepper(token, s.sessionPepper)
	session, user, err := s.store.GetSessionWithUserByHash(ctx, sessionHash, time.Now())
	if err != nil {
		if errors.Is(err, ErrNotAuthenticated) {
			return Session{}, User{}, ErrNotAuthenticated
		}
		return Session{}, User{}, fmt.Errorf("resolve session: %w", err)
	}
	if user.Status != StatusActive {
		return Session{}, User{}, ErrAccountUnavailable
	}
	return session, user, nil
}

func (s *Service) Logout(ctx context.Context, token string, ipSummary string, userAgent string) error {
	if strings.TrimSpace(token) == "" {
		return ErrNotAuthenticated
	}
	sessionHash := security.HashWithPepper(token, s.sessionPepper)
	now := time.Now()
	if err := s.store.RevokeSessionByHash(ctx, sessionHash, "logout", now); err != nil {
		if errors.Is(err, ErrNotAuthenticated) {
			return ErrNotAuthenticated
		}
		return fmt.Errorf("revoke session: %w", err)
	}
	_ = s.store.InsertAuditLog(ctx, AuditEvent{
		EventType: "auth.logout", Result: "success", IPSummary: ipSummary, UserAgent: truncate(userAgent, 200),
	})
	return nil
}

func (s *Service) createSession(ctx context.Context, user User, ipSummary, userAgent, eventType string) (string, error) {
	rawToken, err := security.NewOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	now := time.Now()
	session := Session{
		ID:          rawToken[:16],
		UserID:      user.ID,
		SessionHash: security.HashWithPepper(rawToken, s.sessionPepper),
		CreatedAt:   now,
		LastSeenAt:  now,
		ExpiresAt:   now.Add(s.sessionTTL),
		IPSummary:   ipSummary,
		UserAgent:   truncate(userAgent, 200),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	_ = s.store.InsertAuditLog(ctx, AuditEvent{
		TargetUserID: ptr(user.ID), EventType: eventType, Result: "success", IPSummary: ipSummary, UserAgent: truncate(userAgent, 200),
	})
	return rawToken, nil
}

func (s *Service) newOneTimeCode() (string, string) {
	raw, _ := security.NewOpaqueToken()
	code := strings.ToUpper(raw[:6])
	return code, security.HashWithPepper(code, s.authCodePepper)
}

func (s *Service) limitAuth(ctx context.Context, action, email, ip string) error {
	keys := []string{
		"auth:" + action + ":ip:" + ip,
		"auth:" + action + ":email:" + email,
		"auth:" + action + ":pair:" + ip + ":" + email,
	}
	for _, key := range keys {
		allowed, _, err := s.limiter.Allow(ctx, key, 15, time.Minute)
		if err != nil {
			return ErrTemporarilyUnavailable
		}
		if !allowed {
			return ErrRateLimited
		}
	}
	return nil
}

func ptr(value string) *string { return &value }

func maybeUserID(found bool, id string) *string {
	if !found {
		return nil
	}
	return &id
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
