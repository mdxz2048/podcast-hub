package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/auth"
)

type AuthStore struct {
	pool *pgxpool.Pool
}

func NewAuthStore(pool *pgxpool.Pool) *AuthStore {
	return &AuthStore{pool: pool}
}

func (s *AuthStore) CreateOrUpdatePendingRegistration(ctx context.Context, email, passwordHash, codeHash string, expiresAt time.Time, maxAttempts int) (auth.User, error) {
	now := time.Now()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	row := tx.QueryRow(ctx, `
		INSERT INTO users(id, email_normalized, role, status, created_at, updated_at)
		VALUES ($1, $2, 'user', 'pending_verification', $3, $3)
		ON CONFLICT (email_normalized)
		DO UPDATE SET status='pending_verification', updated_at=excluded.updated_at
		RETURNING id::text
	`, uuid.New(), email, now)
	if err := row.Scan(&userID); err != nil {
		return auth.User{}, fmt.Errorf("upsert user: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_credentials(user_id, password_hash, password_hash_algorithm, password_updated_at, failed_login_count)
		VALUES ($1::uuid, $2, 'argon2id', $3, 0)
		ON CONFLICT (user_id) DO UPDATE SET password_hash=excluded.password_hash, password_updated_at=excluded.password_updated_at, failed_login_count=0, locked_until=NULL
	`, userID, passwordHash, now); err != nil {
		return auth.User{}, fmt.Errorf("upsert credential: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE email_verifications
		SET replaced_at=$1
		WHERE email_normalized=$2 AND purpose='register' AND consumed_at IS NULL AND replaced_at IS NULL
	`, now, email); err != nil {
		return auth.User{}, fmt.Errorf("replace old verification: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO email_verifications(id, user_id, email_normalized, purpose, code_hash, expires_at, max_attempts, created_at)
		VALUES ($1, $2::uuid, $3, 'register', $4, $5, $6, $7)
	`, uuid.New(), userID, email, codeHash, expiresAt, maxAttempts, now); err != nil {
		return auth.User{}, fmt.Errorf("insert verification code: %w", err)
	}
	user, err := s.getUserByIDTx(ctx, tx, userID)
	if err != nil {
		return auth.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return auth.User{}, fmt.Errorf("commit tx: %w", err)
	}
	return user, nil
}

func (s *AuthStore) VerifyRegistrationCode(ctx context.Context, email, codeHash string, now time.Time) (auth.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var verificationID, userID, dbCodeHash string
	var expiresAt time.Time
	var consumedAt *time.Time
	var attempts, maxAttempts int
	err = tx.QueryRow(ctx, `
		SELECT id::text, user_id::text, code_hash, expires_at, consumed_at, attempt_count, max_attempts
		FROM email_verifications
		WHERE email_normalized=$1 AND purpose='register' AND replaced_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, email).Scan(&verificationID, &userID, &dbCodeHash, &expiresAt, &consumedAt, &attempts, &maxAttempts)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrInvalidOrExpiredCode
		}
		return auth.User{}, fmt.Errorf("lookup verification code: %w", err)
	}
	if consumedAt != nil {
		return auth.User{}, auth.ErrInvalidOrExpiredCode
	}
	if now.After(expiresAt) {
		return auth.User{}, auth.ErrVerificationCodeExpired
	}
	if attempts >= maxAttempts {
		return auth.User{}, auth.ErrTooManyAttempts
	}
	if dbCodeHash != codeHash {
		if _, err := tx.Exec(ctx, `UPDATE email_verifications SET attempt_count=attempt_count+1 WHERE id=$1::uuid`, verificationID); err != nil {
			return auth.User{}, fmt.Errorf("increment verification attempts: %w", err)
		}
		if attempts+1 >= maxAttempts {
			return auth.User{}, auth.ErrTooManyAttempts
		}
		return auth.User{}, auth.ErrInvalidOrExpiredCode
	}
	if _, err := tx.Exec(ctx, `UPDATE email_verifications SET consumed_at=$1 WHERE id=$2::uuid`, now, verificationID); err != nil {
		return auth.User{}, fmt.Errorf("consume verification code: %w", err)
	}
	tag, err := tx.Exec(ctx, `
		UPDATE users SET status='active', verified_at=$1, updated_at=$1
		WHERE id=$2::uuid AND status='pending_verification'
	`, now, userID)
	if err != nil {
		return auth.User{}, fmt.Errorf("activate user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return auth.User{}, auth.ErrAccountUnavailable
	}
	user, err := s.getUserByIDTx(ctx, tx, userID)
	if err != nil {
		return auth.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return auth.User{}, fmt.Errorf("commit verify registration: %w", err)
	}
	return user, nil
}

func (s *AuthStore) FindUserByEmail(ctx context.Context, email string) (auth.User, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, email_normalized, display_name, role, status, created_at, updated_at, verified_at, deleted_at
		FROM users WHERE email_normalized=$1
	`, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, false, nil
		}
		return auth.User{}, false, fmt.Errorf("find user by email: %w", err)
	}
	return user, true, nil
}

func (s *AuthStore) GetCredentialHash(ctx context.Context, userID string) (string, error) {
	var passwordHash string
	if err := s.pool.QueryRow(ctx, `SELECT password_hash FROM user_credentials WHERE user_id=$1::uuid`, userID).Scan(&passwordHash); err != nil {
		return "", fmt.Errorf("get credential hash: %w", err)
	}
	return passwordHash, nil
}

func (s *AuthStore) IncrementFailedLogin(ctx context.Context, userID string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE user_credentials SET failed_login_count=failed_login_count+1, password_updated_at=password_updated_at WHERE user_id=$1::uuid`, userID)
	if err != nil {
		return fmt.Errorf("increment failed login: %w", err)
	}
	return nil
}

func (s *AuthStore) ResetFailedLogin(ctx context.Context, userID string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE user_credentials SET failed_login_count=0, locked_until=NULL WHERE user_id=$1::uuid`, userID)
	if err != nil {
		return fmt.Errorf("reset failed login: %w", err)
	}
	return nil
}

func (s *AuthStore) UpdatePassword(ctx context.Context, userID, passwordHash string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE user_credentials
		SET password_hash=$1, password_updated_at=$2, failed_login_count=0, locked_until=NULL
		WHERE user_id=$3::uuid
	`, passwordHash, now, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (s *AuthStore) IssuePasswordReset(ctx context.Context, userID, email, proofHash string, expiresAt time.Time, maxAttempts int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	if _, err := tx.Exec(ctx, `
		UPDATE password_resets
		SET replaced_at=$1
		WHERE email_normalized=$2 AND consumed_at IS NULL AND replaced_at IS NULL
	`, now, email); err != nil {
		return fmt.Errorf("replace old reset proof: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO password_resets(id, user_id, email_normalized, proof_hash, proof_type, expires_at, max_attempts, created_at)
		VALUES ($1, $2::uuid, $3, $4, 'code', $5, $6, $7)
	`, uuid.New(), userID, email, proofHash, expiresAt, maxAttempts, now); err != nil {
		return fmt.Errorf("insert reset proof: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *AuthStore) VerifyPasswordReset(ctx context.Context, email, proofHash string, now time.Time) (auth.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var resetID, userID, dbProofHash string
	var expiresAt time.Time
	var consumedAt *time.Time
	var attempts, maxAttempts int
	err = tx.QueryRow(ctx, `
		SELECT id::text, user_id::text, proof_hash, expires_at, consumed_at, attempt_count, max_attempts
		FROM password_resets
		WHERE email_normalized=$1 AND replaced_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, email).Scan(&resetID, &userID, &dbProofHash, &expiresAt, &consumedAt, &attempts, &maxAttempts)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrInvalidOrExpiredProof
		}
		return auth.User{}, fmt.Errorf("lookup reset proof: %w", err)
	}
	if consumedAt != nil {
		return auth.User{}, auth.ErrInvalidOrExpiredProof
	}
	if now.After(expiresAt) {
		return auth.User{}, auth.ErrResetProofExpired
	}
	if attempts >= maxAttempts {
		return auth.User{}, auth.ErrTooManyAttempts
	}
	if dbProofHash != proofHash {
		if _, err := tx.Exec(ctx, `UPDATE password_resets SET attempt_count=attempt_count+1 WHERE id=$1::uuid`, resetID); err != nil {
			return auth.User{}, fmt.Errorf("increment reset attempts: %w", err)
		}
		if attempts+1 >= maxAttempts {
			return auth.User{}, auth.ErrTooManyAttempts
		}
		return auth.User{}, auth.ErrInvalidResetProof
	}
	if _, err := tx.Exec(ctx, `UPDATE password_resets SET consumed_at=$1 WHERE id=$2::uuid`, now, resetID); err != nil {
		return auth.User{}, fmt.Errorf("consume reset proof: %w", err)
	}
	user, err := s.getUserByIDTx(ctx, tx, userID)
	if err != nil {
		return auth.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return auth.User{}, fmt.Errorf("commit reset verify: %w", err)
	}
	return user, nil
}

func (s *AuthStore) CreateSession(ctx context.Context, session auth.Session) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_sessions(id, user_id, session_hash, created_at, last_seen_at, expires_at, ip_summary, user_agent_summary, device_label)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, '')
	`, uuid.New(), session.UserID, session.SessionHash, session.CreatedAt, session.LastSeenAt, session.ExpiresAt, session.IPSummary, session.UserAgent)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *AuthStore) GetSessionWithUserByHash(ctx context.Context, sessionHash string, now time.Time) (auth.Session, auth.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT
			s.id::text, s.user_id::text, s.session_hash, s.created_at, s.last_seen_at, s.expires_at, s.revoked_at, COALESCE(s.revocation_reason, ''), s.ip_summary, s.user_agent_summary,
			u.id::text, u.email_normalized, u.display_name, u.role, u.status, u.created_at, u.updated_at, u.verified_at, u.deleted_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.session_hash=$1
	`, sessionHash)
	var sess auth.Session
	var user auth.User
	err := row.Scan(
		&sess.ID, &sess.UserID, &sess.SessionHash, &sess.CreatedAt, &sess.LastSeenAt, &sess.ExpiresAt, &sess.RevokedAt, &sess.RevocationNote, &sess.IPSummary, &sess.UserAgent,
		&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.VerifiedAt, &user.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
		}
		return auth.Session{}, auth.User{}, fmt.Errorf("fetch session and user: %w", err)
	}
	if sess.RevokedAt != nil || now.After(sess.ExpiresAt) {
		return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
	}
	_, _ = s.pool.Exec(ctx, `UPDATE user_sessions SET last_seen_at=$1 WHERE session_hash=$2`, now, sessionHash)
	return sess, user, nil
}

func (s *AuthStore) RevokeSessionByHash(ctx context.Context, sessionHash, reason string, now time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at=$1, revocation_reason=$2
		WHERE session_hash=$3 AND revoked_at IS NULL
	`, now, reason, sessionHash)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return auth.ErrNotAuthenticated
	}
	return nil
}

func (s *AuthStore) RevokeAllSessionsByUserID(ctx context.Context, userID, reason string, now time.Time) error {
	if _, err := s.pool.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at=$1, revocation_reason=$2
		WHERE user_id=$3::uuid AND revoked_at IS NULL
	`, now, reason, userID); err != nil {
		return fmt.Errorf("revoke all user sessions: %w", err)
	}
	return nil
}

func (s *AuthStore) InsertAuditLog(ctx context.Context, event auth.AuditEvent) error {
	metadata := json.RawMessage(event.Metadata)
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO auth_audit_logs(id, actor_user_id, target_user_id, event_type, result, ip_summary, user_agent_summary, risk_flags, metadata_redacted, created_at)
		VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9::jsonb, $10)
	`, uuid.New(), nullableUUID(event.ActorUserID), nullableUUID(event.TargetUserID), event.EventType, event.Result, event.IPSummary, event.UserAgent, event.RiskFlags, metadata, time.Now())
	if err != nil {
		return fmt.Errorf("insert auth audit log: %w", err)
	}
	return nil
}

func (s *AuthStore) CreateAdminUser(ctx context.Context, email, passwordHash string, now time.Time) (auth.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	userID := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO users(id, email_normalized, role, status, verified_at, created_at, updated_at)
		VALUES ($1, $2, 'admin', 'active', $3, $3, $3)
	`, userID, email, now); err != nil {
		return auth.User{}, fmt.Errorf("insert admin user: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_credentials(user_id, password_hash, password_hash_algorithm, password_updated_at, failed_login_count)
		VALUES ($1, $2, 'argon2id', $3, 0)
	`, userID, passwordHash, now); err != nil {
		return auth.User{}, fmt.Errorf("insert admin credential: %w", err)
	}
	user, err := s.getUserByIDTx(ctx, tx, userID.String())
	if err != nil {
		return auth.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return auth.User{}, fmt.Errorf("commit create admin: %w", err)
	}
	return user, nil
}

func (s *AuthStore) PromoteUserToAdmin(ctx context.Context, userID, passwordHash string, now time.Time) (auth.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE users
		SET role='admin', status='active', verified_at=COALESCE(verified_at, $1), updated_at=$1
		WHERE id=$2::uuid
	`, now, userID)
	if err != nil {
		return auth.User{}, fmt.Errorf("promote user role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return auth.User{}, auth.ErrAccountUnavailable
	}
	if _, err := tx.Exec(ctx, `
		UPDATE user_credentials
		SET password_hash=$1, password_hash_algorithm='argon2id', password_updated_at=$2, failed_login_count=0, locked_until=NULL
		WHERE user_id=$3::uuid
	`, passwordHash, now, userID); err != nil {
		return auth.User{}, fmt.Errorf("update promoted credential: %w", err)
	}
	user, err := s.getUserByIDTx(ctx, tx, userID)
	if err != nil {
		return auth.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return auth.User{}, fmt.Errorf("commit promote admin: %w", err)
	}
	return user, nil
}

func (s *AuthStore) getUserByIDTx(ctx context.Context, tx pgx.Tx, userID string) (auth.User, error) {
	return scanUser(tx.QueryRow(ctx, `
		SELECT id::text, email_normalized, display_name, role, status, created_at, updated_at, verified_at, deleted_at
		FROM users WHERE id=$1::uuid
	`, userID))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (auth.User, error) {
	var user auth.User
	err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.VerifiedAt, &user.DeletedAt)
	if err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func nullableUUID(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}
