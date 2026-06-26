package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mdxz2048/podcast-hub/internal/connectors"
)

const MaxSecretBytes = 64 * 1024

type Service struct {
	store          Store
	connectorStore connectors.Store
	cipher         *SecretCipher
}

func NewService(store Store, connectorStore connectors.Store, cipher *SecretCipher) *Service {
	return &Service{store: store, connectorStore: connectorStore, cipher: cipher}
}

type CreateSourceInput struct {
	ConnectorVersionID string
	Name               string
	Description        string
	TriggerType        string
	AuthMode           string
	ExecutionMode      string
	ConfigJSON         string
	NetworkMode        string
	CreatedBy          *string
}

func (s *Service) ListSources(ctx context.Context) ([]ConnectorSource, error) {
	return s.store.ListSources(ctx)
}

func (s *Service) GetSourceDetail(ctx context.Context, sourceID string) (ConnectorSourceDetail, error) {
	source, found, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("get source: %w", err)
	}
	if !found {
		return ConnectorSourceDetail{}, ErrSourceNotFound
	}
	return s.detailForSource(ctx, source)
}

func (s *Service) CreateSource(ctx context.Context, in CreateSourceInput) (ConnectorSourceDetail, error) {
	if err := validateAlphaSourceFields(in.TriggerType, in.AuthMode, in.ExecutionMode, in.NetworkMode); err != nil {
		return ConnectorSourceDetail{}, err
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.ConnectorVersionID) == "" {
		return ConnectorSourceDetail{}, ErrInvalidInput
	}
	version, connector, manifest, err := s.loadApprovedConnector(ctx, in.ConnectorVersionID)
	if err != nil {
		return ConnectorSourceDetail{}, err
	}
	if err := validateAlphaManifest(manifest, in.AuthMode); err != nil {
		return ConnectorSourceDetail{}, err
	}
	now := time.Now()
	configJSON := strings.TrimSpace(in.ConfigJSON)
	if configJSON == "" {
		configJSON = "{}"
	}
	if !json.Valid([]byte(configJSON)) {
		return ConnectorSourceDetail{}, ErrInvalidInput
	}
	source, err := s.store.CreateSource(ctx, ConnectorSource{
		ID:                 uuid.NewString(),
		ConnectorVersionID: version.ID,
		Name:               strings.TrimSpace(in.Name),
		Description:        strings.TrimSpace(in.Description),
		Status:             SourceStatusDraft,
		TriggerType:        "manual",
		AuthMode:           in.AuthMode,
		ExecutionMode:      "unattended",
		ConfigJSON:         configJSON,
		NetworkMode:        in.NetworkMode,
		CreatedBy:          in.CreatedBy,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("create source: %w", err)
	}
	sourceIDCopy := source.ID
	_ = connector
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SourceID: &sourceIDCopy, ActorUserID: in.CreatedBy, EventType: "source.created", DetailJSON: `{"status":"draft"}`})
	return s.detailForSource(ctx, source)
}

func (s *Service) UpdateSource(ctx context.Context, sourceID string, in UpdateSourceInput, actorUserID *string) (ConnectorSourceDetail, error) {
	if strings.TrimSpace(in.Name) == "" {
		return ConnectorSourceDetail{}, ErrInvalidInput
	}
	if strings.TrimSpace(in.ConfigJSON) == "" {
		in.ConfigJSON = "{}"
	}
	if !json.Valid([]byte(in.ConfigJSON)) {
		return ConnectorSourceDetail{}, ErrInvalidInput
	}
	if in.NetworkMode != "disabled" && in.NetworkMode != "trusted_admin" {
		return ConnectorSourceDetail{}, ErrInvalidInput
	}
	source, err := s.store.UpdateSource(ctx, sourceID, in)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("update source: %w", err)
	}
	sourceIDCopy := source.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SourceID: &sourceIDCopy, ActorUserID: actorUserID, EventType: "source.updated", DetailJSON: `{}`})
	return s.detailForSource(ctx, source)
}

func (s *Service) EnableSource(ctx context.Context, sourceID string, actorUserID *string) (ConnectorSourceDetail, error) {
	source, found, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("get source: %w", err)
	}
	if !found {
		return ConnectorSourceDetail{}, ErrSourceNotFound
	}
	if _, _, _, err := s.loadApprovedConnector(ctx, source.ConnectorVersionID); err != nil {
		return ConnectorSourceDetail{}, err
	}
	detail, err := s.detailForSource(ctx, source)
	if err != nil {
		return ConnectorSourceDetail{}, err
	}
	if len(detail.MissingSecrets) > 0 {
		return ConnectorSourceDetail{}, ErrMissingRequiredSecrets
	}
	for _, binding := range detail.SecretBindings {
		secret, found, err := s.store.GetSecret(ctx, binding.SecretRecordID)
		if err != nil {
			return ConnectorSourceDetail{}, fmt.Errorf("get secret: %w", err)
		}
		if !found {
			return ConnectorSourceDetail{}, ErrSecretNotFound
		}
		if secret.RevokedAt != nil {
			return ConnectorSourceDetail{}, ErrSecretRevoked
		}
	}
	updated, err := s.store.SetSourceStatus(ctx, sourceID, SourceStatusActive)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("enable source: %w", err)
	}
	sourceIDCopy := updated.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SourceID: &sourceIDCopy, ActorUserID: actorUserID, EventType: "source.enabled", DetailJSON: `{"status":"active"}`})
	return s.detailForSource(ctx, updated)
}

func (s *Service) DisableSource(ctx context.Context, sourceID string, actorUserID *string) (ConnectorSourceDetail, error) {
	source, err := s.store.SetSourceStatus(ctx, sourceID, SourceStatusDisabled)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("disable source: %w", err)
	}
	sourceIDCopy := source.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SourceID: &sourceIDCopy, ActorUserID: actorUserID, EventType: "source.disabled", DetailJSON: `{"status":"disabled"}`})
	return s.detailForSource(ctx, source)
}

func (s *Service) ListSecrets(ctx context.Context) ([]SecretRecord, error) {
	return s.store.ListSecrets(ctx)
}

func (s *Service) CreateSecret(ctx context.Context, name string, secretType SecretType, payload []byte, actorUserID *string) (SecretRecord, error) {
	if strings.TrimSpace(name) == "" || (secretType != SecretTypeText && secretType != SecretTypeFile) {
		return SecretRecord{}, ErrInvalidInput
	}
	if len(payload) == 0 || len(payload) > MaxSecretBytes {
		return SecretRecord{}, ErrSecretTooLarge
	}
	encrypted, err := s.cipher.Encrypt(payload)
	if err != nil {
		return SecretRecord{}, fmt.Errorf("encrypt secret: %w", err)
	}
	now := time.Now()
	secret, err := s.store.CreateSecret(ctx, SecretRecord{
		ID:                uuid.NewString(),
		Name:              strings.TrimSpace(name),
		SecretType:        secretType,
		EncryptedPayload:  encrypted,
		EncryptionVersion: EncryptionVersionAESGCMV1,
		CreatedBy:         actorUserID,
		CreatedAt:         now,
	})
	if err != nil {
		return SecretRecord{}, fmt.Errorf("create secret: %w", err)
	}
	secretIDCopy := secret.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SecretID: &secretIDCopy, ActorUserID: actorUserID, EventType: "secret.created", DetailJSON: fmt.Sprintf(`{"secret_type":"%s"}`, secretType)})
	return secret, nil
}

func (s *Service) RevokeSecret(ctx context.Context, secretID string, actorUserID *string) (SecretRecord, error) {
	secret, err := s.store.RevokeSecret(ctx, secretID, time.Now())
	if err != nil {
		return SecretRecord{}, fmt.Errorf("revoke secret: %w", err)
	}
	secretIDCopy := secret.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SecretID: &secretIDCopy, ActorUserID: actorUserID, EventType: "secret.revoked", DetailJSON: `{}`})
	return secret, nil
}

func (s *Service) BindSecret(ctx context.Context, sourceID string, secretName string, secretID string, actorUserID *string) (ConnectorSourceDetail, error) {
	source, found, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("get source: %w", err)
	}
	if !found {
		return ConnectorSourceDetail{}, ErrSourceNotFound
	}
	secret, found, err := s.store.GetSecret(ctx, secretID)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("get secret: %w", err)
	}
	if !found {
		return ConnectorSourceDetail{}, ErrSecretNotFound
	}
	if secret.RevokedAt != nil {
		return ConnectorSourceDetail{}, ErrSecretRevoked
	}
	_, _, manifest, err := s.loadApprovedConnector(ctx, source.ConnectorVersionID)
	if err != nil {
		return ConnectorSourceDetail{}, err
	}
	if !slices.Contains(requiredSecretNames(manifest), secretName) {
		return ConnectorSourceDetail{}, ErrInvalidInput
	}
	binding, err := s.store.CreateSourceSecretBinding(ctx, SourceSecretBinding{
		ID:                uuid.NewString(),
		ConnectorSourceID: source.ID,
		SecretName:        secretName,
		SecretRecordID:    secret.ID,
		CreatedAt:         time.Now(),
	})
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("bind secret: %w", err)
	}
	sourceIDCopy := source.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SourceID: &sourceIDCopy, ActorUserID: actorUserID, EventType: "source.secret_bound", DetailJSON: fmt.Sprintf(`{"secret_name":"%s","binding_id":"%s"}`, secretName, binding.ID)})
	return s.detailForSource(ctx, source)
}

func (s *Service) DeleteBinding(ctx context.Context, sourceID string, bindingID string, actorUserID *string) (ConnectorSourceDetail, error) {
	source, found, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("get source: %w", err)
	}
	if !found {
		return ConnectorSourceDetail{}, ErrSourceNotFound
	}
	if err := s.store.DeleteSourceSecretBinding(ctx, bindingID); err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("delete binding: %w", err)
	}
	sourceIDCopy := source.ID
	_ = s.store.InsertSourceEvent(ctx, SourceEvent{SourceID: &sourceIDCopy, ActorUserID: actorUserID, EventType: "source.secret_unbound", DetailJSON: `{}`})
	return s.detailForSource(ctx, source)
}

func (s *Service) loadApprovedConnector(ctx context.Context, versionID string) (connectors.ConnectorVersion, connectors.Connector, connectors.ConnectorManifest, error) {
	version, found, err := s.connectorStore.GetConnectorVersion(ctx, versionID)
	if err != nil {
		return connectors.ConnectorVersion{}, connectors.Connector{}, connectors.ConnectorManifest{}, fmt.Errorf("get connector version: %w", err)
	}
	if !found || version.ReviewStatus != connectors.ReviewStatusApproved {
		return connectors.ConnectorVersion{}, connectors.Connector{}, connectors.ConnectorManifest{}, ErrConnectorVersionInvalid
	}
	connector, found, err := s.connectorStore.GetConnector(ctx, version.ConnectorID)
	if err != nil {
		return connectors.ConnectorVersion{}, connectors.Connector{}, connectors.ConnectorManifest{}, fmt.Errorf("get connector: %w", err)
	}
	if !found || connector.Status != connectors.ConnectorStatusActive {
		return connectors.ConnectorVersion{}, connectors.Connector{}, connectors.ConnectorManifest{}, ErrConnectorUnavailable
	}
	var manifest connectors.ConnectorManifest
	if err := json.Unmarshal([]byte(version.ManifestJSON), &manifest); err != nil {
		return connectors.ConnectorVersion{}, connectors.Connector{}, connectors.ConnectorManifest{}, ErrConnectorVersionInvalid
	}
	return version, connector, manifest, nil
}

func (s *Service) detailForSource(ctx context.Context, source ConnectorSource) (ConnectorSourceDetail, error) {
	_, _, manifest, err := s.loadApprovedConnector(ctx, source.ConnectorVersionID)
	if err != nil {
		return ConnectorSourceDetail{}, err
	}
	bindings, err := s.store.ListSourceSecretBindings(ctx, source.ID)
	if err != nil {
		return ConnectorSourceDetail{}, fmt.Errorf("list bindings: %w", err)
	}
	required := requiredSecretNames(manifest)
	bound := map[string]struct{}{}
	for _, binding := range bindings {
		bound[binding.SecretName] = struct{}{}
	}
	missing := make([]string, 0)
	for _, name := range required {
		if _, ok := bound[name]; !ok {
			missing = append(missing, name)
		}
	}
	return ConnectorSourceDetail{Source: source, SecretBindings: bindings, RequiredSecrets: required, MissingSecrets: missing}, nil
}

func validateAlphaSourceFields(triggerType string, authMode string, executionMode string, networkMode string) error {
	if triggerType != "manual" || executionMode != "unattended" {
		return ErrUnsupportedAlphaMode
	}
	if authMode != "none" && authMode != "reusable_session" {
		return ErrUnsupportedAlphaMode
	}
	if networkMode != "disabled" && networkMode != "trusted_admin" {
		return ErrInvalidInput
	}
	return nil
}

func validateAlphaManifest(manifest connectors.ConnectorManifest, authMode string) error {
	if len(manifest.Trigger.Allowed) != 1 || manifest.Trigger.Allowed[0] != "manual" {
		return ErrUnsupportedAlphaMode
	}
	if manifest.Auth.Mode != authMode || (manifest.Auth.Mode != "none" && manifest.Auth.Mode != "reusable_session") {
		return ErrUnsupportedAlphaMode
	}
	if manifest.Execution.Mode != "unattended" {
		return ErrUnsupportedAlphaMode
	}
	return nil
}

func requiredSecretNames(manifest connectors.ConnectorManifest) []string {
	names := make([]string, 0, len(manifest.Secrets))
	for _, secret := range manifest.Secrets {
		name := strings.TrimSpace(secret.Name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}
