package jobs

import (
	"context"
	"time"
)

type Store interface {
	ListJobs(ctx context.Context) ([]ImportJob, error)
	GetJob(ctx context.Context, jobID string) (ImportJob, bool, error)
	HasActiveJobForSource(ctx context.Context, sourceID string) (bool, error)
	ClaimNextQueuedJob(ctx context.Context) (ImportJob, bool, error)
	CreateJob(ctx context.Context, job ImportJob) (ImportJob, error)
	UpdateJobStatus(ctx context.Context, jobID string, in UpdateJobStatusInput) (ImportJob, error)
	ListJobEvents(ctx context.Context, jobID string) ([]ImportJobEvent, error)
	InsertJobEvent(ctx context.Context, event ImportJobEvent) error
	ListJobArtifacts(ctx context.Context, jobID string) ([]ImportJobArtifact, error)
	InsertJobArtifact(ctx context.Context, artifact ImportJobArtifact) error
}

type UpdateJobStatusInput struct {
	Status                  JobStatus
	StartedAt               *time.Time
	FinishedAt              *time.Time
	CancellationRequestedAt *time.Time
	FailureCode             string
	FailureMessageRedacted  string
}
