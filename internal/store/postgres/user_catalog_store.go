package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/mdxz2048/podcast-hub/internal/publication"
)

func (s *PublicationStore) ListAuthorizedPrograms(ctx context.Context, userID string) ([]publication.UserProgram, error) {
	rows, err := s.pool.Query(ctx, authorizedProgramSelect+`
		WHERE p.status='published' AND pag.user_id=$1::uuid
	`+authorizedProgramGroup+`
		ORDER BY p.updated_at DESC, p.title ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.UserProgram
	for rows.Next() {
		item, err := scanUserProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PublicationStore) GetAuthorizedProgram(ctx context.Context, userID string, programID string) (publication.UserProgram, bool, error) {
	row := s.pool.QueryRow(ctx, authorizedProgramSelect+`
		WHERE p.status='published' AND pag.user_id=$1::uuid AND p.id=$2::uuid
	`+authorizedProgramGroup+`
		LIMIT 1
	`, userID, programID)
	item, err := scanUserProgram(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.UserProgram{}, false, nil
	}
	return item, err == nil, err
}

func (s *PublicationStore) ListAuthorizedEpisodes(ctx context.Context, userID string, programID string) ([]publication.UserEpisode, error) {
	rows, err := s.pool.Query(ctx, authorizedEpisodeSelect+`
		AND pag.user_id=$1::uuid AND p.id=$2::uuid
		ORDER BY e.published_at DESC, e.id ASC
	`, userID, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.UserEpisode
	for rows.Next() {
		item, err := scanUserEpisode(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PublicationStore) GetAuthorizedEpisode(ctx context.Context, userID string, episodeID string) (publication.UserEpisode, bool, error) {
	row := s.pool.QueryRow(ctx, authorizedEpisodeSelect+`
		AND pag.user_id=$1::uuid AND e.id=$2::uuid
		LIMIT 1
	`, userID, episodeID)
	item, err := scanUserEpisode(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.UserEpisode{}, false, nil
	}
	return item, err == nil, err
}

func (s *PublicationStore) ListUserCollections(ctx context.Context, userID string) ([]publication.UserCollection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, user_id::text, title, description, created_at, updated_at
		FROM user_collections
		WHERE user_id=$1::uuid
		ORDER BY updated_at DESC, created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.UserCollection
	for rows.Next() {
		item, err := scanUserCollection(rows)
		if err != nil {
			return nil, err
		}
		programs, err := s.listVisibleCollectionPrograms(ctx, userID, item.ID)
		if err != nil {
			return nil, err
		}
		item.Programs = programs
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PublicationStore) CreateUserCollection(ctx context.Context, collection publication.UserCollection) (publication.UserCollection, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO user_collections(id, user_id, title, description, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $5)
		RETURNING id::text, user_id::text, title, description, created_at, updated_at
	`, collection.ID, collection.UserID, collection.Title, collection.Description, collection.CreatedAt)
	return scanUserCollection(row)
}

func (s *PublicationStore) UpdateUserCollection(ctx context.Context, userID string, collectionID string, title *string, description *string, updatedAt time.Time) (publication.UserCollection, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE user_collections
		SET title=COALESCE($3, title), description=COALESCE($4, description), updated_at=$5
		WHERE id=$1::uuid AND user_id=$2::uuid
		RETURNING id::text, user_id::text, title, description, created_at, updated_at
	`, collectionID, userID, title, description, updatedAt)
	item, err := scanUserCollection(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return publication.UserCollection{}, publication.ErrCollectionNotFound
	}
	if err != nil {
		return publication.UserCollection{}, err
	}
	item.Programs, err = s.listVisibleCollectionPrograms(ctx, userID, item.ID)
	return item, err
}

func (s *PublicationStore) DeleteUserCollection(ctx context.Context, userID string, collectionID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM user_collections WHERE id=$1::uuid AND user_id=$2::uuid`, collectionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return publication.ErrCollectionNotFound
	}
	return nil
}

func (s *PublicationStore) AddProgramToCollection(ctx context.Context, userID string, collectionID string, programID string, addedAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO user_collection_programs(collection_id, program_id, added_at)
		SELECT uc.id, $3::uuid, $4
		FROM user_collections uc
		WHERE uc.id=$1::uuid AND uc.user_id=$2::uuid
		ON CONFLICT(collection_id, program_id) DO NOTHING
	`, collectionID, userID, programID, addedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_collections WHERE id=$1::uuid AND user_id=$2::uuid)`, collectionID, userID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return publication.ErrCollectionNotFound
		}
	}
	_, err = s.pool.Exec(ctx, `UPDATE user_collections SET updated_at=$3 WHERE id=$1::uuid AND user_id=$2::uuid`, collectionID, userID, addedAt)
	return err
}

func (s *PublicationStore) RemoveProgramFromCollection(ctx context.Context, userID string, collectionID string, programID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM user_collection_programs ucp
		USING user_collections uc
		WHERE ucp.collection_id=uc.id AND uc.id=$1::uuid AND uc.user_id=$2::uuid AND ucp.program_id=$3::uuid
	`, collectionID, userID, programID)
	if err != nil {
		return err
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_collections WHERE id=$1::uuid AND user_id=$2::uuid)`, collectionID, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return publication.ErrCollectionNotFound
	}
	if tag.RowsAffected() > 0 {
		_, err = s.pool.Exec(ctx, `UPDATE user_collections SET updated_at=NOW() WHERE id=$1::uuid AND user_id=$2::uuid`, collectionID, userID)
	}
	return err
}

func (s *PublicationStore) listVisibleCollectionPrograms(ctx context.Context, userID string, collectionID string) ([]publication.UserProgram, error) {
	rows, err := s.pool.Query(ctx, authorizedProgramSelect+`
		JOIN user_collection_programs ucp ON ucp.program_id=p.id AND ucp.collection_id=$2::uuid
		WHERE p.status='published' AND pag.user_id=$1::uuid
	`+authorizedProgramGroup+`
		ORDER BY ucp.added_at DESC, p.title ASC
	`, userID, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []publication.UserProgram
	for rows.Next() {
		item, err := scanUserProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

const authorizedProgramSelect = `
	SELECT p.id::text, p.title, p.description, p.author, p.language, p.status,
	       COUNT(DISTINCT CASE WHEN ma.id IS NOT NULL THEN e.id END)::int AS episode_count, p.updated_at
	FROM programs p
	JOIN program_access_grants pag ON pag.program_id=p.id AND pag.status='active' AND pag.revoked_at IS NULL
	LEFT JOIN episodes e ON e.program_id=p.id AND e.status='published'
	LEFT JOIN media_assets ma ON ma.owner_type='episode' AND ma.owner_id=e.id AND ma.media_kind='audio' AND ma.status='published' AND ma.delivery_status='published' AND ma.deleted_at IS NULL AND COALESCE(ma.published_storage_key,'') <> ''
`

const authorizedProgramGroup = `
	GROUP BY p.id, p.title, p.description, p.author, p.language, p.status, p.updated_at
`

const authorizedEpisodeSelect = `
	SELECT e.id::text, e.program_id::text, e.title, e.description, e.published_at, e.duration_seconds, e.status, ma.delivery_status
	FROM episodes e
	JOIN programs p ON p.id=e.program_id
	JOIN program_access_grants pag ON pag.program_id=p.id AND pag.status='active' AND pag.revoked_at IS NULL
	JOIN media_assets ma ON ma.owner_type='episode' AND ma.owner_id=e.id AND ma.media_kind='audio' AND ma.status='published' AND ma.delivery_status='published' AND ma.deleted_at IS NULL AND COALESCE(ma.published_storage_key,'') <> ''
	WHERE e.status='published' AND p.status='published'
`

func scanUserProgram(row interface{ Scan(dest ...any) error }) (publication.UserProgram, error) {
	var item publication.UserProgram
	err := row.Scan(&item.ID, &item.Title, &item.Description, &item.Author, &item.Language, &item.Status, &item.EpisodeCount, &item.UpdatedAt)
	return item, err
}

func scanUserEpisode(row interface{ Scan(dest ...any) error }) (publication.UserEpisode, error) {
	var item publication.UserEpisode
	err := row.Scan(&item.ID, &item.ProgramID, &item.Title, &item.Description, &item.PublishedAt, &item.DurationSeconds, &item.Status, &item.MediaStatus)
	return item, err
}

func scanUserCollection(row interface{ Scan(dest ...any) error }) (publication.UserCollection, error) {
	var item publication.UserCollection
	err := row.Scan(&item.ID, &item.UserID, &item.Title, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
