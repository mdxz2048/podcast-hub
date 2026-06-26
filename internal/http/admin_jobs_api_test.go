package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/connectors"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

func TestImportJobAPIRequiresAuthAndAdmin(t *testing.T) {
	server := newJobAPITestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/sources/s1/import-jobs", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "u1", Email: "user@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive}, nil
	}
	req = httptest.NewRequest(http.MethodPost, "/admin/sources/s1/import-jobs", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestImportJobAPICreateListCancelAndNoSensitiveLeak(t *testing.T) {
	server := newJobAPITestServer(t)
	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}

	req := adminJobRequest(http.MethodPost, "/admin/sources/s1/import-jobs")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertJobResponseSafe(t, rec.Body.String())
	var createResp struct {
		Job jobs.ImportJob `json:"job"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Job.Status != jobs.JobStatusQueued {
		t.Fatalf("expected queued job, got %s", createResp.Job.Status)
	}

	req = adminJobRequest(http.MethodGet, "/admin/import-jobs")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", rec.Code)
	}

	req = adminJobRequest(http.MethodPost, "/admin/import-jobs/"+createResp.Job.ID+"/cancel")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected cancel 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"cancelled"`) {
		t.Fatalf("expected cancelled response, got %s", rec.Body.String())
	}
	assertJobResponseSafe(t, rec.Body.String())
}

func newJobAPITestServer(t *testing.T) *Server {
	t.Helper()
	cfg := config.Config{
		AppEnv:            "development",
		FrontendOrigin:    "http://127.0.0.1:5173",
		SessionCookieName: "podcast_hub_session",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	connectorStore := newJobAPIConnectorStore(t)
	sourceStore := newJobAPISourceStore()
	sourceStore.sources["s1"] = sources.ConnectorSource{ID: "s1", ConnectorVersionID: "v1", Name: "Source", Status: sources.SourceStatusActive, TriggerType: "manual", AuthMode: "none", ExecutionMode: "unattended", ConfigJSON: "{}", NetworkMode: "disabled"}
	cipher, err := sources.NewSecretCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	sourceService := sources.NewService(sourceStore, connectorStore, cipher)
	jobService := jobs.NewService(newJobAPIJobStore(), sourceService)
	return NewServer(cfg, nil, nil, HealthDependencies{}, nil, sourceService, jobService)
}

func adminJobRequest(method string, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.Header.Set("X-CSRF-Token", "csrf")
	req.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "admin-token"})
	req.AddCookie(&http.Cookie{Name: "podcast_hub_csrf", Value: "csrf"})
	return req
}

func assertJobResponseSafe(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"secret", "ciphertext", "package_storage_key", "quarantine/", "approved/", "/Users/", ".local/"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("job response leaked %q: %s", forbidden, body)
		}
	}
}

type jobAPIJobStore struct {
	byID   map[string]jobs.ImportJob
	events []jobs.ImportJobEvent
}

func newJobAPIJobStore() *jobAPIJobStore {
	return &jobAPIJobStore{byID: map[string]jobs.ImportJob{}}
}
func (s *jobAPIJobStore) ListJobs(context.Context) ([]jobs.ImportJob, error) {
	items := []jobs.ImportJob{}
	for _, job := range s.byID {
		items = append(items, job)
	}
	return items, nil
}
func (s *jobAPIJobStore) GetJob(_ context.Context, id string) (jobs.ImportJob, bool, error) {
	job, ok := s.byID[id]
	return job, ok, nil
}
func (s *jobAPIJobStore) HasActiveJobForSource(_ context.Context, sourceID string) (bool, error) {
	for _, job := range s.byID {
		if job.ConnectorSourceID == sourceID && (job.Status == jobs.JobStatusQueued || job.Status == jobs.JobStatusRunning) {
			return true, nil
		}
	}
	return false, nil
}
func (s *jobAPIJobStore) CreateJob(_ context.Context, job jobs.ImportJob) (jobs.ImportJob, error) {
	s.byID[job.ID] = job
	return job, nil
}
func (s *jobAPIJobStore) UpdateJobStatus(_ context.Context, id string, in jobs.UpdateJobStatusInput) (jobs.ImportJob, error) {
	job := s.byID[id]
	job.Status = in.Status
	job.CancellationRequestedAt = in.CancellationRequestedAt
	job.FinishedAt = in.FinishedAt
	s.byID[id] = job
	return job, nil
}
func (s *jobAPIJobStore) ListJobEvents(_ context.Context, jobID string) ([]jobs.ImportJobEvent, error) {
	items := []jobs.ImportJobEvent{}
	for _, event := range s.events {
		if event.ImportJobID == jobID {
			items = append(items, event)
		}
	}
	return items, nil
}
func (s *jobAPIJobStore) InsertJobEvent(_ context.Context, event jobs.ImportJobEvent) error {
	s.events = append(s.events, event)
	return nil
}
func (s *jobAPIJobStore) ListJobArtifacts(context.Context, string) ([]jobs.ImportJobArtifact, error) {
	return []jobs.ImportJobArtifact{}, nil
}

type jobAPIConnectorStore struct {
	connectors map[string]connectors.Connector
	versions   map[string]connectors.ConnectorVersion
}

func newJobAPIConnectorStore(t *testing.T) *jobAPIConnectorStore {
	t.Helper()
	manifest := connectors.ConnectorManifest{
		SpecVersion:   1,
		ID:            "example",
		Name:          "Example",
		Version:       "1.0.0",
		Runtime:       connectors.ConnectorRuntime{Language: "python", Profile: "python-basic", Entrypoint: "src/connector.py"},
		IngestionType: "connector",
		Trigger:       connectors.ConnectorTrigger{Allowed: []string{"manual"}},
		Auth:          connectors.ConnectorAuth{Mode: "none"},
		Execution:     connectors.ConnectorExecution{Mode: "unattended", TimeoutSeconds: 900, MemoryMB: 512, MaxDownloadSizeMB: 2048},
		Outputs:       connectors.ConnectorOutputs{Type: "podcast_episode_bundle"},
	}
	body, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	return &jobAPIConnectorStore{
		connectors: map[string]connectors.Connector{"c1": {ID: "c1", Slug: "example", Name: "Example", Status: connectors.ConnectorStatusActive}},
		versions:   map[string]connectors.ConnectorVersion{"v1": {ID: "v1", ConnectorID: "c1", Version: "1.0.0", ReviewStatus: connectors.ReviewStatusApproved, ManifestJSON: string(body)}},
	}
}
func (s *jobAPIConnectorStore) FindConnectorBySlug(context.Context, string) (connectors.Connector, bool, error) {
	return connectors.Connector{}, false, nil
}
func (s *jobAPIConnectorStore) CreateConnector(context.Context, connectors.Connector) (connectors.Connector, error) {
	return connectors.Connector{}, nil
}
func (s *jobAPIConnectorStore) UpdateConnectorStatus(context.Context, string, connectors.ConnectorStatus) (connectors.Connector, error) {
	return connectors.Connector{}, nil
}
func (s *jobAPIConnectorStore) GetConnector(_ context.Context, id string) (connectors.Connector, bool, error) {
	item, ok := s.connectors[id]
	return item, ok, nil
}
func (s *jobAPIConnectorStore) ListConnectors(context.Context) ([]connectors.Connector, error) {
	return nil, nil
}
func (s *jobAPIConnectorStore) ListConnectorVersions(context.Context, string) ([]connectors.ConnectorVersion, error) {
	return nil, nil
}
func (s *jobAPIConnectorStore) GetConnectorVersion(_ context.Context, id string) (connectors.ConnectorVersion, bool, error) {
	item, ok := s.versions[id]
	return item, ok, nil
}
func (s *jobAPIConnectorStore) GetConnectorVersionByVersion(context.Context, string, string) (connectors.ConnectorVersion, bool, error) {
	return connectors.ConnectorVersion{}, false, nil
}
func (s *jobAPIConnectorStore) CreateConnectorVersion(context.Context, connectors.ConnectorVersion) (connectors.ConnectorVersion, error) {
	return connectors.ConnectorVersion{}, nil
}
func (s *jobAPIConnectorStore) UpdateConnectorVersionReview(context.Context, string, connectors.UpdateVersionReviewInput) (connectors.ConnectorVersion, error) {
	return connectors.ConnectorVersion{}, nil
}
func (s *jobAPIConnectorStore) InsertConnectorEvent(context.Context, connectors.ConnectorEvent) error {
	return nil
}

type jobAPISourceStore struct {
	sources  map[string]sources.ConnectorSource
	secrets  map[string]sources.SecretRecord
	bindings map[string]sources.SourceSecretBinding
}

func newJobAPISourceStore() *jobAPISourceStore {
	return &jobAPISourceStore{sources: map[string]sources.ConnectorSource{}, secrets: map[string]sources.SecretRecord{}, bindings: map[string]sources.SourceSecretBinding{}}
}
func (s *jobAPISourceStore) ListSources(context.Context) ([]sources.ConnectorSource, error) {
	return nil, nil
}
func (s *jobAPISourceStore) GetSource(_ context.Context, id string) (sources.ConnectorSource, bool, error) {
	item, ok := s.sources[id]
	return item, ok, nil
}
func (s *jobAPISourceStore) CreateSource(context.Context, sources.ConnectorSource) (sources.ConnectorSource, error) {
	return sources.ConnectorSource{}, nil
}
func (s *jobAPISourceStore) UpdateSource(context.Context, string, sources.UpdateSourceInput) (sources.ConnectorSource, error) {
	return sources.ConnectorSource{}, nil
}
func (s *jobAPISourceStore) SetSourceStatus(context.Context, string, sources.SourceStatus) (sources.ConnectorSource, error) {
	return sources.ConnectorSource{}, nil
}
func (s *jobAPISourceStore) ListSecrets(context.Context) ([]sources.SecretRecord, error) {
	return nil, nil
}
func (s *jobAPISourceStore) GetSecret(_ context.Context, id string) (sources.SecretRecord, bool, error) {
	item, ok := s.secrets[id]
	return item, ok, nil
}
func (s *jobAPISourceStore) CreateSecret(context.Context, sources.SecretRecord) (sources.SecretRecord, error) {
	return sources.SecretRecord{}, nil
}
func (s *jobAPISourceStore) RevokeSecret(context.Context, string, time.Time) (sources.SecretRecord, error) {
	return sources.SecretRecord{}, nil
}
func (s *jobAPISourceStore) ListSourceSecretBindings(_ context.Context, sourceID string) ([]sources.SourceSecretBinding, error) {
	items := []sources.SourceSecretBinding{}
	for _, binding := range s.bindings {
		if binding.ConnectorSourceID == sourceID {
			items = append(items, binding)
		}
	}
	return items, nil
}
func (s *jobAPISourceStore) CreateSourceSecretBinding(context.Context, sources.SourceSecretBinding) (sources.SourceSecretBinding, error) {
	return sources.SourceSecretBinding{}, nil
}
func (s *jobAPISourceStore) DeleteSourceSecretBinding(context.Context, string) error {
	return nil
}
func (s *jobAPISourceStore) InsertSourceEvent(context.Context, sources.SourceEvent) error {
	return nil
}
