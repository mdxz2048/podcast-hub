package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/connectors"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

func TestCreateManualJobPreconditions(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*jobTestRig)
		sourceID  string
		wantError error
	}{
		{name: "source missing", sourceID: "missing", wantError: sources.ErrSourceNotFound},
		{name: "source disabled", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedConnector("v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "none", []string{"manual"}, "unattended", nil)
			r.seedSource("s1", "v1", sources.SourceStatusDisabled, "manual", "none", "unattended")
		}, wantError: sources.ErrUnsupportedAlphaMode},
		{name: "connector disabled", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedConnector("v1", connectors.ConnectorStatusDisabled, connectors.ReviewStatusApproved, "none", []string{"manual"}, "unattended", nil)
			r.seedSource("s1", "v1", sources.SourceStatusActive, "manual", "none", "unattended")
		}, wantError: sources.ErrConnectorUnavailable},
		{name: "version not approved", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedConnector("v1", connectors.ConnectorStatusActive, connectors.ReviewStatusRejected, "none", []string{"manual"}, "unattended", nil)
			r.seedSource("s1", "v1", sources.SourceStatusActive, "manual", "none", "unattended")
		}, wantError: sources.ErrConnectorVersionInvalid},
		{name: "required secret missing", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedConnector("v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "reusable_session", []string{"manual"}, "unattended", []string{"session_file"})
			r.seedSource("s1", "v1", sources.SourceStatusActive, "manual", "reusable_session", "unattended")
		}, wantError: sources.ErrMissingRequiredSecrets},
		{name: "secret revoked", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedConnector("v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "reusable_session", []string{"manual"}, "unattended", []string{"session_file"})
			r.seedSource("s1", "v1", sources.SourceStatusActive, "manual", "reusable_session", "unattended")
			r.seedSecret("sec1", true)
			r.bindSecret("s1", "session_file", "sec1")
		}, wantError: sources.ErrSecretRevoked},
		{name: "scheduled interactive rejected", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedConnector("v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "qr_each_run", []string{"scheduled"}, "interactive", nil)
			r.seedSource("s1", "v1", sources.SourceStatusActive, "scheduled", "qr_each_run", "interactive")
		}, wantError: sources.ErrUnsupportedAlphaMode},
		{name: "active job conflict", sourceID: "s1", setup: func(r *jobTestRig) {
			r.seedRunnableSource()
			r.jobs.byID["job-existing"] = ImportJob{ID: "job-existing", ConnectorSourceID: "s1", Status: JobStatusQueued}
		}, wantError: ErrActiveJobExists},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rig := newJobTestRig(t)
			if tt.setup != nil {
				tt.setup(rig)
			}
			_, err := rig.service.CreateManualJob(context.Background(), tt.sourceID, nil)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("expected %v, got %v", tt.wantError, err)
			}
		})
	}
}

func TestCreateManualJobAndCancelRules(t *testing.T) {
	rig := newJobTestRig(t)
	rig.seedRunnableSource()

	job, err := rig.service.CreateManualJob(context.Background(), "s1", nil)
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if job.Status != JobStatusQueued {
		t.Fatalf("expected queued, got %s", job.Status)
	}
	if len(rig.jobs.events) == 0 {
		t.Fatal("expected queued event")
	}

	cancelled, err := rig.service.CancelJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("cancel queued job: %v", err)
	}
	if cancelled.Status != JobStatusCancelled || cancelled.FinishedAt == nil {
		t.Fatalf("expected queued job to become cancelled with finished_at, got %+v", cancelled)
	}
	again, err := rig.service.CancelJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("terminal cancel should be idempotent: %v", err)
	}
	if again.Status != JobStatusCancelled {
		t.Fatalf("expected idempotent cancelled status, got %s", again.Status)
	}
}

func TestRunningCancelAndStateTransitions(t *testing.T) {
	rig := newJobTestRig(t)
	rig.seedRunnableSource()
	job, err := rig.service.CreateManualJob(context.Background(), "s1", nil)
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	running, err := rig.service.TransitionJob(context.Background(), job.ID, JobStatusRunning, "", "")
	if err != nil {
		t.Fatalf("queued to running: %v", err)
	}
	if running.StartedAt == nil {
		t.Fatal("expected started_at")
	}
	cancelRequested, err := rig.service.CancelJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("cancel running job: %v", err)
	}
	if cancelRequested.Status != JobStatusRunning || cancelRequested.CancellationRequestedAt == nil || cancelRequested.FinishedAt != nil {
		t.Fatalf("running cancel should only request cancellation, got %+v", cancelRequested)
	}
	completed, err := rig.service.TransitionJob(context.Background(), job.ID, JobStatusCompleted, "", "")
	if err != nil {
		t.Fatalf("running to completed: %v", err)
	}
	if completed.Status != JobStatusCompleted || completed.FinishedAt == nil {
		t.Fatalf("expected completed with finished_at, got %+v", completed)
	}
	if _, err := rig.service.TransitionJob(context.Background(), job.ID, JobStatusRunning, "", ""); !errors.Is(err, ErrInvalidJobState) {
		t.Fatalf("expected terminal state to reject transition, got %v", err)
	}
}

type jobTestRig struct {
	connectors *memoryConnectorStoreForJobs
	sources    *memorySourceStoreForJobs
	jobs       *memoryJobStore
	sourceSvc  *sources.Service
	service    *Service
}

func newJobTestRig(t *testing.T) *jobTestRig {
	t.Helper()
	connectorStore := newMemoryConnectorStoreForJobs()
	sourceStore := newMemorySourceStoreForJobs()
	cipher, err := sources.NewSecretCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	sourceSvc := sources.NewService(sourceStore, connectorStore, cipher)
	jobStore := newMemoryJobStore()
	return &jobTestRig{connectors: connectorStore, sources: sourceStore, jobs: jobStore, sourceSvc: sourceSvc, service: NewService(jobStore, sourceSvc)}
}

func (r *jobTestRig) seedRunnableSource() {
	r.seedConnector("v1", connectors.ConnectorStatusActive, connectors.ReviewStatusApproved, "none", []string{"manual"}, "unattended", nil)
	r.seedSource("s1", "v1", sources.SourceStatusActive, "manual", "none", "unattended")
}

func (r *jobTestRig) seedConnector(versionID string, connectorStatus connectors.ConnectorStatus, reviewStatus connectors.ReviewStatus, authMode string, triggers []string, executionMode string, secretNames []string) {
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
		Outputs:       connectors.ConnectorOutputs{Type: "podcast_episode_bundle"},
	}
	body, _ := json.Marshal(manifest)
	connectorID := "c-" + versionID
	r.connectors.connectorsByID[connectorID] = connectors.Connector{ID: connectorID, Slug: "example-" + versionID, Name: "Example", Status: connectorStatus}
	r.connectors.versionsByID[versionID] = connectors.ConnectorVersion{ID: versionID, ConnectorID: connectorID, Version: "1.0.0", ReviewStatus: reviewStatus, ManifestJSON: string(body)}
}

func (r *jobTestRig) seedSource(id string, versionID string, status sources.SourceStatus, trigger string, authMode string, execution string) {
	r.sources.sources[id] = sources.ConnectorSource{
		ID:                 id,
		ConnectorVersionID: versionID,
		Name:               "Source",
		Status:             status,
		TriggerType:        trigger,
		AuthMode:           authMode,
		ExecutionMode:      execution,
		ConfigJSON:         "{}",
		NetworkMode:        "disabled",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func (r *jobTestRig) seedSecret(id string, revoked bool) {
	secret := sources.SecretRecord{ID: id, Name: "Secret", SecretType: sources.SecretTypeFile, EncryptedPayload: "ciphertext", EncryptionVersion: sources.EncryptionVersionAESGCMV1}
	if revoked {
		now := time.Now()
		secret.RevokedAt = &now
	}
	r.sources.secrets[id] = secret
}

func (r *jobTestRig) bindSecret(sourceID string, secretName string, secretID string) {
	r.sources.bindings["binding-"+sourceID+"-"+secretName] = sources.SourceSecretBinding{ID: "binding-" + sourceID + "-" + secretName, ConnectorSourceID: sourceID, SecretName: secretName, SecretRecordID: secretID}
}

type memoryJobStore struct {
	byID      map[string]ImportJob
	events    []ImportJobEvent
	artifacts []ImportJobArtifact
}

func newMemoryJobStore() *memoryJobStore {
	return &memoryJobStore{byID: map[string]ImportJob{}}
}

func (m *memoryJobStore) ListJobs(context.Context) ([]ImportJob, error) {
	items := make([]ImportJob, 0, len(m.byID))
	for _, job := range m.byID {
		items = append(items, job)
	}
	return items, nil
}
func (m *memoryJobStore) GetJob(_ context.Context, jobID string) (ImportJob, bool, error) {
	job, ok := m.byID[jobID]
	return job, ok, nil
}
func (m *memoryJobStore) HasActiveJobForSource(_ context.Context, sourceID string) (bool, error) {
	for _, job := range m.byID {
		if job.ConnectorSourceID == sourceID && (job.Status == JobStatusQueued || job.Status == JobStatusRunning) {
			return true, nil
		}
	}
	return false, nil
}
func (m *memoryJobStore) CreateJob(_ context.Context, job ImportJob) (ImportJob, error) {
	m.byID[job.ID] = job
	return job, nil
}
func (m *memoryJobStore) UpdateJobStatus(_ context.Context, jobID string, in UpdateJobStatusInput) (ImportJob, error) {
	job := m.byID[jobID]
	job.Status = in.Status
	if in.StartedAt != nil {
		job.StartedAt = in.StartedAt
	}
	if in.FinishedAt != nil {
		job.FinishedAt = in.FinishedAt
	}
	if in.CancellationRequestedAt != nil {
		job.CancellationRequestedAt = in.CancellationRequestedAt
	}
	job.FailureCode = in.FailureCode
	job.FailureMessageRedacted = in.FailureMessageRedacted
	m.byID[jobID] = job
	return job, nil
}
func (m *memoryJobStore) ListJobEvents(_ context.Context, jobID string) ([]ImportJobEvent, error) {
	items := []ImportJobEvent{}
	for _, event := range m.events {
		if event.ImportJobID == jobID {
			items = append(items, event)
		}
	}
	return items, nil
}
func (m *memoryJobStore) InsertJobEvent(_ context.Context, event ImportJobEvent) error {
	m.events = append(m.events, event)
	return nil
}
func (m *memoryJobStore) ListJobArtifacts(_ context.Context, jobID string) ([]ImportJobArtifact, error) {
	items := []ImportJobArtifact{}
	for _, artifact := range m.artifacts {
		if artifact.ImportJobID == jobID {
			items = append(items, artifact)
		}
	}
	return items, nil
}

type memoryConnectorStoreForJobs struct {
	connectorsByID map[string]connectors.Connector
	versionsByID   map[string]connectors.ConnectorVersion
}

func newMemoryConnectorStoreForJobs() *memoryConnectorStoreForJobs {
	return &memoryConnectorStoreForJobs{connectorsByID: map[string]connectors.Connector{}, versionsByID: map[string]connectors.ConnectorVersion{}}
}
func (m *memoryConnectorStoreForJobs) FindConnectorBySlug(context.Context, string) (connectors.Connector, bool, error) {
	return connectors.Connector{}, false, nil
}
func (m *memoryConnectorStoreForJobs) CreateConnector(context.Context, connectors.Connector) (connectors.Connector, error) {
	return connectors.Connector{}, nil
}
func (m *memoryConnectorStoreForJobs) UpdateConnectorStatus(context.Context, string, connectors.ConnectorStatus) (connectors.Connector, error) {
	return connectors.Connector{}, nil
}
func (m *memoryConnectorStoreForJobs) GetConnector(_ context.Context, connectorID string) (connectors.Connector, bool, error) {
	item, ok := m.connectorsByID[connectorID]
	return item, ok, nil
}
func (m *memoryConnectorStoreForJobs) ListConnectors(context.Context) ([]connectors.Connector, error) {
	return nil, nil
}
func (m *memoryConnectorStoreForJobs) ListConnectorVersions(context.Context, string) ([]connectors.ConnectorVersion, error) {
	return nil, nil
}
func (m *memoryConnectorStoreForJobs) GetConnectorVersion(_ context.Context, versionID string) (connectors.ConnectorVersion, bool, error) {
	item, ok := m.versionsByID[versionID]
	return item, ok, nil
}
func (m *memoryConnectorStoreForJobs) GetConnectorVersionByVersion(context.Context, string, string) (connectors.ConnectorVersion, bool, error) {
	return connectors.ConnectorVersion{}, false, nil
}
func (m *memoryConnectorStoreForJobs) CreateConnectorVersion(context.Context, connectors.ConnectorVersion) (connectors.ConnectorVersion, error) {
	return connectors.ConnectorVersion{}, nil
}
func (m *memoryConnectorStoreForJobs) UpdateConnectorVersionReview(context.Context, string, connectors.UpdateVersionReviewInput) (connectors.ConnectorVersion, error) {
	return connectors.ConnectorVersion{}, nil
}
func (m *memoryConnectorStoreForJobs) InsertConnectorEvent(context.Context, connectors.ConnectorEvent) error {
	return nil
}

type memorySourceStoreForJobs struct {
	sources  map[string]sources.ConnectorSource
	secrets  map[string]sources.SecretRecord
	bindings map[string]sources.SourceSecretBinding
}

func newMemorySourceStoreForJobs() *memorySourceStoreForJobs {
	return &memorySourceStoreForJobs{sources: map[string]sources.ConnectorSource{}, secrets: map[string]sources.SecretRecord{}, bindings: map[string]sources.SourceSecretBinding{}}
}
func (m *memorySourceStoreForJobs) ListSources(context.Context) ([]sources.ConnectorSource, error) {
	return nil, nil
}
func (m *memorySourceStoreForJobs) GetSource(_ context.Context, sourceID string) (sources.ConnectorSource, bool, error) {
	item, ok := m.sources[sourceID]
	return item, ok, nil
}
func (m *memorySourceStoreForJobs) CreateSource(context.Context, sources.ConnectorSource) (sources.ConnectorSource, error) {
	return sources.ConnectorSource{}, nil
}
func (m *memorySourceStoreForJobs) UpdateSource(context.Context, string, sources.UpdateSourceInput) (sources.ConnectorSource, error) {
	return sources.ConnectorSource{}, nil
}
func (m *memorySourceStoreForJobs) SetSourceStatus(context.Context, string, sources.SourceStatus) (sources.ConnectorSource, error) {
	return sources.ConnectorSource{}, nil
}
func (m *memorySourceStoreForJobs) ListSecrets(context.Context) ([]sources.SecretRecord, error) {
	return nil, nil
}
func (m *memorySourceStoreForJobs) GetSecret(_ context.Context, secretID string) (sources.SecretRecord, bool, error) {
	item, ok := m.secrets[secretID]
	return item, ok, nil
}
func (m *memorySourceStoreForJobs) CreateSecret(context.Context, sources.SecretRecord) (sources.SecretRecord, error) {
	return sources.SecretRecord{}, nil
}
func (m *memorySourceStoreForJobs) RevokeSecret(context.Context, string, time.Time) (sources.SecretRecord, error) {
	return sources.SecretRecord{}, nil
}
func (m *memorySourceStoreForJobs) ListSourceSecretBindings(_ context.Context, sourceID string) ([]sources.SourceSecretBinding, error) {
	items := []sources.SourceSecretBinding{}
	for _, binding := range m.bindings {
		if binding.ConnectorSourceID == sourceID {
			items = append(items, binding)
		}
	}
	return items, nil
}
func (m *memorySourceStoreForJobs) CreateSourceSecretBinding(context.Context, sources.SourceSecretBinding) (sources.SourceSecretBinding, error) {
	return sources.SourceSecretBinding{}, nil
}
func (m *memorySourceStoreForJobs) DeleteSourceSecretBinding(context.Context, string) error {
	return nil
}
func (m *memorySourceStoreForJobs) InsertSourceEvent(context.Context, sources.SourceEvent) error {
	return nil
}
