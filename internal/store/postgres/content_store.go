package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/content"
)

type ContentStore struct{ pool *pgxpool.Pool }

func NewContentStore(pool *pgxpool.Pool) *ContentStore { return &ContentStore{pool: pool} }

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
		RETURNING id::text, target_type, target_id::text, review_kind, status, requested_by_job_id::text, review_note, created_at
	`, id, targetType, targetID, reviewKind, jobID)
	var item content.ReviewItem
	err := row.Scan(&item.ID, &item.TargetType, &item.TargetID, &item.ReviewKind, &item.Status, &item.RequestedByJobID, &item.ReviewNote, &item.CreatedAt)
	return item, err
}

func (s *ContentStore) CreateMediaAsset(ctx context.Context, in content.CreateMediaAssetInput) (content.MediaAsset, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO media_assets(id, owner_type, owner_id, import_job_id, artifact_id, media_kind, staged_storage_key, content_type, size_bytes, sha256, status, created_at)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5::uuid, $6, $7, $8, $9, $10, 'staged', NOW())
		RETURNING id::text, owner_type, owner_id::text, import_job_id::text, artifact_id::text, media_kind, staged_storage_key, content_type, size_bytes, sha256, status, created_at
	`, uuid.NewString(), in.OwnerType, in.OwnerID, in.ImportJobID, in.ArtifactID, in.MediaKind, in.StagedStorageKey, in.ContentType, in.SizeBytes, in.SHA256)
	var item content.MediaAsset
	err := row.Scan(&item.ID, &item.OwnerType, &item.OwnerID, &item.ImportJobID, &item.ArtifactID, &item.MediaKind, &item.StagedStorageKey, &item.ContentType, &item.SizeBytes, &item.SHA256, &item.Status, &item.CreatedAt)
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

func emptyToNilString(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}
