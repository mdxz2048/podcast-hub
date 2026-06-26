package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/mail"
	"github.com/mdxz2048/podcast-hub/internal/security"
)

type testStore struct {
	usersByEmail     map[string]User
	passwordHashByID map[string]string
	regCodes         map[string]verificationCode
	resetProofs      map[string]verificationCode
	sessionsByHash   map[string]Session
	auditEvents      []AuditEvent
}

type verificationCode struct {
	hash      string
	expiresAt time.Time
	used      bool
	attempts  int
	max       int
}

func newTestStore() *testStore {
	return &testStore{
		usersByEmail:     map[string]User{},
		passwordHashByID: map[string]string{},
		regCodes:         map[string]verificationCode{},
		resetProofs:      map[string]verificationCode{},
		sessionsByHash:   map[string]Session{},
	}
}

func (s *testStore) CreateOrUpdatePendingRegistration(_ context.Context, email, passwordHash, codeHash string, expiresAt time.Time, maxAttempts int) (User, error) {
	user, found := s.usersByEmail[email]
	if !found {
		user = User{
			ID:        "user_" + strings.ReplaceAll(email, "@", "_"),
			Email:     email,
			Role:      RoleUser,
			Status:    StatusPendingVerification,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	user.Status = StatusPendingVerification
	user.UpdatedAt = time.Now()
	s.usersByEmail[email] = user
	s.passwordHashByID[user.ID] = passwordHash
	s.regCodes[email] = verificationCode{hash: codeHash, expiresAt: expiresAt, max: maxAttempts}
	return user, nil
}

func (s *testStore) VerifyRegistrationCode(_ context.Context, email, codeHash string, now time.Time) (User, error) {
	code, ok := s.regCodes[email]
	if !ok || code.used || now.After(code.expiresAt) {
		return User{}, ErrInvalidOrExpiredCode
	}
	if code.attempts >= code.max {
		return User{}, ErrTooManyAttempts
	}
	if code.hash != codeHash {
		code.attempts++
		s.regCodes[email] = code
		if code.attempts >= code.max {
			return User{}, ErrTooManyAttempts
		}
		return User{}, ErrInvalidOrExpiredCode
	}
	code.used = true
	s.regCodes[email] = code
	user := s.usersByEmail[email]
	user.Status = StatusActive
	nowCopy := now
	user.VerifiedAt = &nowCopy
	s.usersByEmail[email] = user
	return user, nil
}

func (s *testStore) FindUserByEmail(_ context.Context, email string) (User, bool, error) {
	user, ok := s.usersByEmail[email]
	return user, ok, nil
}

func (s *testStore) GetCredentialHash(_ context.Context, userID string) (string, error) {
	hash, ok := s.passwordHashByID[userID]
	if !ok {
		return "", errors.New("credential not found")
	}
	return hash, nil
}

func (s *testStore) IncrementFailedLogin(_ context.Context, _ string, _ time.Time) error { return nil }
func (s *testStore) ResetFailedLogin(_ context.Context, _ string, _ time.Time) error     { return nil }

func (s *testStore) UpdatePassword(_ context.Context, userID, passwordHash string, _ time.Time) error {
	s.passwordHashByID[userID] = passwordHash
	return nil
}

func (s *testStore) IssuePasswordReset(_ context.Context, _ string, email, proofHash string, expiresAt time.Time, maxAttempts int) error {
	s.resetProofs[email] = verificationCode{hash: proofHash, expiresAt: expiresAt, max: maxAttempts}
	return nil
}

func (s *testStore) VerifyPasswordReset(_ context.Context, email, proofHash string, now time.Time) (User, error) {
	code, ok := s.resetProofs[email]
	if !ok || code.used || now.After(code.expiresAt) {
		return User{}, ErrInvalidOrExpiredProof
	}
	if code.attempts >= code.max {
		return User{}, ErrTooManyAttempts
	}
	if code.hash != proofHash {
		code.attempts++
		s.resetProofs[email] = code
		if code.attempts >= code.max {
			return User{}, ErrTooManyAttempts
		}
		return User{}, ErrInvalidOrExpiredProof
	}
	code.used = true
	s.resetProofs[email] = code
	user, ok := s.usersByEmail[email]
	if !ok {
		return User{}, ErrInvalidOrExpiredProof
	}
	return user, nil
}

func (s *testStore) CreateSession(_ context.Context, session Session) error {
	s.sessionsByHash[session.SessionHash] = session
	return nil
}

func (s *testStore) GetSessionWithUserByHash(_ context.Context, sessionHash string, now time.Time) (Session, User, error) {
	session, ok := s.sessionsByHash[sessionHash]
	if !ok {
		return Session{}, User{}, ErrNotAuthenticated
	}
	if session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return Session{}, User{}, ErrNotAuthenticated
	}
	for _, user := range s.usersByEmail {
		if user.ID == session.UserID {
			return session, user, nil
		}
	}
	return Session{}, User{}, ErrNotAuthenticated
}

func (s *testStore) RevokeSessionByHash(_ context.Context, sessionHash, reason string, now time.Time) error {
	session, ok := s.sessionsByHash[sessionHash]
	if !ok {
		return ErrNotAuthenticated
	}
	session.RevokedAt = &now
	session.RevocationNote = reason
	s.sessionsByHash[sessionHash] = session
	return nil
}

func (s *testStore) RevokeAllSessionsByUserID(_ context.Context, userID, reason string, now time.Time) error {
	for hash, session := range s.sessionsByHash {
		if session.UserID != userID {
			continue
		}
		session.RevokedAt = &now
		session.RevocationNote = reason
		s.sessionsByHash[hash] = session
	}
	return nil
}

func (s *testStore) InsertAuditLog(_ context.Context, event AuditEvent) error {
	s.auditEvents = append(s.auditEvents, event)
	return nil
}

type alwaysAllowLimiter struct{}

func (alwaysAllowLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) (bool, time.Duration, error) {
	return true, time.Second, nil
}

type failingLimiter struct{}

func (failingLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) (bool, time.Duration, error) {
	return false, 0, errors.New("redis unavailable")
}

func newTestService(store *testStore) (*Service, *mail.MemoryMailer) {
	mailer := &mail.MemoryMailer{}
	svc := NewService(store, mailer, security.MockTurnstileVerifier{}, alwaysAllowLimiter{}, Options{
		SessionPepper:    "session-pepper",
		AuthCodePepper:   "code-pepper",
		SessionTTL:       24 * time.Hour,
		CodeTTL:          5 * time.Minute,
		ResetProofTTL:    5 * time.Minute,
		CodeMaxAttempts:  3,
		ResetMaxAttempts: 3,
	})
	return svc, mailer
}

func TestRegistrationVerifyAndLoginFlow(t *testing.T) {
	store := newTestStore()
	service, mailer := newTestService(store)

	_, _, _, err := service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "user@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "test-agent",
	})
	if err != nil {
		t.Fatalf("request registration code: %v", err)
	}
	if store.usersByEmail["user@example.invalid"].Status != StatusPendingVerification {
		t.Fatalf("expected pending_verification status after register request")
	}
	code := mailer.Messages[0].Secret
	verifiedUser, _, err := service.VerifyRegistrationCode(context.Background(), VerifyCodeInput{
		Email:     "user@example.invalid",
		Code:      code,
		IPSummary: "127.0.0.1",
		UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("verify registration code: %v", err)
	}
	if verifiedUser.Status != StatusActive {
		t.Fatalf("expected active status after verification")
	}
	_, _, err = service.VerifyRegistrationCode(context.Background(), VerifyCodeInput{
		Email:     "user@example.invalid",
		Code:      code,
		IPSummary: "127.0.0.1",
		UserAgent: "test-agent",
	})
	if !errors.Is(err, ErrInvalidOrExpiredCode) {
		t.Fatalf("expected single-use verification code")
	}
	_, _, err = service.Login(context.Background(), LoginInput{
		Email:     "user@example.invalid",
		Password:  "strong-password-123",
		IPSummary: "127.0.0.1",
		UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("login should succeed: %v", err)
	}
}

func TestRegistrationCodeResendInvalidatesOldCode(t *testing.T) {
	store := newTestStore()
	service, mailer := newTestService(store)

	_, _, _, _ = service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "resend@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	firstCode := mailer.Messages[len(mailer.Messages)-1].Secret
	_, _, _, _ = service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "resend@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	secondCode := mailer.Messages[len(mailer.Messages)-1].Secret
	if firstCode == secondCode {
		t.Fatalf("expected rotated verification code")
	}
	_, _, err := service.VerifyRegistrationCode(context.Background(), VerifyCodeInput{
		Email:     "resend@example.invalid",
		Code:      firstCode,
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	if !errors.Is(err, ErrInvalidOrExpiredCode) {
		t.Fatalf("expected old code to be invalid after resend")
	}
}

func TestRegistrationCodeAttemptLimit(t *testing.T) {
	store := newTestStore()
	service, _ := newTestService(store)
	_, _, _, _ = service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "attempts@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	var err error
	for i := 0; i < 3; i++ {
		_, _, err = service.VerifyRegistrationCode(context.Background(), VerifyCodeInput{
			Email:     "attempts@example.invalid",
			Code:      "WRONG1",
			IPSummary: "127.0.0.1",
			UserAgent: "ua",
		})
	}
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected too many attempts error, got %v", err)
	}
}

func TestLoginFailureAndSuspendedAccount(t *testing.T) {
	store := newTestStore()
	service, _ := newTestService(store)
	_, _, _, _ = service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "suspended@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	user := store.usersByEmail["suspended@example.invalid"]
	user.Status = StatusSuspended
	store.usersByEmail["suspended@example.invalid"] = user

	_, _, err := service.Login(context.Background(), LoginInput{
		Email:     "suspended@example.invalid",
		Password:  "strong-password-123",
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	if !errors.Is(err, ErrAccountUnavailable) {
		t.Fatalf("expected suspended user to be blocked")
	}

	_, _, err = service.Login(context.Background(), LoginInput{
		Email:     "unknown@example.invalid",
		Password:  "wrong-password",
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected generic invalid credentials")
	}
}

func TestSuspendedAccountCannotRestartRegistration(t *testing.T) {
	store := newTestStore()
	service, _ := newTestService(store)
	store.usersByEmail["blocked@example.invalid"] = User{
		ID:        "blocked-user",
		Email:     "blocked@example.invalid",
		Role:      RoleUser,
		Status:    StatusSuspended,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _, _, err := service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "blocked@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	if !errors.Is(err, ErrAccountUnavailable) {
		t.Fatalf("expected suspended account registration to be blocked")
	}
}

func TestLogoutRevokesCurrentSession(t *testing.T) {
	store := newTestStore()
	service, mailer := newTestService(store)
	_, _, _, _ = service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "logout@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	code := mailer.Messages[0].Secret
	_, _, _ = service.VerifyRegistrationCode(context.Background(), VerifyCodeInput{
		Email:     "logout@example.invalid",
		Code:      code,
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	_, token, err := service.Login(context.Background(), LoginInput{
		Email:     "logout@example.invalid",
		Password:  "strong-password-123",
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if err := service.Logout(context.Background(), token, "127.0.0.1", "ua"); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if _, _, err := service.ResolveSession(context.Background(), token); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("expected session to be invalid after logout")
	}
}

func TestPasswordResetRevokesSessions(t *testing.T) {
	store := newTestStore()
	service, mailer := newTestService(store)
	_, _, _, _ = service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "reset@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	code := mailer.Messages[0].Secret
	_, _, _ = service.VerifyRegistrationCode(context.Background(), VerifyCodeInput{
		Email:     "reset@example.invalid",
		Code:      code,
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	_, token, err := service.Login(context.Background(), LoginInput{
		Email:     "reset@example.invalid",
		Password:  "strong-password-123",
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := service.RequestPasswordReset(context.Background(), PasswordResetRequestInput{
		Email:     "reset@example.invalid",
		IPSummary: "127.0.0.1",
		UserAgent: "ua",
	}); err != nil {
		t.Fatalf("request password reset: %v", err)
	}
	var resetCode string
	for _, message := range mailer.Messages {
		if message.Kind == "password_reset_proof" {
			resetCode = message.Secret
		}
	}
	if resetCode == "" {
		t.Fatalf("missing password reset proof")
	}
	if err := service.VerifyPasswordReset(context.Background(), PasswordResetVerifyInput{
		Email:           "reset@example.invalid",
		Proof:           resetCode,
		NewPassword:     "new-strong-password-456",
		ConfirmPassword: "new-strong-password-456",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	}); err != nil {
		t.Fatalf("verify password reset: %v", err)
	}
	if _, _, err := service.ResolveSession(context.Background(), token); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("expected old session revoked after password reset")
	}
}

func TestRateLimiterFailureFailsClosed(t *testing.T) {
	store := newTestStore()
	mailer := &mail.MemoryMailer{}
	service := NewService(store, mailer, security.MockTurnstileVerifier{}, failingLimiter{}, Options{
		SessionPepper:    "session-pepper",
		AuthCodePepper:   "code-pepper",
		SessionTTL:       24 * time.Hour,
		CodeTTL:          5 * time.Minute,
		ResetProofTTL:    5 * time.Minute,
		CodeMaxAttempts:  3,
		ResetMaxAttempts: 3,
	})
	_, _, _, err := service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "user@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	if !errors.Is(err, ErrTemporarilyUnavailable) {
		t.Fatalf("expected fail-closed on limiter failure")
	}
}

func TestAuditLogsAreRedacted(t *testing.T) {
	store := newTestStore()
	service, _ := newTestService(store)
	_, _, _, err := service.RequestRegistrationCode(context.Background(), RegisterCodeInput{
		Email:           "audit@example.invalid",
		Password:        "strong-password-123",
		ConfirmPassword: "strong-password-123",
		IPSummary:       "127.0.0.1",
		UserAgent:       "ua",
	})
	if err != nil {
		t.Fatalf("request registration code: %v", err)
	}
	for _, event := range store.auditEvents {
		if strings.Contains(strings.ToLower(fmt.Sprintf("%v", event)), "strong-password") {
			t.Fatalf("audit logs must not contain plaintext password")
		}
	}
}
