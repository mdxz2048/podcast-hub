package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

type JobStore struct{ pool *pgxpool.Pool }

func NewJobStore(pool *pgxpool.Pool) *JobStore { return &JobStore{pool: pool} }

func (s *JobStore) ListJobs(ctx context.Context) ([]jobs.ImportJob, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, connector_source_id::text, connector_version_id::text, status, requested_by::text, trigger_type, auth_mode, execution_mode, started_at, finished_at, timeout_at, cancellation_requested_at, failure_code, failure_message_redacted, created_at, updated_at FROM import_jobs ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()
	items := []jobs.ImportJob{}
	for rows.Next() {
		item, scanErr := scanJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *JobStore) GetJob(ctx context.Context, jobID string) (jobs.ImportJob, bool, error) {
	row := s.pool.QueryRow(ctx, `SELECT id::text, connector_source_id::text, connector_version_id::text, status, requested_by::text, trigger_type, auth_mode, execution_mode, started_at, finished_at, timeout_at, cancellation_requested_at, failure_code, failure_message_redacted, created_at, updated_at FROM import_jobs WHERE id=$1::uuid`, jobID)
	item, err := scanJob(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return jobs.ImportJob{}, false, nil
		}
		return jobs.ImportJob{}, false, fmt.Errorf("get job: %w", err)
	}
	return item, true, nil
}

func (s *JobStore) HasActiveJobForSource(ctx context.Context, sourceID string) (bool, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM import_jobs WHERE connector_source_id=$1::uuid AND status IN ('queued', 'running'))`, sourceID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active job for source: %w", err)
	}
	return exists, nil
}

func (s *JobStore) CreateJob(ctx context.Context, job jobs.ImportJob) (jobs.ImportJob, error) {
	var requestedBy any
	if job.RequestedBy != nil && *job.RequestedBy != "" {
		requestedBy = *job.RequestedBy
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO import_jobs(id, connector_source_id, connector_version_id, status, requested_by, trigger_type, auth_mode, execution_mode, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, $6, $7, $8, $9, $10)
		RETURNING id::text, connector_source_id::text, connector_version_id::text, status, requested_by::text, trigger_type, auth_mode, execution_mode, started_at, finished_at, timeout_at, cancellation_requested_at, failure_code, failure_message_redacted, created_at, updated_at
	`, job.ID, job.ConnectorSourceID, job.ConnectorVersionID, job.Status, requestedBy, job.TriggerType, job.AuthMode, job.ExecutionMode, job.CreatedAt, job.UpdatedAt)
	created, err := scanJob(row)
	if err != nil {
		return jobs.ImportJob{}, fmt.Errorf("create job: %w", err)
	}
	return created, nil
}

func (s *JobStore) UpdateJobStatus(ctx context.Context, jobID string, in jobs.UpdateJobStatusInput) (jobs.ImportJob, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE import_jobs
		SET status=$1, started_at=COALESCE($2, started_at), finished_at=COALESCE($3, finished_at), cancellation_requested_at=COALESCE($4, cancellation_requested_at), failure_code=$5, failure_message_redacted=$6, updated_at=$7
		WHERE id=$8::uuid
		RETURNING id::text, connector_source_id::text, connector_version_id::text, status, requested_by::text, trigger_type, auth_mode, execution_mode, started_at, finished_at, timeout_at, cancellation_requested_at, failure_code, failure_message_redacted, created_at, updated_at
	`, in.Status, in.StartedAt, in.FinishedAt, in.CancellationRequestedAt, emptyToNil(in.FailureCode), emptyToNil(in.FailureMessageRedacted), time.Now(), jobID)
	updated, err := scanJob(row)
	if err != nil {
		return jobs.ImportJob{}, fmt.Errorf("update job: %w", err)
	}
	return updated, nil
}

func (s *JobStore) ListJobEvents(ctx context.Context, jobID string) ([]jobs.ImportJobEvent, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, import_job_id::text, event_type, level, message_redacted, metadata_redacted::text, created_at FROM import_job_events WHERE import_job_id=$1::uuid ORDER BY created_at ASC`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list job events: %w", err)
	}
	defer rows.Close()
	items := []jobs.ImportJobEvent{}
	for rows.Next() {
		var item jobs.ImportJobEvent
		if err := rows.Scan(&item.ID, &item.ImportJobID, &item.EventType, &item.Level, &item.MessageRedacted, &item.MetadataRedacted, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *JobStore) InsertJobEvent(ctx context.Context, event jobs.ImportJobEvent) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO import_job_events(id, import_job_id, event_type, level, message_redacted, metadata_redacted, created_at) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6::jsonb, $7)`, event.ID, event.ImportJobID, event.EventType, event.Level, event.MessageRedacted, defaultJSON(event.MetadataRedacted), event.CreatedAt)
	return err
}

func (s *JobStore) ListJobArtifacts(ctx context.Context, jobID string) ([]jobs.ImportJobArtifact, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, import_job_id::text, artifact_type, relative_path, size_bytes, sha256, created_at FROM import_job_artifacts WHERE import_job_id=$1::uuid ORDER BY created_at ASC`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list job artifacts: %w", err)
	}
	defer rows.Close()
	items := []jobs.ImportJobArtifact{}
	for rows.Next() {
		var item jobs.ImportJobArtifact
		if err := rows.Scan(&item.ID, &item.ImportJobID, &item.ArtifactType, &item.RelativePath, &item.SizeBytes, &item.SHA256, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func scanJob(row interface{ Scan(dest ...any) error }) (jobs.ImportJob, error) {
	var item jobs.ImportJob
	var requestedBy *string
	err := row.Scan(&item.ID, &item.ConnectorSourceID, &item.ConnectorVersionID, &item.Status, &requestedBy, &item.TriggerType, &item.AuthMode, &item.ExecutionMode, &item.StartedAt, &item.FinishedAt, &item.TimeoutAt, &item.CancellationRequestedAt, &item.FailureCode, &item.FailureMessageRedacted, &item.CreatedAt, &item.UpdatedAt)
	item.RequestedBy = requestedBy
	return item, err
}

func emptyToNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}
