package jobs

import "time"

type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type ImportJob struct {
	ID                      string     `json:"id"`
	ConnectorSourceID       string     `json:"connector_source_id"`
	ConnectorVersionID      string     `json:"connector_version_id"`
	Status                  JobStatus  `json:"status"`
	RequestedBy             *string    `json:"requested_by,omitempty"`
	TriggerType             string     `json:"trigger_type"`
	AuthMode                string     `json:"auth_mode"`
	ExecutionMode           string     `json:"execution_mode"`
	StartedAt               *time.Time `json:"started_at,omitempty"`
	FinishedAt              *time.Time `json:"finished_at,omitempty"`
	TimeoutAt               *time.Time `json:"timeout_at,omitempty"`
	CancellationRequestedAt *time.Time `json:"cancellation_requested_at,omitempty"`
	FailureCode             string     `json:"failure_code,omitempty"`
	FailureMessageRedacted  string     `json:"failure_message_redacted,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type ImportJobEvent struct {
	ID               string    `json:"id"`
	ImportJobID      string    `json:"import_job_id"`
	EventType        string    `json:"event_type"`
	Level            string    `json:"level"`
	MessageRedacted  string    `json:"message_redacted"`
	MetadataRedacted string    `json:"metadata_redacted"`
	CreatedAt        time.Time `json:"created_at"`
}

type ImportJobArtifact struct {
	ID           string    `json:"id"`
	ImportJobID  string    `json:"import_job_id"`
	ArtifactType string    `json:"artifact_type"`
	RelativePath string    `json:"relative_path"`
	SizeBytes    int64     `json:"size_bytes"`
	SHA256       string    `json:"sha256"`
	CreatedAt    time.Time `json:"created_at"`
	StorageKey   string    `json:"-"`
}
