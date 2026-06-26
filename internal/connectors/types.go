package connectors

import "time"

type ConnectorStatus string

const (
	ConnectorStatusActive   ConnectorStatus = "active"
	ConnectorStatusDisabled ConnectorStatus = "disabled"
)

type ReviewStatus string

const (
	ReviewStatusPendingReview ReviewStatus = "pending_review"
	ReviewStatusApproved      ReviewStatus = "approved"
	ReviewStatusRejected      ReviewStatus = "rejected"
	ReviewStatusDisabled      ReviewStatus = "disabled"
)

type Connector struct {
	ID          string          `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      ConnectorStatus `json:"status"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ConnectorVersion struct {
	ID                    string       `json:"id"`
	ConnectorID           string       `json:"connector_id"`
	Version               string       `json:"version"`
	ReviewStatus          ReviewStatus `json:"review_status"`
	RuntimeProfile        string       `json:"runtime_profile"`
	Entrypoint            string       `json:"entrypoint"`
	ManifestJSON          string       `json:"manifest_json"`
	PackageSHA256         string       `json:"package_sha256"`
	PackageSizeBytes      int64        `json:"package_size_bytes"`
	PackageStorageKey     string       `json:"package_storage_key"`
	ValidationSummaryJSON string       `json:"validation_summary_json"`
	UploadedBy            *string      `json:"uploaded_by,omitempty"`
	ReviewedBy            *string      `json:"reviewed_by,omitempty"`
	ReviewedAt            *time.Time   `json:"reviewed_at,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
}

type ValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type ValidationSummary struct {
	IsValid bool              `json:"is_valid"`
	Issues  []ValidationIssue `json:"issues"`
}

type ConnectorManifest struct {
	SpecVersion   int                    `yaml:"spec_version" json:"spec_version"`
	ID            string                 `yaml:"id" json:"id"`
	Name          string                 `yaml:"name" json:"name"`
	Version       string                 `yaml:"version" json:"version"`
	Description   string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Runtime       ConnectorRuntime       `yaml:"runtime" json:"runtime"`
	IngestionType string                 `yaml:"ingestion_type" json:"ingestion_type"`
	Trigger       ConnectorTrigger       `yaml:"trigger" json:"trigger"`
	Auth          ConnectorAuth          `yaml:"auth" json:"auth"`
	Execution     ConnectorExecution     `yaml:"execution" json:"execution"`
	Inputs        []map[string]any       `yaml:"inputs" json:"inputs"`
	Secrets       []ConnectorSecretDecl  `yaml:"secrets" json:"secrets"`
	Network       ConnectorNetworkPolicy `yaml:"network" json:"network"`
	Outputs       ConnectorOutputs       `yaml:"outputs" json:"outputs"`
}

type ConnectorRuntime struct {
	Language   string `yaml:"language" json:"language"`
	Profile    string `yaml:"profile" json:"profile"`
	Entrypoint string `yaml:"entrypoint" json:"entrypoint"`
}

type ConnectorTrigger struct {
	Allowed []string `yaml:"allowed" json:"allowed"`
}

type ConnectorAuth struct {
	Mode string `yaml:"mode" json:"mode"`
}

type ConnectorExecution struct {
	Mode              string `yaml:"mode" json:"mode"`
	TimeoutSeconds    int    `yaml:"timeout_seconds" json:"timeout_seconds"`
	MemoryMB          int    `yaml:"memory_mb" json:"memory_mb"`
	MaxDownloadSizeMB int    `yaml:"max_download_size_mb" json:"max_download_size_mb"`
}

type ConnectorSecretDecl struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Value       string `yaml:"value,omitempty" json:"-"`
}

type ConnectorNetworkPolicy struct {
	Allowlist []string `yaml:"allowlist" json:"allowlist"`
}

type ConnectorOutputs struct {
	Type string `yaml:"type" json:"type"`
}

type PackageRef struct {
	StorageKey string
	SizeBytes  int64
	SHA256     string
}
