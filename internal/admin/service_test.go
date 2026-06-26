package admin

import (
	"context"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/auth"
)

type fakeStore struct {
	usersByEmail map[string]auth.User
	hashByUserID map[string]string
	auditEvents  []auth.AuditEvent
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		usersByEmail: map[string]auth.User{},
		hashByUserID: map[string]string{},
	}
}

func (s *fakeStore) FindUserByEmail(_ context.Context, email string) (auth.User, bool, error) {
	user, ok := s.usersByEmail[email]
	return user, ok, nil
}

func (s *fakeStore) CreateAdminUser(_ context.Context, email, passwordHash string, now time.Time) (auth.User, error) {
	user := auth.User{
		ID:         "admin-created",
		Email:      email,
		Role:       auth.RoleAdmin,
		Status:     auth.StatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
		VerifiedAt: &now,
	}
	s.usersByEmail[email] = user
	s.hashByUserID[user.ID] = passwordHash
	return user, nil
}

func (s *fakeStore) PromoteUserToAdmin(_ context.Context, userID, passwordHash string, now time.Time) (auth.User, error) {
	for email, user := range s.usersByEmail {
		if user.ID == userID {
			user.Role = auth.RoleAdmin
			user.Status = auth.StatusActive
			user.VerifiedAt = &now
			user.UpdatedAt = now
			s.usersByEmail[email] = user
			s.hashByUserID[user.ID] = passwordHash
			return user, nil
		}
	}
	return auth.User{}, auth.ErrInternal
}

func (s *fakeStore) InsertAuditLog(_ context.Context, event auth.AuditEvent) error {
	s.auditEvents = append(s.auditEvents, event)
	return nil
}

func TestSeedAdminCreatesAccount(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store)
	result, user, err := svc.SeedAdmin(context.Background(), SeedInput{
		Email:    "Admin@Example.invalid",
		Password: "StrongPassw0rd!",
		AppEnv:   "development",
	})
	if err != nil {
		t.Fatalf("SeedAdmin returned error: %v", err)
	}
	if result != SeedCreated {
		t.Fatalf("expected created, got %s", result)
	}
	if user.Role != auth.RoleAdmin {
		t.Fatalf("expected admin role, got %s", user.Role)
	}
	if user.Status != auth.StatusActive {
		t.Fatalf("expected active status, got %s", user.Status)
	}
}

func TestSeedAdminIsIdempotentForExistingAdmin(t *testing.T) {
	store := newFakeStore()
	now := time.Now()
	store.usersByEmail["admin@example.invalid"] = auth.User{
		ID:        "existing-admin",
		Email:     "admin@example.invalid",
		Role:      auth.RoleAdmin,
		Status:    auth.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	svc := NewService(store)
	result, user, err := svc.SeedAdmin(context.Background(), SeedInput{
		Email:    "admin@example.invalid",
		Password: "StrongPassw0rd!",
		AppEnv:   "development",
	})
	if err != nil {
		t.Fatalf("SeedAdmin returned error: %v", err)
	}
	if result != SeedIdempotent {
		t.Fatalf("expected idempotent, got %s", result)
	}
	if user.ID != "existing-admin" {
		t.Fatalf("expected existing admin, got %s", user.ID)
	}
}

func TestSeedAdminRejectsImplicitPromotion(t *testing.T) {
	store := newFakeStore()
	now := time.Now()
	store.usersByEmail["user@example.invalid"] = auth.User{
		ID:        "existing-user",
		Email:     "user@example.invalid",
		Role:      auth.RoleUser,
		Status:    auth.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	svc := NewService(store)
	_, _, err := svc.SeedAdmin(context.Background(), SeedInput{
		Email:    "user@example.invalid",
		Password: "StrongPassw0rd!",
		AppEnv:   "development",
	})
	if err != ErrUserExistsNotAdmin {
		t.Fatalf("expected ErrUserExistsNotAdmin, got %v", err)
	}
}

func TestSeedAdminPromoteOnlyAllowedInDevelopment(t *testing.T) {
	store := newFakeStore()
	now := time.Now()
	store.usersByEmail["user@example.invalid"] = auth.User{
		ID:        "existing-user",
		Email:     "user@example.invalid",
		Role:      auth.RoleUser,
		Status:    auth.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	svc := NewService(store)
	_, _, err := svc.SeedAdmin(context.Background(), SeedInput{
		Email:        "user@example.invalid",
		Password:     "StrongPassw0rd!",
		AppEnv:       "production",
		AllowPromote: true,
	})
	if err != ErrPromoteRequiresDev {
		t.Fatalf("expected ErrPromoteRequiresDev, got %v", err)
	}
}
