package auth

import (
	"context"
	"time"
)

type Store interface {
	CreateOrUpdatePendingRegistration(ctx context.Context, email, passwordHash, codeHash string, expiresAt time.Time, maxAttempts int) (User, error)
	VerifyRegistrationCode(ctx context.Context, email, codeHash string, now time.Time) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, bool, error)
	GetCredentialHash(ctx context.Context, userID string) (string, error)
	IncrementFailedLogin(ctx context.Context, userID string, now time.Time) error
	ResetFailedLogin(ctx context.Context, userID string, now time.Time) error
	UpdatePassword(ctx context.Context, userID, passwordHash string, now time.Time) error
	IssuePasswordReset(ctx context.Context, userID, email, proofHash string, expiresAt time.Time, maxAttempts int) error
	VerifyPasswordReset(ctx context.Context, email, proofHash string, now time.Time) (User, error)

	CreateSession(ctx context.Context, session Session) error
	GetSessionWithUserByHash(ctx context.Context, sessionHash string, now time.Time) (Session, User, error)
	RevokeSessionByHash(ctx context.Context, sessionHash, reason string, now time.Time) error
	RevokeAllSessionsByUserID(ctx context.Context, userID, reason string, now time.Time) error

	InsertAuditLog(ctx context.Context, event AuditEvent) error
}

type Mailer interface {
	SendRegistrationCode(ctx context.Context, to, code string, expiresIn time.Duration) error
	SendPasswordResetProof(ctx context.Context, to, proof string, expiresIn time.Duration) error
	SendPasswordResetNotice(ctx context.Context, to string) error
}

type TurnstileVerifier interface {
	Verify(ctx context.Context, token, remoteIP string) error
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error)
}
