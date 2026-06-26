package auth

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type UserStatus string

const (
	StatusPendingVerification UserStatus = "pending_verification"
	StatusActive              UserStatus = "active"
	StatusSuspended           UserStatus = "suspended"
	StatusDeleted             UserStatus = "deleted"
)

type User struct {
	ID           string
	Email        string
	DisplayName  string
	Role         UserRole
	Status       UserStatus
	VerifiedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	LastPassword time.Time
}

type Session struct {
	ID             string
	UserID         string
	SessionHash    string
	CreatedAt      time.Time
	LastSeenAt     time.Time
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	RevocationNote string
	IPSummary      string
	UserAgent      string
}

type AuditEvent struct {
	ActorUserID  *string
	TargetUserID *string
	EventType    string
	Result       string
	IPSummary    string
	UserAgent    string
	RiskFlags    string
	Metadata     string
}

type RegisterCodeInput struct {
	Email           string
	Password        string
	ConfirmPassword string
	IPSummary       string
	UserAgent       string
}

type VerifyCodeInput struct {
	Email     string
	Code      string
	IPSummary string
	UserAgent string
}

type LoginInput struct {
	Email     string
	Password  string
	IPSummary string
	UserAgent string
}

type PasswordResetRequestInput struct {
	Email     string
	IPSummary string
	UserAgent string
}

type PasswordResetVerifyInput struct {
	Email           string
	Proof           string
	NewPassword     string
	ConfirmPassword string
	IPSummary       string
	UserAgent       string
}
