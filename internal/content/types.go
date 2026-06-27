package content

import "time"

type ProgramStatus string
type EpisodeStatus string
type ReviewStatus string
type MediaStatus string

const (
	ProgramStatusReviewPending ProgramStatus = "review_pending"
	EpisodeStatusReviewPending EpisodeStatus = "review_pending"
	ReviewStatusPending        ReviewStatus  = "pending"
	MediaStatusStaged          MediaStatus   = "staged"
)

type Program struct {
	ID                  string        `json:"id"`
	CanonicalKey        string        `json:"canonical_key"`
	Title               string        `json:"title"`
	Description         string        `json:"description"`
	Author              string        `json:"author"`
	Language            string        `json:"language"`
	Status              ProgramStatus `json:"status"`
	CreatedFromSourceID string        `json:"created_from_source_id"`
	CreatedFromJobID    string        `json:"created_from_job_id"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

type Episode struct {
	ID                string        `json:"id"`
	ProgramID         string        `json:"program_id"`
	ExternalEpisodeID string        `json:"external_episode_id"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	PublishedAt       time.Time     `json:"published_at"`
	DurationSeconds   int           `json:"duration_seconds"`
	Status            EpisodeStatus `json:"status"`
	SourceJobID       string        `json:"source_job_id"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type ReviewItem struct {
	ID               string       `json:"id"`
	TargetType       string       `json:"target_type"`
	TargetID         string       `json:"target_id"`
	ReviewKind       string       `json:"review_kind"`
	Status           ReviewStatus `json:"status"`
	RequestedByJobID *string      `json:"requested_by_job_id,omitempty"`
	ReviewNote       string       `json:"review_note"`
	CreatedAt        time.Time    `json:"created_at"`
}

type MediaAsset struct {
	ID               string      `json:"id"`
	OwnerType        string      `json:"owner_type"`
	OwnerID          string      `json:"owner_id"`
	ImportJobID      string      `json:"import_job_id"`
	ArtifactID       string      `json:"artifact_id"`
	MediaKind        string      `json:"media_kind"`
	ContentType      string      `json:"content_type"`
	SizeBytes        int64       `json:"size_bytes"`
	SHA256           string      `json:"sha256"`
	Status           MediaStatus `json:"status"`
	CreatedAt        time.Time   `json:"created_at"`
	StagedStorageKey string      `json:"-"`
}

type PublicationEvent struct {
	ID               string    `json:"id"`
	TargetType       string    `json:"target_type"`
	TargetID         string    `json:"target_id"`
	EventType        string    `json:"event_type"`
	ActorID          *string   `json:"actor_id,omitempty"`
	MetadataRedacted string    `json:"metadata_redacted"`
	CreatedAt        time.Time `json:"created_at"`
}

type IntakeRun struct {
	ID                       string    `json:"id"`
	ImportJobID              string    `json:"import_job_id"`
	Status                   string    `json:"status"`
	ValidationIssuesRedacted string    `json:"validation_issues_redacted"`
	ProgramID                *string   `json:"program_id,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type UpsertProgramInput struct {
	SourceID          string
	JobID             string
	ExternalProgramID string
	Title             string
	Description       string
	Author            string
	Language          string
}

type UpsertEpisodeInput struct {
	ProgramID         string
	ExternalEpisodeID string
	Title             string
	Description       string
	PublishedAt       time.Time
	DurationSeconds   int
	SourceJobID       string
}

type CreateMediaAssetInput struct {
	OwnerType        string
	OwnerID          string
	ImportJobID      string
	ArtifactID       string
	MediaKind        string
	StagedStorageKey string
	ContentType      string
	SizeBytes        int64
	SHA256           string
}
