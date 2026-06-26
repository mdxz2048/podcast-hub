package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/security"
	"github.com/mdxz2048/podcast-hub/internal/users"
)

var (
	ErrInvalidEmail        = errors.New("invalid_email")
	ErrPasswordRequired    = errors.New("password_required")
	ErrPromoteRequiresDev  = errors.New("promote_requires_development")
	ErrUserExistsNotAdmin  = errors.New("user_exists_not_admin")
	ErrCreateAdminRejected = errors.New("create_admin_rejected")
)

type Store interface {
	FindUserByEmail(ctx context.Context, email string) (auth.User, bool, error)
	CreateAdminUser(ctx context.Context, email, passwordHash string, now time.Time) (auth.User, error)
	PromoteUserToAdmin(ctx context.Context, userID, passwordHash string, now time.Time) (auth.User, error)
	InsertAuditLog(ctx context.Context, event auth.AuditEvent) error
}

type SeedInput struct {
	Email        string
	Password     string
	AllowPromote bool
	AppEnv       string
	IPSummary    string
	UserAgent    string
}

type SeedResult string

const (
	SeedCreated    SeedResult = "created"
	SeedIdempotent SeedResult = "idempotent"
	SeedPromoted   SeedResult = "promoted"
)

type Service struct {
	store  Store
	hasher security.Argon2idHasher
}

func NewService(store Store) *Service {
	return &Service{
		store:  store,
		hasher: security.DefaultArgon2idHasher(),
	}
}

func (s *Service) SeedAdmin(ctx context.Context, in SeedInput) (SeedResult, auth.User, error) {
	email, err := users.NormalizeEmail(in.Email)
	if err != nil {
		return "", auth.User{}, ErrInvalidEmail
	}
	password := strings.TrimSpace(in.Password)
	if password == "" {
		return "", auth.User{}, ErrPasswordRequired
	}
	passwordHash, err := s.hasher.HashPassword(password)
	if err != nil {
		return "", auth.User{}, auth.ErrWeakPassword
	}
	now := time.Now()

	foundUser, found, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return "", auth.User{}, fmt.Errorf("find user: %w", err)
	}
	if found {
		if foundUser.Role == auth.RoleAdmin {
			_ = s.store.InsertAuditLog(ctx, auth.AuditEvent{
				TargetUserID: &foundUser.ID,
				EventType:    "auth.admin_seed_idempotent",
				Result:       "already_admin",
				IPSummary:    in.IPSummary,
				UserAgent:    truncate(in.UserAgent, 200),
				Metadata:     `{"email_hint":"` + users.EmailHint(email) + `"}`,
			})
			return SeedIdempotent, foundUser, nil
		}
		if !in.AllowPromote {
			_ = s.store.InsertAuditLog(ctx, auth.AuditEvent{
				TargetUserID: &foundUser.ID,
				EventType:    "auth.admin_seed_blocked",
				Result:       "existing_non_admin",
				IPSummary:    in.IPSummary,
				UserAgent:    truncate(in.UserAgent, 200),
			})
			return "", auth.User{}, ErrUserExistsNotAdmin
		}
		if !strings.EqualFold(in.AppEnv, "development") {
			_ = s.store.InsertAuditLog(ctx, auth.AuditEvent{
				TargetUserID: &foundUser.ID,
				EventType:    "auth.admin_seed_blocked",
				Result:       "promote_not_allowed_env",
				IPSummary:    in.IPSummary,
				UserAgent:    truncate(in.UserAgent, 200),
			})
			return "", auth.User{}, ErrPromoteRequiresDev
		}
		user, err := s.store.PromoteUserToAdmin(ctx, foundUser.ID, passwordHash, now)
		if err != nil {
			return "", auth.User{}, fmt.Errorf("promote user to admin: %w", err)
		}
		_ = s.store.InsertAuditLog(ctx, auth.AuditEvent{
			TargetUserID: &user.ID,
			EventType:    "auth.admin_seed_promoted",
			Result:       "success",
			IPSummary:    in.IPSummary,
			UserAgent:    truncate(in.UserAgent, 200),
		})
		return SeedPromoted, user, nil
	}

	user, err := s.store.CreateAdminUser(ctx, email, passwordHash, now)
	if err != nil {
		return "", auth.User{}, fmt.Errorf("create admin user: %w", err)
	}
	_ = s.store.InsertAuditLog(ctx, auth.AuditEvent{
		TargetUserID: &user.ID,
		EventType:    "auth.admin_seed_created",
		Result:       "success",
		IPSummary:    in.IPSummary,
		UserAgent:    truncate(in.UserAgent, 200),
		Metadata:     `{"email_hint":"` + users.EmailHint(email) + `"}`,
	})
	return SeedCreated, user, nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
