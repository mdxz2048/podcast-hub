package sources

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/connectors"
)

func TestSourceRequiresApprovedActiveConnector(t *testing.T) {
	store := newMemorySourceStore()
	connectorStore := newMemoryConnectorStoreForSources()
	service := newTestSourceService(store, connectorStore)

	if _, err := service.CreateSource(context.Background(), CreateSourceInput{
		ConnectorVersionID: "missing",
		Name:               "Source",
		TriggerType:        "manual",
		AuthMode:           "none",
		ExecutionMode:      "unattended",
		ConfigJSON:         "{}",
		NetworkMode:        "disabled",
	}); err != ErrConnectorVersionInvalid {
		t.Fatalf("expected invalid connector version, got %v", err)
	}

	connectorStore.seedApproved(t, "v1", connectors.ConnectorStatusDisabled, connectors.ReviewStatusApproved, "none", []string{"manual"}, "unattended", nil)
	if _, err := service.CreateSource(context.Background(), CreateSourceInput{
		ConnectorVersionID: "v1",
		Name:               "Source",
		TriggerType:        "manual",
		AuthMode:           "none",
		ExecutionMode:      "unattended",
		ConfigJSON:         "{}",
		NetworkMode:        "disabled",
	}); err != ErrConnectorUnavailable {
		t.Fatalf("expected unavailable connector, got %v", err)
	}
}

func TestSourceAlphaModeValidation(t *testing.T) {
	store := newMemorySourceStore()
	connectorStore := newMemoryConnectorStoreForSources()
	service := newTestSourceService(store, connectorStore)
	connectorStore.seedApproved(t, "v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "qr_each_run", []string{"scheduled"}, "interactive", nil)

	if _, err := service.CreateSource(context.Background(), CreateSourceInput{
		ConnectorVersionID: "v1",
		Name:               "Source",
		TriggerType:        "scheduled",
		AuthMode:           "qr_each_run",
		ExecutionMode:      "interactive",
		ConfigJSON:         "{}",
		NetworkMode:        "disabled",
	}); err != ErrUnsupportedAlphaMode {
		t.Fatalf("expected unsupported alpha mode, got %v", err)
	}
}

func TestSecretBindingAndRevocationRules(t *testing.T) {
	store := newMemorySourceStore()
	connectorStore := newMemoryConnectorStoreForSources()
	service := newTestSourceService(store, connectorStore)
	connectorStore.seedApproved(t, "v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "reusable_session", []string{"manual"}, "unattended", []string{"session_file"})

	detail, err := service.CreateSource(context.Background(), CreateSourceInput{
		ConnectorVersionID: "v1",
		Name:               "Source",
		TriggerType:        "manual",
		AuthMode:           "reusable_session",
		ExecutionMode:      "unattended",
		ConfigJSON:         "{}",
		NetworkMode:        "trusted_admin",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	if len(detail.MissingSecrets) != 1 {
		t.Fatalf("expected missing secret, got %+v", detail.MissingSecrets)
	}
	if _, err := service.EnableSource(context.Background(), detail.Source.ID, nil); err != ErrMissingRequiredSecrets {
		t.Fatalf("expected missing secret on enable, got %v", err)
	}

	secret, err := service.CreateSecret(context.Background(), "session", SecretTypeFile, []byte("plain secret"), nil)
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if secret.EncryptedPayload == "plain secret" || secret.EncryptedPayload == "" {
		t.Fatalf("secret payload was not encrypted")
	}
	detail, err = service.BindSecret(context.Background(), detail.Source.ID, "session_file", secret.ID, nil)
	if err != nil {
		t.Fatalf("bind secret: %v", err)
	}
	if len(detail.MissingSecrets) != 0 {
		t.Fatalf("expected no missing secrets, got %+v", detail.MissingSecrets)
	}
	if _, err := service.EnableSource(context.Background(), detail.Source.ID, nil); err != nil {
		t.Fatalf("enable source: %v", err)
	}
	if _, err := service.RevokeSecret(context.Background(), secret.ID, nil); err != nil {
		t.Fatalf("revoke secret: %v", err)
	}
	if _, err := service.EnableSource(context.Background(), detail.Source.ID, nil); err != ErrSecretRevoked {
		t.Fatalf("expected revoked secret error, got %v", err)
	}
}

func newTestSourceService(store *memorySourceStore, connectorStore *memoryConnectorStoreForSources) *Service {
	cipher, err := NewSecretCipher("12345678901234567890123456789012")
	if err != nil {
		panic(err)
	}
	return NewService(store, connectorStore, cipher)
}

type memoryConnectorStoreForSources struct {
	connectorsByID map[string]connectors.Connector
	versionsByID   map[string]connectors.ConnectorVersion
}

func newMemoryConnectorStoreForSources() *memoryConnectorStoreForSources {
	return &memoryConnectorStoreForSources{connectorsByID: map[string]connectors.Connector{}, versionsByID: map[string]connectors.ConnectorVersion{}}
}

func (m *memoryConnectorStoreForSources) seedApproved(t *testing.T, versionID string, connectorStatus connectors.ConnectorStatus, reviewStatus connectors.ReviewStatus, authMode string, triggers []string, executionMode string, secretNames []string) {
	t.Helper()
	secretDecls := make([]connectors.ConnectorSecretDecl, 0, len(secretNames))
	for _, name := range secretNames {
		secretDecls = append(secretDecls, connectors.ConnectorSecretDecl{Name: name, Description: name})
	}
	manifest := connectors.ConnectorManifest{
		SpecVersion:   1,
		ID:            "example-connector",
		Name:          "Example",
		Version:       "1.0.0",
		Runtime:       connectors.ConnectorRuntime{Language: "python", Profile: "python-basic", Entrypoint: "src/connector.py"},
		IngestionType: "connector",
		Trigger:       connectors.ConnectorTrigger{Allowed: triggers},
		Auth:          connectors.ConnectorAuth{Mode: authMode},
		Execution:     connectors.ConnectorExecution{Mode: executionMode, TimeoutSeconds: 900, MemoryMB: 512, MaxDownloadSizeMB: 2048},
		Secrets:       secretDecls,
		Network:       connectors.ConnectorNetworkPolicy{},
		Outputs:       connectors.ConnectorOutputs{Type: "podcast_episode_bundle"},
	}
	body, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	connectorID := "c-" + versionID
	m.connectorsByID[connectorID] = connectors.Connector{ID: connectorID, Slug: "example-" + versionID, Name: "Example", Status: connectorStatus}
	m.versionsByID[versionID] = connectors.ConnectorVersion{ID: versionID, ConnectorID: connectorID, Version: "1.0.0", ReviewStatus: reviewStatus, ManifestJSON: string(body)}
}

func (m *memoryConnectorStoreForSources) FindConnectorBySlug(context.Context, string) (connectors.Connector, bool, error) {
	return connectors.Connector{}, false, nil
}
func (m *memoryConnectorStoreForSources) CreateConnector(context.Context, connectors.Connector) (connectors.Connector, error) {
	return connectors.Connector{}, nil
}
func (m *memoryConnectorStoreForSources) UpdateConnectorStatus(context.Context, string, connectors.ConnectorStatus) (connectors.Connector, error) {
	return connectors.Connector{}, nil
}
func (m *memoryConnectorStoreForSources) GetConnector(_ context.Context, connectorID string) (connectors.Connector, bool, error) {
	item, ok := m.connectorsByID[connectorID]
	return item, ok, nil
}
func (m *memoryConnectorStoreForSources) ListConnectors(context.Context) ([]connectors.Connector, error) {
	return nil, nil
}
func (m *memoryConnectorStoreForSources) ListConnectorVersions(context.Context, string) ([]connectors.ConnectorVersion, error) {
	return nil, nil
}
func (m *memoryConnectorStoreForSources) GetConnectorVersion(_ context.Context, versionID string) (connectors.ConnectorVersion, bool, error) {
	item, ok := m.versionsByID[versionID]
	return item, ok, nil
}
func (m *memoryConnectorStoreForSources) GetConnectorVersionByVersion(context.Context, string, string) (connectors.ConnectorVersion, bool, error) {
	return connectors.ConnectorVersion{}, false, nil
}
func (m *memoryConnectorStoreForSources) CreateConnectorVersion(context.Context, connectors.ConnectorVersion) (connectors.ConnectorVersion, error) {
	return connectors.ConnectorVersion{}, nil
}
func (m *memoryConnectorStoreForSources) UpdateConnectorVersionReview(context.Context, string, connectors.UpdateVersionReviewInput) (connectors.ConnectorVersion, error) {
	return connectors.ConnectorVersion{}, nil
}
func (m *memoryConnectorStoreForSources) InsertConnectorEvent(context.Context, connectors.ConnectorEvent) error {
	return nil
}

type memorySourceStore struct {
	sources  map[string]ConnectorSource
	secrets  map[string]SecretRecord
	bindings map[string]SourceSecretBinding
	events   []SourceEvent
}

func newMemorySourceStore() *memorySourceStore {
	return &memorySourceStore{sources: map[string]ConnectorSource{}, secrets: map[string]SecretRecord{}, bindings: map[string]SourceSecretBinding{}}
}

func (m *memorySourceStore) ListSources(context.Context) ([]ConnectorSource, error) {
	items := make([]ConnectorSource, 0, len(m.sources))
	for _, item := range m.sources {
		items = append(items, item)
	}
	return items, nil
}
func (m *memorySourceStore) GetSource(_ context.Context, sourceID string) (ConnectorSource, bool, error) {
	item, ok := m.sources[sourceID]
	return item, ok, nil
}
func (m *memorySourceStore) CreateSource(_ context.Context, source ConnectorSource) (ConnectorSource, error) {
	m.sources[source.ID] = source
	return source, nil
}
func (m *memorySourceStore) UpdateSource(_ context.Context, sourceID string, in UpdateSourceInput) (ConnectorSource, error) {
	item := m.sources[sourceID]
	item.Name = in.Name
	item.Description = in.Description
	item.ConfigJSON = in.ConfigJSON
	item.NetworkMode = in.NetworkMode
	m.sources[sourceID] = item
	return item, nil
}
func (m *memorySourceStore) SetSourceStatus(_ context.Context, sourceID string, status SourceStatus) (ConnectorSource, error) {
	item := m.sources[sourceID]
	item.Status = status
	m.sources[sourceID] = item
	return item, nil
}
func (m *memorySourceStore) ListSecrets(context.Context) ([]SecretRecord, error) {
	items := make([]SecretRecord, 0, len(m.secrets))
	for _, item := range m.secrets {
		items = append(items, item)
	}
	return items, nil
}
func (m *memorySourceStore) GetSecret(_ context.Context, secretID string) (SecretRecord, bool, error) {
	item, ok := m.secrets[secretID]
	return item, ok, nil
}
func (m *memorySourceStore) CreateSecret(_ context.Context, secret SecretRecord) (SecretRecord, error) {
	m.secrets[secret.ID] = secret
	return secret, nil
}
func (m *memorySourceStore) RevokeSecret(_ context.Context, secretID string, revokedAt time.Time) (SecretRecord, error) {
	item := m.secrets[secretID]
	item.RevokedAt = &revokedAt
	m.secrets[secretID] = item
	return item, nil
}
func (m *memorySourceStore) ListSourceSecretBindings(_ context.Context, sourceID string) ([]SourceSecretBinding, error) {
	items := []SourceSecretBinding{}
	for _, item := range m.bindings {
		if item.ConnectorSourceID == sourceID {
			items = append(items, item)
		}
	}
	return items, nil
}
func (m *memorySourceStore) CreateSourceSecretBinding(_ context.Context, binding SourceSecretBinding) (SourceSecretBinding, error) {
	m.bindings[binding.ID] = binding
	return binding, nil
}
func (m *memorySourceStore) DeleteSourceSecretBinding(_ context.Context, bindingID string) error {
	delete(m.bindings, bindingID)
	return nil
}
func (m *memorySourceStore) InsertSourceEvent(_ context.Context, event SourceEvent) error {
	m.events = append(m.events, event)
	return nil
}
