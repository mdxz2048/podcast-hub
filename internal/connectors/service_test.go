package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"
)

type memoryStore struct {
	connectorsByID map[string]Connector
	bySlug         map[string]string
	versionsByID   map[string]ConnectorVersion
	events         []ConnectorEvent
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		connectorsByID: map[string]Connector{},
		bySlug:         map[string]string{},
		versionsByID:   map[string]ConnectorVersion{},
	}
}

func (m *memoryStore) FindConnectorBySlug(_ context.Context, slug string) (Connector, bool, error) {
	id, ok := m.bySlug[slug]
	if !ok {
		return Connector{}, false, nil
	}
	connector := m.connectorsByID[id]
	return connector, true, nil
}

func (m *memoryStore) CreateConnector(_ context.Context, connector Connector) (Connector, error) {
	m.connectorsByID[connector.ID] = connector
	m.bySlug[connector.Slug] = connector.ID
	return connector, nil
}

func (m *memoryStore) UpdateConnectorStatus(_ context.Context, connectorID string, status ConnectorStatus) (Connector, error) {
	connector := m.connectorsByID[connectorID]
	connector.Status = status
	connector.UpdatedAt = time.Now()
	m.connectorsByID[connectorID] = connector
	return connector, nil
}

func (m *memoryStore) GetConnector(_ context.Context, connectorID string) (Connector, bool, error) {
	connector, ok := m.connectorsByID[connectorID]
	return connector, ok, nil
}

func (m *memoryStore) ListConnectors(_ context.Context) ([]Connector, error) {
	items := make([]Connector, 0, len(m.connectorsByID))
	for _, c := range m.connectorsByID {
		items = append(items, c)
	}
	return items, nil
}

func (m *memoryStore) ListConnectorVersions(_ context.Context, connectorID string) ([]ConnectorVersion, error) {
	items := make([]ConnectorVersion, 0)
	for _, v := range m.versionsByID {
		if v.ConnectorID == connectorID {
			items = append(items, v)
		}
	}
	return items, nil
}

func (m *memoryStore) GetConnectorVersion(_ context.Context, versionID string) (ConnectorVersion, bool, error) {
	v, ok := m.versionsByID[versionID]
	return v, ok, nil
}

func (m *memoryStore) GetConnectorVersionByVersion(_ context.Context, connectorID, version string) (ConnectorVersion, bool, error) {
	for _, v := range m.versionsByID {
		if v.ConnectorID == connectorID && v.Version == version {
			return v, true, nil
		}
	}
	return ConnectorVersion{}, false, nil
}

func (m *memoryStore) CreateConnectorVersion(_ context.Context, version ConnectorVersion) (ConnectorVersion, error) {
	m.versionsByID[version.ID] = version
	return version, nil
}

func (m *memoryStore) UpdateConnectorVersionReview(_ context.Context, versionID string, in UpdateVersionReviewInput) (ConnectorVersion, error) {
	v := m.versionsByID[versionID]
	v.ReviewStatus = in.ReviewStatus
	v.ReviewedBy = in.ReviewedBy
	v.ReviewedAt = in.ReviewedAt
	if in.PackageStorageKey != nil {
		v.PackageStorageKey = *in.PackageStorageKey
	}
	m.versionsByID[versionID] = v
	return v, nil
}

func (m *memoryStore) InsertConnectorEvent(_ context.Context, event ConnectorEvent) error {
	m.events = append(m.events, event)
	return nil
}

type memoryPackageStore struct {
	packages map[string][]byte
}

func newMemoryPackageStore() *memoryPackageStore {
	return &memoryPackageStore{packages: map[string][]byte{}}
}

func (m *memoryPackageStore) PutQuarantine(_ context.Context, packageName string, content io.Reader) (PackageRef, error) {
	body, _ := io.ReadAll(content)
	key := "quarantine/" + packageName
	m.packages[key] = body
	return PackageRef{StorageKey: key, SizeBytes: int64(len(body)), SHA256: "sha"}, nil
}
func (m *memoryPackageStore) Read(_ context.Context, ref PackageRef) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.packages[ref.StorageKey])), nil
}
func (m *memoryPackageStore) PromoteApproved(_ context.Context, ref PackageRef) (PackageRef, error) {
	data := m.packages[ref.StorageKey]
	delete(m.packages, ref.StorageKey)
	newKey := "approved/" + ref.StorageKey
	m.packages[newKey] = data
	ref.StorageKey = newKey
	return ref, nil
}
func (m *memoryPackageStore) Delete(_ context.Context, ref PackageRef) error {
	delete(m.packages, ref.StorageKey)
	return nil
}

func TestServiceUploadAndStateMachine(t *testing.T) {
	store := newMemoryStore()
	packageStore := newMemoryPackageStore()
	service := NewService(store, packageStore)
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(manifestForVersion("1.0.0"))},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	actorID := "admin-1"

	uploadResult, err := service.Upload(context.Background(), UploadInput{
		ConnectorID: "example-connector",
		Version:     "1.0.0",
		PackageName: "connector.zip",
		Content:     bytes.NewReader(zipData),
		UploadedBy:  &actorID,
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if uploadResult.Version.ReviewStatus != ReviewStatusPendingReview {
		t.Fatalf("expected pending_review, got %s", uploadResult.Version.ReviewStatus)
	}
	if len(store.events) == 0 {
		t.Fatal("expected connector event on upload")
	}

	if _, err := service.Upload(context.Background(), UploadInput{
		ConnectorID: "example-connector",
		Version:     "1.0.0",
		PackageName: "connector.zip",
		Content:     bytes.NewReader(zipData),
		UploadedBy:  &actorID,
	}); err != ErrVersionAlreadyExists {
		t.Fatalf("expected ErrVersionAlreadyExists, got %v", err)
	}

	approved, err := service.ApproveVersion(context.Background(), uploadResult.Version.ID, &actorID)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if approved.ReviewStatus != ReviewStatusApproved {
		t.Fatalf("expected approved status, got %s", approved.ReviewStatus)
	}
	if !stringsHasPrefix(approved.PackageStorageKey, "approved/") {
		t.Fatalf("expected promoted package key, got %s", approved.PackageStorageKey)
	}

	disabled, err := service.DisableVersion(context.Background(), uploadResult.Version.ID, &actorID)
	if err != nil {
		t.Fatalf("disable version failed: %v", err)
	}
	if disabled.ReviewStatus != ReviewStatusDisabled {
		t.Fatalf("expected disabled review status, got %s", disabled.ReviewStatus)
	}
}

func TestServiceRejectAndDisabledConnectorBlockApprove(t *testing.T) {
	store := newMemoryStore()
	packageStore := newMemoryPackageStore()
	service := NewService(store, packageStore)
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(manifestForVersion("1.0.1"))},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	actorID := "admin-1"
	uploadResult, err := service.Upload(context.Background(), UploadInput{
		ConnectorID: "example-connector",
		Version:     "1.0.1",
		PackageName: "connector.zip",
		Content:     bytes.NewReader(zipData),
		UploadedBy:  &actorID,
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	rejected, err := service.RejectVersion(context.Background(), uploadResult.Version.ID, &actorID)
	if err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	if rejected.ReviewStatus != ReviewStatusRejected {
		t.Fatalf("expected rejected status, got %s", rejected.ReviewStatus)
	}
	if _, err := service.ApproveVersion(context.Background(), uploadResult.Version.ID, &actorID); err != ErrVersionNotPendingReview {
		t.Fatalf("expected ErrVersionNotPendingReview, got %v", err)
	}

	uploadResult2, err := service.Upload(context.Background(), UploadInput{
		ConnectorID: "example-connector",
		Version:     "1.0.2",
		PackageName: "connector.zip",
		Content: bytes.NewReader(buildZip(t, []zipEntry{
			{name: "manifest.yaml", body: []byte(manifestForVersion("1.0.2"))},
			{name: "requirements.lock", body: []byte("pkg==1.0.0")},
			{name: "README.md", body: []byte("# readme")},
			{name: "src/connector.py", body: []byte("print('ok')")},
		})),
		UploadedBy: &actorID,
	})
	if err != nil {
		t.Fatalf("second upload failed: %v", err)
	}
	if _, err := service.SetConnectorStatus(context.Background(), uploadResult2.Connector.ID, ConnectorStatusDisabled, &actorID); err != nil {
		t.Fatalf("disable connector failed: %v", err)
	}
	if _, err := service.ApproveVersion(context.Background(), uploadResult2.Version.ID, &actorID); err != ErrConnectorDisabled {
		t.Fatalf("expected ErrConnectorDisabled, got %v", err)
	}
}

func TestValidationSummaryJSONIsStored(t *testing.T) {
	store := newMemoryStore()
	packageStore := newMemoryPackageStore()
	service := NewService(store, packageStore)
	badManifest := []byte("spec_version: 1\nid: bad\n")
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: badManifest},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	actorID := "admin-1"
	result, err := service.Upload(context.Background(), UploadInput{
		ConnectorID: "example-connector",
		Version:     "2.0.0",
		PackageName: "connector.zip",
		Content:     bytes.NewReader(zipData),
		UploadedBy:  &actorID,
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	var summary ValidationSummary
	if err := json.Unmarshal([]byte(result.Version.ValidationSummaryJSON), &summary); err != nil {
		t.Fatalf("decode validation summary json: %v", err)
	}
	if summary.IsValid {
		t.Fatal("expected invalid summary")
	}
}

func stringsHasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}

func manifestForVersion(version string) string {
	return strings.ReplaceAll(validManifestYAML, "version: 1.0.0", "version: "+version)
}
