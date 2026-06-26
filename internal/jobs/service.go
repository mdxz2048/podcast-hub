package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mdxz2048/podcast-hub/internal/sources"
)

type Service struct {
	store     Store
	sourceSvc *sources.Service
}

func NewService(store Store, sourceSvc *sources.Service) *Service {
	return &Service{store: store, sourceSvc: sourceSvc}
}

func (s *Service) ListJobs(ctx context.Context) ([]ImportJob, error) {
	return s.store.ListJobs(ctx)
}

func (s *Service) GetJob(ctx context.Context, jobID string) (ImportJob, error) {
	job, found, err := s.store.GetJob(ctx, jobID)
	if err != nil {
		return ImportJob{}, fmt.Errorf("get job: %w", err)
	}
	if !found {
		return ImportJob{}, ErrJobNotFound
	}
	return job, nil
}

func (s *Service) ListEvents(ctx context.Context, jobID string) ([]ImportJobEvent, error) {
	if _, err := s.GetJob(ctx, jobID); err != nil {
		return nil, err
	}
	return s.store.ListJobEvents(ctx, jobID)
}

func (s *Service) ListArtifacts(ctx context.Context, jobID string) ([]ImportJobArtifact, error) {
	if _, err := s.GetJob(ctx, jobID); err != nil {
		return nil, err
	}
	return s.store.ListJobArtifacts(ctx, jobID)
}

func (s *Service) ClaimNextQueuedJob(ctx context.Context) (ImportJob, bool, error) {
	job, found, err := s.store.ClaimNextQueuedJob(ctx)
	if err != nil {
		return ImportJob{}, false, fmt.Errorf("claim queued job: %w", err)
	}
	if !found {
		return ImportJob{}, false, nil
	}
	now := time.Now()
	_ = s.store.InsertJobEvent(ctx, ImportJobEvent{ID: uuid.NewString(), ImportJobID: job.ID, EventType: "runner.claimed", Level: "info", MessageRedacted: "Runner claimed import job.", MetadataRedacted: `{}`, CreatedAt: now})
	return job, true, nil
}

func (s *Service) AppendEvent(ctx context.Context, event ImportJobEvent) error {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.MetadataRedacted == "" {
		event.MetadataRedacted = `{}`
	}
	return s.store.InsertJobEvent(ctx, event)
}

func (s *Service) AppendArtifact(ctx context.Context, artifact ImportJobArtifact) error {
	if artifact.ID == "" {
		artifact.ID = uuid.NewString()
	}
	if artifact.CreatedAt.IsZero() {
		artifact.CreatedAt = time.Now()
	}
	return s.store.InsertJobArtifact(ctx, artifact)
}

func (s *Service) CreateManualJob(ctx context.Context, sourceID string, requestedBy *string) (ImportJob, error) {
	detail, err := s.sourceSvc.ValidateRunnableSource(ctx, sourceID)
	if err != nil {
		return ImportJob{}, err
	}
	source := detail.Source
	hasActive, err := s.store.HasActiveJobForSource(ctx, source.ID)
	if err != nil {
		return ImportJob{}, fmt.Errorf("check active job: %w", err)
	}
	if hasActive {
		return ImportJob{}, ErrActiveJobExists
	}
	now := time.Now()
	job, err := s.store.CreateJob(ctx, ImportJob{
		ID:                 uuid.NewString(),
		ConnectorSourceID:  source.ID,
		ConnectorVersionID: source.ConnectorVersionID,
		Status:             JobStatusQueued,
		RequestedBy:        requestedBy,
		TriggerType:        "manual",
		AuthMode:           source.AuthMode,
		ExecutionMode:      "unattended",
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return ImportJob{}, fmt.Errorf("create job: %w", err)
	}
	_ = s.store.InsertJobEvent(ctx, ImportJobEvent{ID: uuid.NewString(), ImportJobID: job.ID, EventType: "job.queued", Level: "info", MessageRedacted: "Import job queued.", MetadataRedacted: `{}`, CreatedAt: now})
	return job, nil
}

func (s *Service) CancelJob(ctx context.Context, jobID string) (ImportJob, error) {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return ImportJob{}, err
	}
	now := time.Now()
	switch job.Status {
	case JobStatusQueued:
		updated, err := s.store.UpdateJobStatus(ctx, jobID, UpdateJobStatusInput{Status: JobStatusCancelled, CancellationRequestedAt: &now, FinishedAt: &now})
		if err != nil {
			return ImportJob{}, fmt.Errorf("cancel job: %w", err)
		}
		_ = s.store.InsertJobEvent(ctx, ImportJobEvent{ID: uuid.NewString(), ImportJobID: job.ID, EventType: "job.cancelled", Level: "warning", MessageRedacted: "Import job cancelled before execution.", MetadataRedacted: `{}`, CreatedAt: now})
		return updated, nil
	case JobStatusRunning:
		updated, err := s.store.UpdateJobStatus(ctx, jobID, UpdateJobStatusInput{Status: JobStatusRunning, CancellationRequestedAt: &now})
		if err != nil {
			return ImportJob{}, fmt.Errorf("request running job cancellation: %w", err)
		}
		_ = s.store.InsertJobEvent(ctx, ImportJobEvent{ID: uuid.NewString(), ImportJobID: job.ID, EventType: "job.cancel_requested", Level: "warning", MessageRedacted: "Import job cancellation requested.", MetadataRedacted: `{}`, CreatedAt: now})
		return updated, nil
	case JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return job, nil
	default:
		return ImportJob{}, ErrInvalidJobState
	}
}

func (s *Service) TransitionJob(ctx context.Context, jobID string, status JobStatus, failureCode string, failureMessageRedacted string) (ImportJob, error) {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return ImportJob{}, err
	}
	if !isLegalTransition(job.Status, status) {
		return ImportJob{}, ErrInvalidJobState
	}
	now := time.Now()
	input := UpdateJobStatusInput{Status: status, FailureCode: failureCode, FailureMessageRedacted: failureMessageRedacted}
	if status == JobStatusRunning {
		input.StartedAt = &now
	}
	if status == JobStatusCompleted || status == JobStatusFailed || status == JobStatusCancelled {
		input.FinishedAt = &now
	}
	updated, err := s.store.UpdateJobStatus(ctx, jobID, input)
	if err != nil {
		return ImportJob{}, fmt.Errorf("transition job: %w", err)
	}
	_ = s.store.InsertJobEvent(ctx, ImportJobEvent{ID: uuid.NewString(), ImportJobID: job.ID, EventType: "job." + string(status), Level: "info", MessageRedacted: "Import job state changed.", MetadataRedacted: `{}`, CreatedAt: now})
	return updated, nil
}

func isLegalTransition(from JobStatus, to JobStatus) bool {
	switch from {
	case JobStatusQueued:
		return to == JobStatusRunning || to == JobStatusCancelled
	case JobStatusRunning:
		return to == JobStatusCompleted || to == JobStatusFailed || to == JobStatusCancelled
	default:
		return false
	}
}
