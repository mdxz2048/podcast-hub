package http

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/connectors"
)

func TestConnectorUploadRequiresAuthAndAdmin(t *testing.T) {
	server := newConnectorTestServer(t)

	req := httptest.NewRequest(stdhttp.MethodPost, "/admin/connectors/upload", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=foo")
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "u1", Email: "user@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive}, nil
	}
	req = httptest.NewRequest(stdhttp.MethodPost, "/admin/connectors/upload", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=foo")
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestConnectorUploadAndReviewFlow(t *testing.T) {
	server, store := newConnectorTestServerWithStore(t)
	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}
	rec, response := uploadConnectorPackage(t, server, "example-connector", "1.0.0", buildValidConnectorZip(t))
	if !response.ValidationSummary.IsValid {
		t.Fatalf("expected valid upload, got body=%s", rec.Body.String())
	}
	assertNoPackageLeak(t, rec.Body.String())
	connectorID, _ := response.Connector["id"].(string)
	if connectorID == "" {
		t.Fatalf("missing connector id in response: %s", rec.Body.String())
	}
	versionsReq := httptest.NewRequest(stdhttp.MethodGet, "/admin/connectors/"+connectorID+"/versions", nil)
	versionsReq.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "admin-token"})
	versionsRec := httptest.NewRecorder()
	server.Router().ServeHTTP(versionsRec, versionsReq)
	if versionsRec.Code != stdhttp.StatusOK {
		t.Fatalf("expected versions 200, got %d", versionsRec.Code)
	}
	assertNoPackageLeak(t, versionsRec.Body.String())

	versionID, _ := response.Version["id"].(string)
	stateReq := adminRequest(stdhttp.MethodPost, "/admin/connector-versions/"+versionID+"/approve", nil)
	stateRec := httptest.NewRecorder()
	server.Router().ServeHTTP(stateRec, stateReq)
	if stateRec.Code != stdhttp.StatusOK {
		t.Fatalf("expected approve 200, got %d, body=%s", stateRec.Code, stateRec.Body.String())
	}
	stateReq = adminRequest(stdhttp.MethodPost, "/admin/connector-versions/"+versionID+"/disable", nil)
	stateRec = httptest.NewRecorder()
	server.Router().ServeHTTP(stateRec, stateReq)
	if stateRec.Code != stdhttp.StatusOK {
		t.Fatalf("expected disable version 200, got %d, body=%s", stateRec.Code, stateRec.Body.String())
	}
	connectorReq := adminRequest(stdhttp.MethodPost, "/admin/connectors/"+connectorID+"/disable", nil)
	connectorRec := httptest.NewRecorder()
	server.Router().ServeHTTP(connectorRec, connectorReq)
	if connectorRec.Code != stdhttp.StatusOK {
		t.Fatalf("expected disable connector 200, got %d, body=%s", connectorRec.Code, connectorRec.Body.String())
	}
	connectorReq = adminRequest(stdhttp.MethodPost, "/admin/connectors/"+connectorID+"/enable", nil)
	connectorRec = httptest.NewRecorder()
	server.Router().ServeHTTP(connectorRec, connectorReq)
	if connectorRec.Code != stdhttp.StatusOK {
		t.Fatalf("expected enable connector 200, got %d, body=%s", connectorRec.Code, connectorRec.Body.String())
	}
	if events := len(store.events); events < 4 {
		t.Fatalf("expected connector events, got %d", events)
	}
}

func TestConnectorUploadHTTPValidationFailures(t *testing.T) {
	tests := []struct {
		name      string
		content   []byte
		version   string
		wantCode  int
		wantIssue string
	}{
		{name: "non zip", content: []byte("not a zip"), version: "1.0.0", wantCode: stdhttp.StatusBadRequest},
		{name: "zip slip", content: connectorZip(t, validManifestYAML("1.0.0"), []testZipEntry{{name: "../escape.py", body: []byte("x")}}), version: "1.0.0", wantCode: stdhttp.StatusCreated, wantIssue: "zip_path_invalid"},
		{name: "duplicate path", content: connectorZip(t, validManifestYAML("1.0.19"), []testZipEntry{{name: "src/connector.py", body: []byte("print('duplicate')")}}), version: "1.0.19", wantCode: stdhttp.StatusCreated, wantIssue: "zip_duplicate_path"},
		{name: "symlink", content: connectorZip(t, validManifestYAML("1.0.1"), []testZipEntry{{name: "src/link.py", body: []byte("x"), mode: os.ModeSymlink}}), version: "1.0.1", wantCode: stdhttp.StatusCreated, wantIssue: "zip_symlink_forbidden"},
		{name: "compression ratio", content: connectorZip(t, validManifestYAML("1.0.16"), []testZipEntry{{name: "fixtures/repeated.txt", body: bytes.Repeat([]byte("A"), 1024*1024)}}), version: "1.0.16", wantCode: stdhttp.StatusCreated, wantIssue: "zip_compression_ratio_too_high"},
		{name: "too many files", content: manyFilesZip(t, validManifestYAML("1.0.2"), 220), version: "1.0.2", wantCode: stdhttp.StatusCreated, wantIssue: "zip_too_many_files"},
		{name: "single file too large", content: connectorZip(t, validManifestYAML("1.0.17"), []testZipEntry{{name: "fixtures/large.txt", body: bytes.Repeat([]byte("B"), 6*1024*1024)}}), version: "1.0.17", wantCode: stdhttp.StatusCreated, wantIssue: "file_too_large"},
		{name: "uncompressed total too large", content: connectorZip(t, validManifestYAML("1.0.18"), []testZipEntry{{name: "fixtures/huge.txt", body: bytes.Repeat([]byte("C"), 51*1024*1024)}}), version: "1.0.18", wantCode: stdhttp.StatusCreated, wantIssue: "zip_uncompressed_too_large"},
		{name: "sensitive file", content: connectorZip(t, validManifestYAML("1.0.3"), []testZipEntry{{name: ".env", body: []byte("SECRET=value")}}), version: "1.0.3", wantCode: stdhttp.StatusCreated, wantIssue: "forbidden_file"},
		{name: "binary", content: connectorZip(t, validManifestYAML("1.0.4"), []testZipEntry{{name: "src/blob.py", body: []byte{0x00, 0x01}}}), version: "1.0.4", wantCode: stdhttp.StatusCreated, wantIssue: "binary_forbidden"},
		{name: "missing manifest", content: connectorZipWithoutManifest(t, "1.0.5"), version: "1.0.5", wantCode: stdhttp.StatusCreated, wantIssue: "manifest_missing"},
		{name: "unknown manifest field", content: connectorZip(t, validManifestYAML("1.0.6")+"\nunknown: true\n", nil), version: "1.0.6", wantCode: stdhttp.StatusCreated, wantIssue: "manifest_invalid"},
		{name: "id mismatch", content: connectorZip(t, strings.ReplaceAll(validManifestYAML("1.0.7"), "id: example-connector", "id: other-connector"), nil), version: "1.0.7", wantCode: stdhttp.StatusCreated, wantIssue: "manifest_id_mismatch"},
		{name: "version mismatch", content: connectorZip(t, validManifestYAML("9.9.9"), nil), version: "1.0.8", wantCode: stdhttp.StatusCreated, wantIssue: "manifest_version_mismatch"},
		{name: "non python", content: connectorZip(t, strings.ReplaceAll(validManifestYAML("1.0.9"), "language: python", "language: go"), nil), version: "1.0.9", wantCode: stdhttp.StatusCreated, wantIssue: "runtime_language_invalid"},
		{name: "bad profile", content: connectorZip(t, strings.ReplaceAll(validManifestYAML("1.0.10"), "profile: python-basic", "profile: python-root"), nil), version: "1.0.10", wantCode: stdhttp.StatusCreated, wantIssue: "runtime_profile_invalid"},
		{name: "missing requirements", content: connectorZipMissingRequirements(t, "1.0.11"), version: "1.0.11", wantCode: stdhttp.StatusCreated, wantIssue: "requirements_lock_missing"},
		{name: "missing entrypoint", content: connectorZip(t, strings.ReplaceAll(validManifestYAML("1.0.12"), "src/connector.py", "src/missing.py"), nil), version: "1.0.12", wantCode: stdhttp.StatusCreated, wantIssue: "entrypoint_not_found"},
		{name: "qr scheduled", content: connectorZip(t, strings.ReplaceAll(strings.ReplaceAll(validManifestYAML("1.0.13"), "mode: none", "mode: qr_each_run"), "- manual", "- scheduled"), nil), version: "1.0.13", wantCode: stdhttp.StatusCreated, wantIssue: "qr_scheduled_forbidden"},
		{name: "network", content: connectorZip(t, strings.ReplaceAll(validManifestYAML("1.0.14"), "- api.example.invalid", "- https://api.example.invalid/path"), nil), version: "1.0.14", wantCode: stdhttp.StatusCreated, wantIssue: "network_host_invalid"},
		{name: "secret value", content: connectorZip(t, strings.ReplaceAll(validManifestYAML("1.0.15"), "secrets: []", "secrets:\n  - name: api_token\n    description: token\n    value: plain"), nil), version: "1.0.15", wantCode: stdhttp.StatusCreated, wantIssue: "secret_value_forbidden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newAdminConnectorTestServer(t)
			rec, response := uploadConnectorPackage(t, server, "example-connector", tt.version, tt.content)
			if rec.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d body=%s", tt.wantCode, rec.Code, rec.Body.String())
			}
			assertNoPackageLeak(t, rec.Body.String())
			if tt.wantIssue != "" && !hasHTTPValidationIssue(response.ValidationSummary.Issues, tt.wantIssue) {
				t.Fatalf("expected issue %s, got %+v", tt.wantIssue, response.ValidationSummary.Issues)
			}
		})
	}
}

func TestConnectorUploadRejectsDuplicateVersion(t *testing.T) {
	server := newAdminConnectorTestServer(t)
	rec, _ := uploadConnectorPackage(t, server, "example-connector", "1.0.0", buildValidConnectorZip(t))
	if rec.Code != stdhttp.StatusCreated {
		t.Fatalf("expected first upload 201, got %d", rec.Code)
	}
	rec, _ = uploadConnectorPackage(t, server, "example-connector", "1.0.0", buildValidConnectorZip(t))
	if rec.Code != stdhttp.StatusConflict {
		t.Fatalf("expected duplicate 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func newConnectorTestServer(t *testing.T) *Server {
	t.Helper()
	server, _ := newConnectorTestServerWithStore(t)
	return server
}

func newConnectorTestServerWithStore(t *testing.T) (*Server, *httpMemoryConnectorStore) {
	t.Helper()
	cfg := config.Config{
		AppEnv:            "development",
		FrontendOrigin:    "http://127.0.0.1:5173",
		SessionCookieName: "podcast_hub_session",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	store := newHTTPMemoryConnectorStore()
	service := connectors.NewService(store, newHTTPMemoryPackageStore())
	return NewServer(cfg, nil, nil, HealthDependencies{}, service), store
}

func newAdminConnectorTestServer(t *testing.T) *Server {
	t.Helper()
	server := newConnectorTestServer(t)
	server.resolveSessionFn = func(_ context.Context, token string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}
	return server
}

type httpMemoryConnectorStore struct {
	connectors map[string]connectors.Connector
	slugIndex  map[string]string
	versions   map[string]connectors.ConnectorVersion
	events     []connectors.ConnectorEvent
}

func newHTTPMemoryConnectorStore() *httpMemoryConnectorStore {
	return &httpMemoryConnectorStore{
		connectors: map[string]connectors.Connector{},
		slugIndex:  map[string]string{},
		versions:   map[string]connectors.ConnectorVersion{},
	}
}

func (s *httpMemoryConnectorStore) FindConnectorBySlug(_ context.Context, slug string) (connectors.Connector, bool, error) {
	id, ok := s.slugIndex[slug]
	if !ok {
		return connectors.Connector{}, false, nil
	}
	return s.connectors[id], true, nil
}
func (s *httpMemoryConnectorStore) CreateConnector(_ context.Context, connector connectors.Connector) (connectors.Connector, error) {
	s.connectors[connector.ID] = connector
	s.slugIndex[connector.Slug] = connector.ID
	return connector, nil
}
func (s *httpMemoryConnectorStore) UpdateConnectorStatus(_ context.Context, connectorID string, status connectors.ConnectorStatus) (connectors.Connector, error) {
	c := s.connectors[connectorID]
	c.Status = status
	s.connectors[connectorID] = c
	return c, nil
}
func (s *httpMemoryConnectorStore) GetConnector(_ context.Context, connectorID string) (connectors.Connector, bool, error) {
	c, ok := s.connectors[connectorID]
	return c, ok, nil
}
func (s *httpMemoryConnectorStore) ListConnectors(_ context.Context) ([]connectors.Connector, error) {
	items := make([]connectors.Connector, 0, len(s.connectors))
	for _, c := range s.connectors {
		items = append(items, c)
	}
	return items, nil
}
func (s *httpMemoryConnectorStore) ListConnectorVersions(_ context.Context, connectorID string) ([]connectors.ConnectorVersion, error) {
	items := make([]connectors.ConnectorVersion, 0)
	for _, v := range s.versions {
		if v.ConnectorID == connectorID {
			items = append(items, v)
		}
	}
	return items, nil
}
func (s *httpMemoryConnectorStore) GetConnectorVersion(_ context.Context, versionID string) (connectors.ConnectorVersion, bool, error) {
	v, ok := s.versions[versionID]
	return v, ok, nil
}
func (s *httpMemoryConnectorStore) GetConnectorVersionByVersion(_ context.Context, connectorID, version string) (connectors.ConnectorVersion, bool, error) {
	for _, v := range s.versions {
		if v.ConnectorID == connectorID && v.Version == version {
			return v, true, nil
		}
	}
	return connectors.ConnectorVersion{}, false, nil
}
func (s *httpMemoryConnectorStore) CreateConnectorVersion(_ context.Context, version connectors.ConnectorVersion) (connectors.ConnectorVersion, error) {
	s.versions[version.ID] = version
	return version, nil
}
func (s *httpMemoryConnectorStore) UpdateConnectorVersionReview(_ context.Context, versionID string, in connectors.UpdateVersionReviewInput) (connectors.ConnectorVersion, error) {
	v := s.versions[versionID]
	v.ReviewStatus = in.ReviewStatus
	v.ReviewedBy = in.ReviewedBy
	v.ReviewedAt = in.ReviewedAt
	if in.PackageStorageKey != nil {
		v.PackageStorageKey = *in.PackageStorageKey
	}
	s.versions[versionID] = v
	return v, nil
}
func (s *httpMemoryConnectorStore) InsertConnectorEvent(_ context.Context, event connectors.ConnectorEvent) error {
	s.events = append(s.events, event)
	return nil
}

type httpMemoryPackageStore struct {
	contents map[string][]byte
	next     int
}

func newHTTPMemoryPackageStore() *httpMemoryPackageStore {
	return &httpMemoryPackageStore{contents: map[string][]byte{}}
}
func (s *httpMemoryPackageStore) PutQuarantine(_ context.Context, packageName string, content io.Reader) (connectors.PackageRef, error) {
	_ = packageName
	body, _ := io.ReadAll(content)
	s.next++
	key := "quarantine/test-package-" + string(rune('a'+s.next)) + ".zip"
	s.contents[key] = body
	return connectors.PackageRef{StorageKey: key, SizeBytes: int64(len(body)), SHA256: "sha"}, nil
}
func (s *httpMemoryPackageStore) Read(_ context.Context, ref connectors.PackageRef) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.contents[ref.StorageKey])), nil
}
func (s *httpMemoryPackageStore) PromoteApproved(_ context.Context, ref connectors.PackageRef) (connectors.PackageRef, error) {
	ref.StorageKey = "approved/" + ref.StorageKey
	return ref, nil
}
func (s *httpMemoryPackageStore) Delete(_ context.Context, ref connectors.PackageRef) error {
	delete(s.contents, ref.StorageKey)
	return nil
}

type connectorUploadResponse struct {
	Connector         map[string]any `json:"connector"`
	Version           map[string]any `json:"version"`
	ValidationSummary struct {
		IsValid bool `json:"is_valid"`
		Issues  []struct {
			Code string `json:"code"`
		} `json:"issues"`
	} `json:"validation_summary"`
}

func uploadConnectorPackage(t *testing.T, server *Server, connectorID string, version string, content []byte) (*httptest.ResponseRecorder, connectorUploadResponse) {
	t.Helper()
	body, contentType := multipartUploadBody(t, connectorID, version, "connector.zip", content)
	req := adminRequest(stdhttp.MethodPost, "/admin/connectors/upload", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	var response connectorUploadResponse
	if rec.Code == stdhttp.StatusCreated {
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode upload response: %v", err)
		}
	}
	return rec, response
}

func adminRequest(method string, target string, body io.Reader) *stdhttp.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.Header.Set("X-CSRF-Token", "csrf")
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_session", Value: "admin-token"})
	req.AddCookie(&stdhttp.Cookie{Name: "podcast_hub_csrf", Value: "csrf"})
	return req
}

func assertNoPackageLeak(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"package_storage_key", "quarantine/", "approved/", "connector-packages", "connector.py"} {
		if strings.Contains(body, forbidden) && forbidden != "connector.py" {
			t.Fatalf("response leaked package detail %q: %s", forbidden, body)
		}
	}
}

func hasHTTPValidationIssue(issues []struct {
	Code string `json:"code"`
}, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func buildValidConnectorZip(t *testing.T) []byte {
	t.Helper()
	return connectorZip(t, validManifestYAML("1.0.0"), nil)
}

type testZipEntry struct {
	name string
	body []byte
	mode os.FileMode
}

func connectorZip(t *testing.T, manifest string, extra []testZipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	write := func(entry testZipEntry) {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		f, err := w.CreateHeader(header)
		if err != nil {
			t.Fatalf("create zip file: %v", err)
		}
		if _, err := f.Write(entry.body); err != nil {
			t.Fatalf("write zip file: %v", err)
		}
	}
	write(testZipEntry{name: "manifest.yaml", body: []byte(manifest)})
	write(testZipEntry{name: "requirements.lock", body: []byte("pkg==1.0.0")})
	write(testZipEntry{name: "README.md", body: []byte("# readme")})
	write(testZipEntry{name: "src/connector.py", body: []byte("print('ok')")})
	for _, entry := range extra {
		write(entry)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func connectorZipWithoutManifest(t *testing.T, version string) []byte {
	t.Helper()
	_ = version
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, body := range map[string]string{
		"requirements.lock": "pkg==1.0.0",
		"README.md":         "# readme",
		"src/connector.py":  "print('ok')",
	} {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip file: %v", err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatalf("write zip file: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func connectorZipMissingRequirements(t *testing.T, version string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, body := range map[string]string{
		"manifest.yaml":    validManifestYAML(version),
		"README.md":        "# readme",
		"src/connector.py": "print('ok')",
	} {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip file: %v", err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatalf("write zip file: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func manyFilesZip(t *testing.T, manifest string, count int) []byte {
	t.Helper()
	extra := make([]testZipEntry, 0, count)
	for i := 0; i < count; i++ {
		extra = append(extra, testZipEntry{name: "fixtures/file-" + string(rune('a'+(i%26))) + string(rune('a'+((i/26)%26))) + ".txt", body: []byte("x")})
	}
	return connectorZip(t, manifest, extra)
}

func validManifestYAML(version string) string {
	return `spec_version: 1
id: example-connector
name: Example Connector
version: ` + version + `
runtime:
  language: python
  profile: python-basic
  entrypoint: src/connector.py
ingestion_type: connector
trigger:
  allowed:
    - manual
auth:
  mode: none
execution:
  mode: unattended
  timeout_seconds: 900
  memory_mb: 512
  max_download_size_mb: 2048
inputs: []
secrets: []
network:
  allowlist:
    - api.example.invalid
outputs:
  type: podcast_episode_bundle
`
}

func multipartUploadBody(t *testing.T, connectorID, version, fileName string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("connector_id", connectorID)
	_ = writer.WriteField("version", version)
	fileWriter, err := writer.CreateFormFile("package", fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return &body, writer.FormDataContentType()
}
