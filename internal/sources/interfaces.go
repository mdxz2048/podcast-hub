package sources

import (
	"context"
	"time"
)

type Store interface {
	ListSources(ctx context.Context) ([]ConnectorSource, error)
	GetSource(ctx context.Context, sourceID string) (ConnectorSource, bool, error)
	CreateSource(ctx context.Context, source ConnectorSource) (ConnectorSource, error)
	UpdateSource(ctx context.Context, sourceID string, in UpdateSourceInput) (ConnectorSource, error)
	SetSourceStatus(ctx context.Context, sourceID string, status SourceStatus) (ConnectorSource, error)

	ListSecrets(ctx context.Context) ([]SecretRecord, error)
	GetSecret(ctx context.Context, secretID string) (SecretRecord, bool, error)
	CreateSecret(ctx context.Context, secret SecretRecord) (SecretRecord, error)
	RevokeSecret(ctx context.Context, secretID string, revokedAt time.Time) (SecretRecord, error)

	ListSourceSecretBindings(ctx context.Context, sourceID string) ([]SourceSecretBinding, error)
	CreateSourceSecretBinding(ctx context.Context, binding SourceSecretBinding) (SourceSecretBinding, error)
	DeleteSourceSecretBinding(ctx context.Context, bindingID string) error

	InsertSourceEvent(ctx context.Context, event SourceEvent) error
}

type UpdateSourceInput struct {
	Name        string
	Description string
	ConfigJSON  string
	NetworkMode string
}

type SourceEvent struct {
	SourceID    *string
	SecretID    *string
	ActorUserID *string
	EventType   string
	DetailJSON  string
}
