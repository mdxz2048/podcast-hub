package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/publication"
)

type PublicationStore struct{ pool *pgxpool.Pool }

func NewPublicationStore(pool *pgxpool.Pool) *PublicationStore { return &PublicationStore{pool: pool} }

func (s *PublicationStore) FindUserByEmail(ctx context.Context, email string) (auth.User, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, email_normalized, display_name, role, status, created_at, updated_at, verified_at, deleted_at FROM users WHERE email_normalized=$1`, email)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, false, nil
	}
	return user, err == nil, err
}

func (s *PublicationStore) GrantProgramAccess(ctx context.Context, grant publication.ProgramAccessGrant, actorID string) (publication.ProgramAccessGrant, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO program_access_grants(id, user_id, program_id, status, granted_by, reason, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 'active', $4::uuid, $5, $6, $6)
		ON CONFLICT DO NOTHING
		RETURNING id::text, user_id::text, program_id::text, status, reason, created_at, updated_at, revoked_at
	`, grant.ID, grant.UserID, grant.ProgramID, actorID, grant.Reason, grant.CreatedAt)
	item, err := scanProgramAccessGrant(row)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return publication.ProgramAccessGrant{}, err
	}
	row = s.pool.QueryRow(ctx, `SELECT id::text, user_id::text, program_id::text, status, reason, created_at, updated_at, revoked_at FROM program_access_grants WHERE user_id=$1::uuid AND program_id=$2::uuid AND status='active'`, grant.UserID, grant.ProgramID)
	return scanProgramAccessGrant(row)
}

func (s *PublicationStore) ListProgramAccessGrants(ctx context.Context, programID string) ([]publication.ProgramAccessGrant, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, user_id::text, program_id::text, status, reason, created_at, updated_at, revoked_at FROM program_access_grants WHERE program_id=$1::uuid ORDER BY created_at DESC`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.ProgramAccessGrant
	for rows.Next() {
		item, err := scanProgramAccessGrant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PublicationStore) RevokeProgramAccess(ctx context.Context, grantID string, actorID string, reason string, revokedAt time.Time) (publication.ProgramAccessGrant, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE program_access_grants
		SET status='revoked', revoked_by=$2::uuid, reason=$3, revoked_at=$4, updated_at=$4
		WHERE id=$1::uuid AND status='active'
		RETURNING id::text, user_id::text, program_id::text, status, reason, created_at, updated_at, revoked_at
	`, grantID, actorID, reason, revokedAt)
	item, err := scanProgramAccessGrant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.ProgramAccessGrant{}, publication.ErrProgramAccessNotFound
	}
	return item, err
}

func (s *PublicationStore) GetRSSFeed(ctx context.Context, feedID string) (publication.RSSFeed, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, user_id::text, name, token_prefix, status, created_at, last_used_at, rotated_at, revoked_at, expires_at FROM rss_feeds WHERE id=$1::uuid`, feedID)
	item, err := scanRSSFeed(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.RSSFeed{}, false, nil
	}
	return item, err == nil, err
}

func (s *PublicationStore) GetRSSFeedForUser(ctx context.Context, feedID string, userID string) (publication.RSSFeed, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, user_id::text, name, token_prefix, status, created_at, last_used_at, rotated_at, revoked_at, expires_at FROM rss_feeds WHERE id=$1::uuid AND user_id=$2::uuid`, feedID, userID)
	item, err := scanRSSFeed(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.RSSFeed{}, false, nil
	}
	return item, err == nil, err
}

func (s *PublicationStore) ListRSSFeedsByUser(ctx context.Context, userID string) ([]publication.RSSFeed, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, user_id::text, name, token_prefix, status, created_at, last_used_at, rotated_at, revoked_at, expires_at FROM rss_feeds WHERE user_id=$1::uuid ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.RSSFeed
	for rows.Next() {
		item, err := scanRSSFeed(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PublicationStore) ListAdminRSSFeeds(ctx context.Context) ([]publication.AdminRSSFeed, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT rf.id::text, rf.user_id::text, rf.name, rf.token_prefix, rf.status, rf.created_at, rf.last_used_at, rf.rotated_at, rf.revoked_at, rf.expires_at, u.email_normalized
		FROM rss_feeds rf
		JOIN users u ON u.id=rf.user_id
		ORDER BY rf.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.AdminRSSFeed
	for rows.Next() {
		feed, hint, err := scanAdminRSSFeed(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, publication.AdminRSSFeed{RSSFeed: feed, UserEmailHint: hint})
	}
	return items, rows.Err()
}

func (s *PublicationStore) CreateRSSFeed(ctx context.Context, feed publication.RSSFeed, tokenHash string) (publication.RSSFeed, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO rss_feeds(id, user_id, name, token_hash, token_prefix, status, created_at, expires_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, user_id::text, name, token_prefix, status, created_at, last_used_at, rotated_at, revoked_at, expires_at
	`, feed.ID, feed.UserID, feed.Name, tokenHash, feed.TokenPrefix, feed.Status, feed.CreatedAt, feed.ExpiresAt)
	return scanRSSFeed(row)
}

func (s *PublicationStore) RotateRSSFeed(ctx context.Context, feedID string, userID string, tokenHash string, tokenPrefix string, rotatedAt time.Time, expiresAt *time.Time) (publication.RSSFeed, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE rss_feeds
		SET token_hash=$3, token_prefix=$4, status='active', rotated_at=$5, revoked_at=NULL, expires_at=$6
		WHERE id=$1::uuid AND user_id=$2::uuid
		RETURNING id::text, user_id::text, name, token_prefix, status, created_at, last_used_at, rotated_at, revoked_at, expires_at
	`, feedID, userID, tokenHash, tokenPrefix, rotatedAt, expiresAt)
	item, err := scanRSSFeed(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.RSSFeed{}, publication.ErrFeedForbidden
	}
	return item, err
}

func (s *PublicationStore) RevokeRSSFeed(ctx context.Context, feedID string, revokedAt time.Time) (publication.RSSFeed, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE rss_feeds
		SET status='revoked', revoked_at=$2
		WHERE id=$1::uuid
		RETURNING id::text, user_id::text, name, token_prefix, status, created_at, last_used_at, rotated_at, revoked_at, expires_at
	`, feedID, revokedAt)
	item, err := scanRSSFeed(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.RSSFeed{}, publication.ErrFeedNotFound
	}
	return item, err
}

func (s *PublicationStore) DeleteRSSFeed(ctx context.Context, feedID string, userID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM rss_feeds WHERE id=$1::uuid AND user_id=$2::uuid`, feedID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return publication.ErrFeedForbidden
	}
	return nil
}

func (s *PublicationStore) GetRSSFeedByTokenHash(ctx context.Context, tokenHash string, now time.Time) (publication.RSSFeed, auth.User, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT rf.id::text, rf.user_id::text, rf.name, rf.token_prefix, rf.status,
		       rf.created_at, rf.last_used_at, rf.rotated_at, rf.revoked_at, rf.expires_at,
		       u.id::text, u.email_normalized, u.display_name, u.role, u.status, u.created_at, u.updated_at, u.verified_at, u.deleted_at
		FROM rss_feeds rf
		JOIN users u ON u.id=rf.user_id
		WHERE rf.token_hash=$1
	`, tokenHash)
	var feed publication.RSSFeed
	var user auth.User
	var lastUsedAt, rotatedAt, revokedAt, expiresAt, verifiedAt, deletedAt *time.Time
	if err := row.Scan(&feed.ID, &feed.UserID, &feed.Name, &feed.TokenPrefix, &feed.Status, &feed.CreatedAt, &lastUsedAt, &rotatedAt, &revokedAt, &expiresAt, &user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt, &verifiedAt, &deletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return publication.RSSFeed{}, auth.User{}, false, nil
		}
		return publication.RSSFeed{}, auth.User{}, false, err
	}
	feed.LastUsedAt = lastUsedAt
	feed.RotatedAt = rotatedAt
	feed.RevokedAt = revokedAt
	feed.ExpiresAt = expiresAt
	user.VerifiedAt = verifiedAt
	user.DeletedAt = deletedAt
	if feed.Status == publication.FeedStatusActive && feed.ExpiresAt != nil && !feed.ExpiresAt.After(now) {
		feed.Status = publication.FeedStatusExpired
	}
	return feed, user, true, nil
}

func (s *PublicationStore) TouchRSSFeed(ctx context.Context, feedID string, usedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE rss_feeds SET last_used_at=$2 WHERE id=$1::uuid`, feedID, usedAt)
	return err
}

func (s *PublicationStore) GetAuthorizedMediaForUser(ctx context.Context, userID string, episodeID string) (publication.AuthorizedMedia, bool, error) {
	row := s.pool.QueryRow(ctx, mediaAccessSelect+`
		AND u.id=$1::uuid AND e.id=$2::uuid
		LIMIT 1
	`, userID, episodeID)
	item, err := scanAuthorizedMedia(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.AuthorizedMedia{}, false, nil
	}
	return item, err == nil, err
}

func (s *PublicationStore) GetAuthorizedMediaForFeed(ctx context.Context, tokenHash string, episodeID string, now time.Time) (publication.AuthorizedMedia, bool, error) {
	feed, user, found, err := s.GetRSSFeedByTokenHash(ctx, tokenHash, now)
	if err != nil || !found || feed.Status != publication.FeedStatusActive || user.Status != auth.StatusActive {
		return publication.AuthorizedMedia{}, false, err
	}
	row := s.pool.QueryRow(ctx, mediaAccessSelect+`
		AND u.id=$1::uuid AND e.id=$2::uuid
		LIMIT 1
	`, user.ID, episodeID)
	item, err := scanAuthorizedMedia(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.AuthorizedMedia{}, false, nil
	}
	if err != nil {
		return publication.AuthorizedMedia{}, false, err
	}
	item.FeedID = &feed.ID
	item.FeedOwnerUserID = &user.ID
	return item, true, nil
}

func (s *PublicationStore) ListAuthorizedFeedEpisodes(ctx context.Context, tokenHash string, now time.Time) (publication.RSSFeed, auth.User, []publication.FeedEpisode, error) {
	feed, user, found, err := s.GetRSSFeedByTokenHash(ctx, tokenHash, now)
	if err != nil {
		return publication.RSSFeed{}, auth.User{}, nil, err
	}
	if !found {
		return publication.RSSFeed{}, auth.User{}, nil, publication.ErrFeedTokenInvalid
	}
	if user.Status != auth.StatusActive {
		return publication.RSSFeed{}, auth.User{}, nil, publication.ErrFeedForbidden
	}
	rows, err := s.pool.Query(ctx, `
		SELECT e.id::text, e.program_id::text, p.title, p.author, p.language, e.title, e.description, e.published_at,
		       ma.content_type, ma.size_bytes, ma.sha256,
		       GREATEST(COALESCE(ma.published_at, e.published_at), COALESCE(e.updated_at, e.published_at), COALESCE(p.updated_at, p.published_at))
		FROM episodes e
		JOIN programs p ON p.id=e.program_id
		JOIN media_assets ma ON ma.owner_type='episode' AND ma.owner_id=e.id AND ma.media_kind='audio' AND ma.status='published' AND ma.delivery_status='published' AND ma.deleted_at IS NULL AND COALESCE(ma.published_storage_key,'') <> ''
		JOIN program_access_grants pag ON pag.program_id=p.id AND pag.user_id=$1::uuid AND pag.status='active' AND pag.revoked_at IS NULL
		WHERE e.status='published' AND p.status='published'
		ORDER BY e.published_at DESC, e.id ASC
	`, user.ID)
	if err != nil {
		return publication.RSSFeed{}, auth.User{}, nil, err
	}
	defer rows.Close()
	var items []publication.FeedEpisode
	for rows.Next() {
		var item publication.FeedEpisode
		if err := rows.Scan(&item.EpisodeID, &item.ProgramID, &item.ProgramTitle, &item.ProgramAuthor, &item.Language, &item.Title, &item.Description, &item.PublishedAt, &item.ContentType, &item.SizeBytes, &item.SHA256, &item.UpdatedAt); err != nil {
			return publication.RSSFeed{}, auth.User{}, nil, err
		}
		items = append(items, item)
	}
	return feed, user, items, rows.Err()
}

const mediaAccessSelect = `
	SELECT e.id::text, e.program_id::text, ma.content_type, ma.size_bytes, ma.sha256,
	       COALESCE(ma.published_at, e.published_at), ma.published_storage_key, e.title, p.title
	FROM episodes e
	JOIN programs p ON p.id=e.program_id
	JOIN media_assets ma ON ma.owner_type='episode' AND ma.owner_id=e.id AND ma.media_kind='audio' AND ma.status='published' AND ma.delivery_status='published' AND ma.deleted_at IS NULL AND COALESCE(ma.published_storage_key,'') <> ''
	JOIN users u ON u.id=$1::uuid AND u.status='active'
	JOIN program_access_grants pag ON pag.program_id=p.id AND pag.user_id=u.id AND pag.status='active' AND pag.revoked_at IS NULL
	WHERE e.status='published' AND p.status='published'
`

func scanProgramAccessGrant(row interface{ Scan(dest ...any) error }) (publication.ProgramAccessGrant, error) {
	var item publication.ProgramAccessGrant
	err := row.Scan(&item.ID, &item.UserID, &item.ProgramID, &item.Status, &item.Reason, &item.CreatedAt, &item.UpdatedAt, &item.RevokedAt)
	return item, err
}

func scanRSSFeed(row interface{ Scan(dest ...any) error }) (publication.RSSFeed, error) {
	var item publication.RSSFeed
	err := row.Scan(&item.ID, &item.UserID, &item.Name, &item.TokenPrefix, &item.Status, &item.CreatedAt, &item.LastUsedAt, &item.RotatedAt, &item.RevokedAt, &item.ExpiresAt)
	return item, err
}

func scanAdminRSSFeed(row interface{ Scan(dest ...any) error }) (publication.RSSFeed, string, error) {
	var item publication.RSSFeed
	var email string
	err := row.Scan(&item.ID, &item.UserID, &item.Name, &item.TokenPrefix, &item.Status, &item.CreatedAt, &item.LastUsedAt, &item.RotatedAt, &item.RevokedAt, &item.ExpiresAt, &email)
	if err != nil {
		return publication.RSSFeed{}, "", err
	}
	hint := email
	if len(email) > 2 {
		hint = email[:1] + "***"
	}
	return item, hint, nil
}

func scanAuthorizedMedia(row interface{ Scan(dest ...any) error }) (publication.AuthorizedMedia, error) {
	var item publication.AuthorizedMedia
	err := row.Scan(&item.EpisodeID, &item.ProgramID, &item.ContentType, &item.SizeBytes, &item.SHA256, &item.PublishedAt, &item.PublishedKey, &item.EpisodeTitle, &item.ProgramTitle)
	return item, err
}
