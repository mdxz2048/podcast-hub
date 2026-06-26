package sources

import "time"

type SourceStatus string

const (
	SourceStatusDraft    SourceStatus = "draft"
	SourceStatusActive   SourceStatus = "active"
	SourceStatusDisabled SourceStatus = "disabled"
)

type SecretType string

const (
	SecretTypeText SecretType = "text"
	SecretTypeFile SecretType = "file"
)

type ConnectorSource struct {
	ID                 string       `json:"id"`
	ConnectorVersionID string       `json:"connector_version_id"`
	Name               string       `json:"name"`
	Description        string       `json:"description"`
	Status             SourceStatus `json:"status"`
	TriggerType        string       `json:"trigger_type"`
	AuthMode           string       `json:"auth_mode"`
	ExecutionMode      string       `json:"execution_mode"`
	ConfigJSON         string       `json:"config_json"`
	NetworkMode        string       `json:"network_mode"`
	CreatedBy          *string      `json:"created_by,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

type SecretRecord struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	SecretType        SecretType `json:"secret_type"`
	EncryptedPayload  string     `json:"-"`
	EncryptionVersion string     `json:"encryption_version"`
	CreatedBy         *string    `json:"created_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	RotatedAt         *time.Time `json:"rotated_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	BindingCount      int        `json:"binding_count"`
}

type SourceSecretBinding struct {
	ID                string    `json:"id"`
	ConnectorSourceID string    `json:"connector_source_id"`
	SecretName        string    `json:"secret_name"`
	SecretRecordID    string    `json:"secret_record_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type ConnectorSourceDetail struct {
	Source          ConnectorSource       `json:"source"`
	SecretBindings  []SourceSecretBinding `json:"secret_bindings"`
	RequiredSecrets []string              `json:"required_secrets"`
	MissingSecrets  []string              `json:"missing_secrets"`
}
