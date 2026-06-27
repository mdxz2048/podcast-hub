package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/content"
	"github.com/mdxz2048/podcast-hub/internal/intake"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

func TestStagingAPIRequiresAuthAndAdmin(t *testing.T) {
	server := newStagingAPITestServer()
	req := httptest.NewRequest(http.MethodGet, "/admin/staging/programs", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "u1", Email: "user@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive}, nil
	}
	req = httptest.NewRequest(http.MethodGet, "/admin/staging/programs", nil)
	req.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestStagingIntakeAPINoSensitiveLeak(t *testing.T) {
	server := newStagingAPITestServer()
	req := adminJobRequest(http.MethodPost, "/admin/import-jobs/job-completed/intake")
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, forbidden := range []string{"/Users/", ".local/", "storage_key", "staged_storage_key", "secret", "token", "cookie"} {
		if strings.Contains(strings.ToLower(rec.Body.String()), strings.ToLower(forbidden)) {
			t.Fatalf("intake response leaked %q: %s", forbidden, rec.Body.String())
		}
	}

	req = adminJobRequest(http.MethodGet, "/admin/staging/programs")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"review_pending"`) {
		t.Fatalf("expected staging program response, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func newStagingAPITestServer() *Server {
	cfg := config.Config{
		AppEnv:            "development",
		FrontendOrigin:    "http://127.0.0.1:5173",
		SessionCookieName: "podcast_hub_session",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	contentStore := newStagingMemoryContentStore()
	jobStore := &stagingMemoryJobs{
		job: jobs.ImportJob{ID: "job-completed", ConnectorSourceID: "source1", ConnectorVersionID: "version1", Status: jobs.JobStatusCompleted, CreatedAt: time.Now()},
		artifacts: []jobs.ImportJobArtifact{
			{ID: "bundle", ImportJobID: "job-completed", ArtifactType: "metadata_bundle", RelativePath: "bundle.json", SizeBytes: 120, SHA256: strings.Repeat("1", 64), StorageKey: "job-completed/bundle.json", CreatedAt: time.Now()},
			{ID: "audio", ImportJobID: "job-completed", ArtifactType: "audio", RelativePath: "audio.txt", SizeBytes: 12, SHA256: strings.Repeat("2", 64), StorageKey: "job-completed/audio.txt", CreatedAt: time.Now()},
		},
	}
	server := NewServer(cfg, nil, nil, HealthDependencies{}, nil, nil, nil, content.NewService(contentStore), intake.NewService(jobStore, contentStore, stagingArtifactReader{}))
	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}
	return server
}

type stagingMemoryJobs struct {
	job       jobs.ImportJob
	artifacts []jobs.ImportJobArtifact
}

func (m *stagingMemoryJobs) GetJob(context.Context, string) (jobs.ImportJob, error) {
	return m.job, nil
}
func (m *stagingMemoryJobs) ListArtifacts(context.Context, string) ([]jobs.ImportJobArtifact, error) {
	return m.artifacts, nil
}

type stagingArtifactReader struct{}

func (stagingArtifactReader) ReadArtifact(context.Context, jobs.ImportJobArtifact) ([]byte, error) {
	return []byte(`{"schema_version":1,"program":{"external_id":"program-1","title":"Candidate Program","description":"Program description","language":"zh-CN","author":"Fixture","cover_artifact":""},"episodes":[{"external_id":"ep-1","title":"Candidate Episode","description":"Episode description","published_at":"2026-06-27T00:00:00Z","duration_seconds":120,"audio_artifact":"audio.txt","cover_artifact":"","metadata":{}}]}`), nil
}

type stagingMemoryContentStore struct {
	programsByKey map[string]content.Program
	programs      map[string]content.Program
	episodes      map[string]content.Episode
	reviews       []content.ReviewItem
	events        []content.PublicationEvent
	media         []content.MediaAsset
	intakeRuns    map[string]content.IntakeRun
}

func newStagingMemoryContentStore() *stagingMemoryContentStore {
	return &stagingMemoryContentStore{programsByKey: map[string]content.Program{}, programs: map[string]content.Program{}, episodes: map[string]content.Episode{}, intakeRuns: map[string]content.IntakeRun{}}
}

func (m *stagingMemoryContentStore) UpsertProgramFromSource(_ context.Context, in content.UpsertProgramInput) (content.Program, error) {
	key := in.SourceID + ":" + in.ExternalProgramID
	item := m.programsByKey[key]
	if item.ID == "" {
		item = content.Program{ID: "program-1", CanonicalKey: key, CreatedFromSourceID: in.SourceID, CreatedFromJobID: in.JobID, CreatedAt: time.Now()}
	}
	item.Title = in.Title
	item.Description = in.Description
	item.Author = in.Author
	item.Language = in.Language
	item.Status = content.ProgramStatusReviewPending
	item.UpdatedAt = time.Now()
	m.programsByKey[key] = item
	m.programs[item.ID] = item
	return item, nil
}

func (m *stagingMemoryContentStore) UpsertEpisode(_ context.Context, in content.UpsertEpisodeInput) (content.Episode, error) {
	key := in.ProgramID + ":" + in.ExternalEpisodeID
	item := m.episodes[key]
	if item.ID == "" {
		item.ID = "episode-1"
		item.ProgramID = in.ProgramID
		item.ExternalEpisodeID = in.ExternalEpisodeID
		item.CreatedAt = time.Now()
	}
	item.Title = in.Title
	item.Description = in.Description
	item.PublishedAt = in.PublishedAt
	item.DurationSeconds = in.DurationSeconds
	item.SourceJobID = in.SourceJobID
	item.Status = content.EpisodeStatusReviewPending
	item.UpdatedAt = time.Now()
	m.episodes[key] = item
	return item, nil
}

func (m *stagingMemoryContentStore) CreateOrKeepPendingReview(_ context.Context, targetType string, targetID string, reviewKind string, jobID string) (content.ReviewItem, error) {
	item := content.ReviewItem{ID: targetType + "-" + targetID, TargetType: targetType, TargetID: targetID, ReviewKind: reviewKind, Status: content.ReviewStatusPending, RequestedByJobID: &jobID, CreatedAt: time.Now()}
	m.reviews = append(m.reviews, item)
	return item, nil
}
func (m *stagingMemoryContentStore) CreateMediaAsset(_ context.Context, in content.CreateMediaAssetInput) (content.MediaAsset, error) {
	item := content.MediaAsset{ID: "media-" + in.ArtifactID, OwnerType: in.OwnerType, OwnerID: in.OwnerID, ImportJobID: in.ImportJobID, ArtifactID: in.ArtifactID, MediaKind: in.MediaKind, StagedStorageKey: in.StagedStorageKey, SizeBytes: in.SizeBytes, SHA256: in.SHA256, Status: content.MediaStatusStaged, CreatedAt: time.Now()}
	m.media = append(m.media, item)
	return item, nil
}
func (m *stagingMemoryContentStore) InsertPublicationEvent(_ context.Context, event content.PublicationEvent) error {
	m.events = append(m.events, event)
	return nil
}
func (m *stagingMemoryContentStore) GetIntakeRun(_ context.Context, jobID string) (content.IntakeRun, bool, error) {
	item, ok := m.intakeRuns[jobID]
	return item, ok, nil
}
func (m *stagingMemoryContentStore) UpsertIntakeRun(_ context.Context, run content.IntakeRun) (content.IntakeRun, error) {
	m.intakeRuns[run.ImportJobID] = run
	return run, nil
}
func (m *stagingMemoryContentStore) ListStagingPrograms(context.Context) ([]content.Program, error) {
	items := []content.Program{}
	for _, item := range m.programs {
		items = append(items, item)
	}
	return items, nil
}
func (m *stagingMemoryContentStore) GetProgram(_ context.Context, id string) (content.Program, bool, error) {
	item, ok := m.programs[id]
	return item, ok, nil
}
func (m *stagingMemoryContentStore) ListStagingEpisodes(context.Context) ([]content.Episode, error) {
	items := []content.Episode{}
	for _, item := range m.episodes {
		items = append(items, item)
	}
	return items, nil
}
func (m *stagingMemoryContentStore) GetEpisode(_ context.Context, id string) (content.Episode, bool, error) {
	for _, item := range m.episodes {
		if item.ID == id {
			return item, true, nil
		}
	}
	return content.Episode{}, false, nil
}
func (m *stagingMemoryContentStore) ListAdminPrograms(ctx context.Context) ([]content.Program, error) {
	return m.ListStagingPrograms(ctx)
}
func (m *stagingMemoryContentStore) GetAdminProgram(ctx context.Context, id string) (content.Program, bool, error) {
	return m.GetProgram(ctx, id)
}
func (m *stagingMemoryContentStore) GetAdminEpisode(ctx context.Context, id string) (content.Episode, bool, error) {
	return m.GetEpisode(ctx, id)
}
func (m *stagingMemoryContentStore) ListProgramEpisodes(_ context.Context, programID string) ([]content.Episode, error) {
	items := []content.Episode{}
	for _, item := range m.episodes {
		if item.ProgramID == programID {
			items = append(items, item)
		}
	}
	return items, nil
}
func (m *stagingMemoryContentStore) ListReviews(context.Context) ([]content.ReviewItem, error) {
	return m.reviews, nil
}
func (m *stagingMemoryContentStore) GetReview(_ context.Context, id string) (content.ReviewItem, bool, error) {
	for _, item := range m.reviews {
		if item.ID == id {
			return item, true, nil
		}
	}
	return content.ReviewItem{}, false, nil
}
func (m *stagingMemoryContentStore) SetReviewDecision(_ context.Context, id string, status content.ReviewStatus, actorID string, note string) (content.ReviewItem, error) {
	for i, item := range m.reviews {
		if item.ID == id {
			item.Status = status
			item.ReviewedBy = &actorID
			item.ReviewNote = note
			now := time.Now()
			item.ReviewedAt = &now
			m.reviews[i] = item
			return item, nil
		}
	}
	return content.ReviewItem{}, nil
}
func (m *stagingMemoryContentStore) SetProgramStatus(ctx context.Context, id string, status content.ProgramStatus) (content.Program, error) {
	item, _, _ := m.GetProgram(ctx, id)
	item.Status = status
	m.programs[id] = item
	return item, nil
}
func (m *stagingMemoryContentStore) SetEpisodeStatus(ctx context.Context, id string, status content.EpisodeStatus) (content.Episode, error) {
	item, _, _ := m.GetEpisode(ctx, id)
	item.Status = status
	m.episodes[item.ProgramID+":"+item.ExternalEpisodeID] = item
	return item, nil
}
func (m *stagingMemoryContentStore) UpdateProgram(ctx context.Context, id string, in content.UpdateProgramInput) (content.Program, error) {
	item, _, _ := m.GetProgram(ctx, id)
	if in.Title != nil {
		item.Title = *in.Title
	}
	if in.Description != nil {
		item.Description = *in.Description
	}
	if in.Author != nil {
		item.Author = *in.Author
	}
	if in.Language != nil {
		item.Language = *in.Language
	}
	m.programs[id] = item
	return item, nil
}
func (m *stagingMemoryContentStore) UpdateEpisode(ctx context.Context, id string, in content.UpdateEpisodeInput) (content.Episode, error) {
	item, _, _ := m.GetEpisode(ctx, id)
	if in.Title != nil {
		item.Title = *in.Title
	}
	if in.Description != nil {
		item.Description = *in.Description
	}
	if in.DurationSeconds != nil {
		item.DurationSeconds = *in.DurationSeconds
	}
	m.episodes[item.ProgramID+":"+item.ExternalEpisodeID] = item
	return item, nil
}
func (m *stagingMemoryContentStore) CountPendingReviews(_ context.Context, targetType string, targetID string) (int, error) {
	count := 0
	for _, item := range m.reviews {
		if item.TargetType == targetType && item.TargetID == targetID && item.Status == content.ReviewStatusPending {
			count++
		}
	}
	return count, nil
}
func (m *stagingMemoryContentStore) HasApprovedMedia(_ context.Context, episodeID string) (bool, error) {
	for _, item := range m.media {
		if item.OwnerID == episodeID && item.MediaKind == "audio" && item.Status == content.MediaStatusApproved {
			return true, nil
		}
	}
	return false, nil
}
func (m *stagingMemoryContentStore) ApproveMediaForEpisode(_ context.Context, episodeID string) error {
	for i, item := range m.media {
		if item.OwnerID == episodeID {
			item.Status = content.MediaStatusApproved
			item.DeliveryStatus = content.MediaStatusApproved
			m.media[i] = item
		}
	}
	return nil
}
func (m *stagingMemoryContentStore) PromoteEpisodeMedia(_ context.Context, episodeID string) error {
	for i, item := range m.media {
		if item.OwnerID == episodeID && item.Status == content.MediaStatusApproved {
			now := time.Now()
			item.Status = content.MediaStatusPublished
			item.DeliveryStatus = content.MediaStatusPublished
			item.PublishedAt = &now
			item.PublishedKey = "episodes/" + episodeID + "/" + item.ID + ".bin"
			m.media[i] = item
		}
	}
	return nil
}
