package postgres

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/content"
	"github.com/mdxz2048/podcast-hub/internal/media"
)

type ContentStore struct {
	pool       *pgxpool.Pool
	mediaStore *media.LocalStore
}

func NewContentStore(pool *pgxpool.Pool, mediaStore *media.LocalStore) *ContentStore {
	return &ContentStore{pool: pool, mediaStore: mediaStore}
}

func (s *ContentStore) UpsertProgramFromSource(ctx context.Context, in content.UpsertProgramInput) (content.Program, error) {
	now := time.Now()
	canonical := in.SourceID + ":" + in.ExternalProgramID
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return content.Program{}, err
	}
	defer tx.Rollback(ctx)
	var programID string
	err = tx.QueryRow(ctx, `SELECT program_id::text FROM program_sources WHERE connector_source_id=$1::uuid AND external_program_id=$2`, in.SourceID, in.ExternalProgramID).Scan(&programID)
	if err != nil && err != pgx.ErrNoRows {
		return content.Program{}, err
	}
	if err == pgx.ErrNoRows {
		programID = uuid.NewString()
		if _, err := tx.Exec(ctx, `
			INSERT INTO programs(id, canonical_key, title, description, author, language, status, created_from_source_id, created_from_job_id, created_at, updated_at)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, 'review_pending', $7::uuid, $8::uuid, $9, $9)
		`, programID, canonical, in.Title, in.Description, in.Author, in.Language, in.SourceID, in.JobID, now); err != nil {
			return content.Program{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO program_sources(id, program_id, connector_source_id, external_program_id, first_import_job_id, last_import_job_id, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, $5::uuid, $6, $6)
		`, uuid.NewString(), programID, in.SourceID, in.ExternalProgramID, in.JobID, now); err != nil {
			return content.Program{}, err
		}
	} else {
		if _, err := tx.Exec(ctx, `
			UPDATE programs SET title=$1, description=$2, author=$3, language=$4, status='review_pending', updated_at=$5 WHERE id=$6::uuid
		`, in.Title, in.Description, in.Author, in.Language, now, programID); err != nil {
			return content.Program{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE program_sources SET last_import_job_id=$1::uuid, updated_at=$2 WHERE program_id=$3::uuid AND connector_source_id=$4::uuid AND external_program_id=$5`, in.JobID, now, programID, in.SourceID, in.ExternalProgramID); err != nil {
			return content.Program{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return content.Program{}, err
	}
	item, found, err := s.GetProgram(ctx, programID)
	if err != nil || !found {
		return content.Program{}, err
	}
	return item, nil
}

func (s *ContentStore) UpsertEpisode(ctx context.Context, in content.UpsertEpisodeInput) (content.Episode, error) {
	id := uuid.NewString()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO episodes(id, program_id, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, 'review_pending', $8::uuid, NOW(), NOW())
		ON CONFLICT (program_id, external_episode_id)
		DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description, published_at=EXCLUDED.published_at, duration_seconds=EXCLUDED.duration_seconds, status='review_pending', source_job_id=EXCLUDED.source_job_id, updated_at=NOW()
		RETURNING id::text, program_id::text, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id::text, created_at, updated_at
	`, id, in.ProgramID, in.ExternalEpisodeID, in.Title, in.Description, in.PublishedAt, in.DurationSeconds, in.SourceJobID)
	return scanEpisode(row)
}

func (s *ContentStore) CreateOrKeepPendingReview(ctx context.Context, targetType string, targetID string, reviewKind string, jobID string) (content.ReviewItem, error) {
	id := uuid.NewString()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO review_items(id, target_type, target_id, review_kind, status, requested_by_job_id, created_at)
		VALUES ($1::uuid, $2, $3::uuid, $4, 'pending', $5::uuid, NOW())
		ON CONFLICT (target_type, target_id, review_kind) WHERE status='pending'
		DO UPDATE SET requested_by_job_id=EXCLUDED.requested_by_job_id
		RETURNING id::text, target_type, target_id::text, review_kind, status, requested_by_job_id::text, reviewed_by::text, review_note, created_at, reviewed_at
	`, id, targetType, targetID, reviewKind, emptyToNil(jobID))
	return scanReview(row)
}

func (s *ContentStore) CreateMediaAsset(ctx context.Context, in content.CreateMediaAssetInput) (content.MediaAsset, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO media_assets(id, owner_type, owner_id, import_job_id, artifact_id, media_kind, staged_storage_key, content_type, size_bytes, sha256, status, delivery_status, created_at)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5::uuid, $6, $7, $8, $9, $10, 'staged', 'staged', NOW())
		RETURNING id::text, owner_type, owner_id::text, import_job_id::text, artifact_id::text, media_kind, staged_storage_key, content_type, size_bytes, sha256, status, delivery_status, created_at, published_storage_key, published_at, deleted_at
	`, uuid.NewString(), in.OwnerType, in.OwnerID, in.ImportJobID, in.ArtifactID, in.MediaKind, in.StagedStorageKey, in.ContentType, in.SizeBytes, in.SHA256)
	var item content.MediaAsset
	err := row.Scan(&item.ID, &item.OwnerType, &item.OwnerID, &item.ImportJobID, &item.ArtifactID, &item.MediaKind, &item.StagedStorageKey, &item.ContentType, &item.SizeBytes, &item.SHA256, &item.Status, &item.DeliveryStatus, &item.CreatedAt, &item.PublishedKey, &item.PublishedAt, &item.DeletedAt)
	return item, err
}

func (s *ContentStore) InsertPublicationEvent(ctx context.Context, event content.PublicationEvent) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO publication_events(id, target_type, target_id, event_type, actor_id, metadata_redacted, created_at) VALUES ($1::uuid, $2, $3::uuid, $4, $5::uuid, $6::jsonb, $7)`, event.ID, event.TargetType, event.TargetID, event.EventType, event.ActorID, defaultJSON(event.MetadataRedacted), event.CreatedAt)
	return err
}

func (s *ContentStore) GetIntakeRun(ctx context.Context, jobID string) (content.IntakeRun, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, import_job_id::text, status, validation_issues_redacted::text, program_id::text, created_at, updated_at FROM intake_runs WHERE import_job_id=$1::uuid`, jobID)
	item, err := scanIntakeRun(row)
	if err == pgx.ErrNoRows {
		return content.IntakeRun{}, false, nil
	}
	return item, err == nil, err
}

func (s *ContentStore) UpsertIntakeRun(ctx context.Context, run content.IntakeRun) (content.IntakeRun, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO intake_runs(id, import_job_id, status, validation_issues_redacted, program_id, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4::jsonb, $5::uuid, $6, $7)
		ON CONFLICT (import_job_id)
		DO UPDATE SET status=EXCLUDED.status, validation_issues_redacted=EXCLUDED.validation_issues_redacted, program_id=EXCLUDED.program_id, updated_at=EXCLUDED.updated_at
		RETURNING id::text, import_job_id::text, status, validation_issues_redacted::text, program_id::text, created_at, updated_at
	`, run.ID, run.ImportJobID, run.Status, defaultJSON(run.ValidationIssuesRedacted), emptyToNilString(run.ProgramID), run.CreatedAt, run.UpdatedAt)
	item, err := scanIntakeRun(row)
	return item, err
}

func (s *ContentStore) ListStagingPrograms(ctx context.Context) ([]content.Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, canonical_key, title, description, author, language, status, created_from_source_id::text, created_from_job_id::text, created_at, updated_at FROM programs WHERE status IN ('staging','review_pending') ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []content.Program
	for rows.Next() {
		item, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ContentStore) GetProgram(ctx context.Context, programID string) (content.Program, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, canonical_key, title, description, author, language, status, created_from_source_id::text, created_from_job_id::text, created_at, updated_at FROM programs WHERE id=$1::uuid`, programID)
	item, err := scanProgram(row)
	if err == pgx.ErrNoRows {
		return content.Program{}, false, nil
	}
	return item, err == nil, err
}

func (s *ContentStore) ListStagingEpisodes(ctx context.Context) ([]content.Episode, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, program_id::text, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id::text, created_at, updated_at FROM episodes WHERE status IN ('staging','review_pending') ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []content.Episode
	for rows.Next() {
		item, err := scanEpisode(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ContentStore) GetEpisode(ctx context.Context, episodeID string) (content.Episode, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, program_id::text, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id::text, created_at, updated_at FROM episodes WHERE id=$1::uuid`, episodeID)
	item, err := scanEpisode(row)
	if err == pgx.ErrNoRows {
		return content.Episode{}, false, nil
	}
	return item, err == nil, err
}

func (s *ContentStore) ListAdminPrograms(ctx context.Context) ([]content.Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, canonical_key, title, description, author, language, status, created_from_source_id::text, created_from_job_id::text, created_at, updated_at FROM programs ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []content.Program
	for rows.Next() {
		item, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *ContentStore) GetAdminProgram(ctx context.Context, programID string) (content.Program, bool, error) {
	return s.GetProgram(ctx, programID)
}

func (s *ContentStore) GetAdminEpisode(ctx context.Context, episodeID string) (content.Episode, bool, error) {
	return s.GetEpisode(ctx, episodeID)
}

func (s *ContentStore) ListProgramEpisodes(ctx context.Context, programID string) ([]content.Episode, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, program_id::text, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id::text, created_at, updated_at FROM episodes WHERE program_id=$1::uuid ORDER BY published_at DESC`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []content.Episode
	for rows.Next() {
		item, err := scanEpisode(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *ContentStore) ListReviews(ctx context.Context) ([]content.ReviewItem, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, target_type, target_id::text, review_kind, status, requested_by_job_id::text, reviewed_by::text, review_note, created_at, reviewed_at FROM review_items ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []content.ReviewItem
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *ContentStore) GetReview(ctx context.Context, reviewID string) (content.ReviewItem, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, target_type, target_id::text, review_kind, status, requested_by_job_id::text, reviewed_by::text, review_note, created_at, reviewed_at FROM review_items WHERE id=$1::uuid`, reviewID)
	item, err := scanReview(row)
	if err == pgx.ErrNoRows {
		return content.ReviewItem{}, false, nil
	}
	return item, err == nil, err
}

func (s *ContentStore) SetReviewDecision(ctx context.Context, reviewID string, status content.ReviewStatus, actorID string, note string) (content.ReviewItem, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE review_items
		SET status=$2, reviewed_by=$3::uuid, review_note=$4, reviewed_at=NOW()
		WHERE id=$1::uuid AND status='pending'
		RETURNING id::text, target_type, target_id::text, review_kind, status, requested_by_job_id::text, reviewed_by::text, review_note, created_at, reviewed_at
	`, reviewID, status, actorID, note)
	return scanReview(row)
}

func (s *ContentStore) SetProgramStatus(ctx context.Context, programID string, status content.ProgramStatus) (content.Program, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE programs
		SET status=$2, updated_at=NOW(),
		    published_at=CASE WHEN $2='published' THEN NOW() ELSE published_at END,
		    archived_at=CASE WHEN $2='archived' THEN NOW() ELSE archived_at END
		WHERE id=$1::uuid
		RETURNING id::text, canonical_key, title, description, author, language, status, created_from_source_id::text, created_from_job_id::text, created_at, updated_at
	`, programID, status)
	return scanProgram(row)
}

func (s *ContentStore) SetEpisodeStatus(ctx context.Context, episodeID string, status content.EpisodeStatus) (content.Episode, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE episodes
		SET status=$2, updated_at=NOW(),
		    published_to_users_at=CASE WHEN $2='published' THEN NOW() ELSE published_to_users_at END,
		    archived_at=CASE WHEN $2='archived' THEN NOW() ELSE archived_at END
		WHERE id=$1::uuid
		RETURNING id::text, program_id::text, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id::text, created_at, updated_at
	`, episodeID, status)
	return scanEpisode(row)
}

func (s *ContentStore) UpdateProgram(ctx context.Context, programID string, in content.UpdateProgramInput) (content.Program, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE programs
		SET title=COALESCE($2, title), description=COALESCE($3, description), author=COALESCE($4, author), language=COALESCE($5, language), updated_at=NOW()
		WHERE id=$1::uuid
		RETURNING id::text, canonical_key, title, description, author, language, status, created_from_source_id::text, created_from_job_id::text, created_at, updated_at
	`, programID, in.Title, in.Description, in.Author, in.Language)
	return scanProgram(row)
}

func (s *ContentStore) UpdateEpisode(ctx context.Context, episodeID string, in content.UpdateEpisodeInput) (content.Episode, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE episodes
		SET title=COALESCE($2, title), description=COALESCE($3, description), published_at=COALESCE($4, published_at), duration_seconds=COALESCE($5, duration_seconds), updated_at=NOW()
		WHERE id=$1::uuid
		RETURNING id::text, program_id::text, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id::text, created_at, updated_at
	`, episodeID, in.Title, in.Description, in.PublishedAt, in.DurationSeconds)
	return scanEpisode(row)
}

func (s *ContentStore) CountPendingReviews(ctx context.Context, targetType string, targetID string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM review_items WHERE target_type=$1 AND target_id=$2::uuid AND status='pending'`, targetType, targetID).Scan(&count)
	return count, err
}

func (s *ContentStore) HasApprovedMedia(ctx context.Context, episodeID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM media_assets WHERE owner_type='episode' AND owner_id=$1::uuid AND media_kind='audio' AND status='approved'
		) AND NOT EXISTS (
			SELECT 1 FROM media_assets WHERE owner_type='episode' AND owner_id=$1::uuid AND status <> 'approved'
		)
	`, episodeID).Scan(&ok)
	return ok, err
}

func (s *ContentStore) ApproveMediaForEpisode(ctx context.Context, episodeID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE media_assets SET status='approved', delivery_status='approved' WHERE owner_type='episode' AND owner_id=$1::uuid AND status='staged'`, episodeID)
	return err
}

func (s *ContentStore) PromoteEpisodeMedia(ctx context.Context, episodeID string) error {
	if s.mediaStore == nil {
		return fmt.Errorf("media store is not configured")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	type assetRow struct {
		id        string
		stagedKey string
		sha256    string
	}
	rows, err := tx.Query(ctx, `
		SELECT id::text, staged_storage_key, sha256
		FROM media_assets
		WHERE owner_type='episode' AND owner_id=$1::uuid AND status='approved' AND delivery_status='approved'
		ORDER BY created_at ASC
	`, episodeID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var assets []assetRow
	for rows.Next() {
		var item assetRow
		if err := rows.Scan(&item.id, &item.stagedKey, &item.sha256); err != nil {
			return err
		}
		assets = append(assets, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(assets) == 0 {
		return fmt.Errorf("approved media not found")
	}
	var promoted []string
	now := time.Now()
	for _, item := range assets {
		targetKey := filepath.ToSlash(filepath.Join("episodes", episodeID, item.id+"-"+shortHash(item.sha256)+".bin"))
		if err := s.mediaStore.Promote(ctx, item.stagedKey, targetKey); err != nil {
			for _, key := range promoted {
				_ = s.mediaStore.DeletePublished(ctx, key)
			}
			return err
		}
		promoted = append(promoted, targetKey)
		if _, err := tx.Exec(ctx, `
			UPDATE media_assets
			SET status='published', delivery_status='published', published_storage_key=$2, published_at=$3
			WHERE id=$1::uuid
		`, item.id, targetKey, now); err != nil {
			for _, key := range promoted {
				_ = s.mediaStore.DeletePublished(ctx, key)
			}
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		for _, key := range promoted {
			_ = s.mediaStore.DeletePublished(ctx, key)
		}
		return err
	}
	return nil
}

func shortHash(value string) string {
	if len(value) >= 12 {
		return value[:12]
	}
	return value
}

func scanProgram(row interface{ Scan(dest ...any) error }) (content.Program, error) {
	var item content.Program
	err := row.Scan(&item.ID, &item.CanonicalKey, &item.Title, &item.Description, &item.Author, &item.Language, &item.Status, &item.CreatedFromSourceID, &item.CreatedFromJobID, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func scanEpisode(row interface{ Scan(dest ...any) error }) (content.Episode, error) {
	var item content.Episode
	err := row.Scan(&item.ID, &item.ProgramID, &item.ExternalEpisodeID, &item.Title, &item.Description, &item.PublishedAt, &item.DurationSeconds, &item.Status, &item.SourceJobID, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func scanIntakeRun(row interface{ Scan(dest ...any) error }) (content.IntakeRun, error) {
	var item content.IntakeRun
	var programID *string
	err := row.Scan(&item.ID, &item.ImportJobID, &item.Status, &item.ValidationIssuesRedacted, &programID, &item.CreatedAt, &item.UpdatedAt)
	item.ProgramID = programID
	return item, err
}

func scanReview(row interface{ Scan(dest ...any) error }) (content.ReviewItem, error) {
	var item content.ReviewItem
	err := row.Scan(&item.ID, &item.TargetType, &item.TargetID, &item.ReviewKind, &item.Status, &item.RequestedByJobID, &item.ReviewedBy, &item.ReviewNote, &item.CreatedAt, &item.ReviewedAt)
	return item, err
}

func emptyToNilString(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}
