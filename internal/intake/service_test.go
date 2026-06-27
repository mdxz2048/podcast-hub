package intake

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/content"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

func TestIntakeRejectsUnsafeInputs(t *testing.T) {
	tests := []struct {
		name      string
		jobStatus jobs.JobStatus
		artifacts []jobs.ImportJobArtifact
		bundle    string
		wantErr   error
	}{
		{name: "job not completed", jobStatus: jobs.JobStatusQueued, wantErr: ErrJobNotCompleted},
		{name: "missing bundle", jobStatus: jobs.JobStatusCompleted, artifacts: nil, wantErr: ErrBundleMissing},
		{name: "invalid schema", jobStatus: jobs.JobStatusCompleted, artifacts: bundleArtifacts(), bundle: `{"schema_version":1,"program":{"external_id":"p1","title":"Title","description":"Desc","language":"zh-CN","author":"A","unknown":true},"episodes":[]}`, wantErr: ErrBundleInvalid},
		{name: "unregistered artifact", jobStatus: jobs.JobStatusCompleted, artifacts: bundleArtifacts(), bundle: validBundleWithAudio("missing.mp3"), wantErr: ErrBundleInvalid},
		{name: "path escape", jobStatus: jobs.JobStatusCompleted, artifacts: bundleArtifacts(), bundle: validBundleWithAudio("../audio.txt"), wantErr: ErrBundleInvalid},
		{name: "url artifact", jobStatus: jobs.JobStatusCompleted, artifacts: bundleArtifacts(), bundle: validBundleWithAudio("https://example.invalid/audio.mp3"), wantErr: ErrBundleInvalid},
		{name: "secret metadata", jobStatus: jobs.JobStatusCompleted, artifacts: append(bundleArtifacts(), artifact("audio.txt", "audio")), bundle: strings.Replace(validBundleWithAudio("audio.txt"), `"metadata":{}`, `"metadata":{"token":"secret"}`, 1), wantErr: ErrBundleInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rig := newIntakeRig(tt.jobStatus, tt.artifacts, tt.bundle)
			_, err := rig.service.Run(context.Background(), "job1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestIntakeCreatesReviewPendingContentIdempotently(t *testing.T) {
	artifacts := append(bundleArtifacts(), artifact("audio.txt", "audio"))
	rig := newIntakeRig(jobs.JobStatusCompleted, artifacts, validBundleWithAudio("audio.txt"))
	result, err := rig.service.Run(context.Background(), "job1")
	if err != nil {
		t.Fatalf("run intake: %v", err)
	}
	if result.Program == nil || result.Program.Status != content.ProgramStatusReviewPending {
		t.Fatalf("expected review pending program, got %+v", result.Program)
	}
	if len(rig.content.episodes) != 1 {
		t.Fatalf("expected one episode, got %d", len(rig.content.episodes))
	}
	if len(rig.content.reviews) != 2 {
		t.Fatalf("expected program and episode reviews, got %d", len(rig.content.reviews))
	}
	if len(rig.content.events) != 2 {
		t.Fatalf("expected publication events, got %d", len(rig.content.events))
	}
	if len(rig.content.media) != 1 || strings.Contains(rig.content.media[0].StagedStorageKey, "/Users/") {
		t.Fatalf("expected staged media without host path, got %+v", rig.content.media)
	}
	second, err := rig.service.Run(context.Background(), "job1")
	if err != nil {
		t.Fatalf("second intake should be idempotent: %v", err)
	}
	if second.IntakeRun.ID != result.IntakeRun.ID || len(rig.content.programs) != 1 || len(rig.content.episodes) != 1 {
		t.Fatalf("expected idempotent result, got first=%+v second=%+v", result.IntakeRun, second.IntakeRun)
	}
}

func TestIntakeDoesNotMergeSameTitleAcrossSources(t *testing.T) {
	store := newMemoryContentStore()
	for _, sourceID := range []string{"source1", "source2"} {
		rig := newIntakeRig(jobs.JobStatusCompleted, append(bundleArtifacts(), artifact("audio.txt", "audio")), validBundleWithAudio("audio.txt"))
		rig.content = store
		rig.jobs.job.ConnectorSourceID = sourceID
		rig.service = NewService(rig.jobs, store, rig.reader)
		if _, err := rig.service.Run(context.Background(), "job1-"+sourceID); err != nil {
			t.Fatalf("run intake for %s: %v", sourceID, err)
		}
	}
	if len(store.programs) != 2 {
		t.Fatalf("same title across sources must not merge, got %d programs", len(store.programs))
	}
}

type intakeRig struct {
	service *Service
	jobs    *memoryIntakeJobs
	content *memoryContentStore
	reader  memoryArtifactReader
}

func newIntakeRig(status jobs.JobStatus, artifacts []jobs.ImportJobArtifact, bundle string) *intakeRig {
	job := jobs.ImportJob{ID: "job1", ConnectorSourceID: "source1", ConnectorVersionID: "version1", Status: status, CreatedAt: time.Now()}
	jobStore := &memoryIntakeJobs{job: job, artifacts: artifacts}
	reader := memoryArtifactReader{bodyByPath: map[string][]byte{"bundle.json": []byte(bundle)}}
	contentStore := newMemoryContentStore()
	return &intakeRig{jobs: jobStore, content: contentStore, reader: reader, service: NewService(jobStore, contentStore, reader)}
}

func bundleArtifacts() []jobs.ImportJobArtifact {
	return []jobs.ImportJobArtifact{artifact("bundle.json", "metadata_bundle")}
}

func artifact(path string, artifactType string) jobs.ImportJobArtifact {
	return jobs.ImportJobArtifact{ID: path + "-id", ImportJobID: "job1", ArtifactType: artifactType, RelativePath: path, SizeBytes: 12, SHA256: "abc", StorageKey: "job1/" + path, CreatedAt: time.Now()}
}

func validBundleWithAudio(audio string) string {
	return `{"schema_version":1,"program":{"external_id":"program-1","title":"Same Title","description":"Program desc","language":"zh-CN","author":"Author","cover_artifact":""},"episodes":[{"external_id":"ep-1","title":"Episode","description":"Episode desc","published_at":"2026-06-27T00:00:00Z","duration_seconds":120,"audio_artifact":"` + audio + `","cover_artifact":"","metadata":{}}]}`
}

type memoryIntakeJobs struct {
	job       jobs.ImportJob
	artifacts []jobs.ImportJobArtifact
}

func (m *memoryIntakeJobs) GetJob(context.Context, string) (jobs.ImportJob, error) { return m.job, nil }
func (m *memoryIntakeJobs) ListArtifacts(context.Context, string) ([]jobs.ImportJobArtifact, error) {
	return m.artifacts, nil
}

type memoryArtifactReader struct{ bodyByPath map[string][]byte }

func (r memoryArtifactReader) ReadArtifact(_ context.Context, artifact jobs.ImportJobArtifact) ([]byte, error) {
	return r.bodyByPath[artifact.RelativePath], nil
}

type memoryContentStore struct {
	programsByKey map[string]content.Program
	programs      map[string]content.Program
	episodes      map[string]content.Episode
	reviews       []content.ReviewItem
	events        []content.PublicationEvent
	media         []content.MediaAsset
	intakeRuns    map[string]content.IntakeRun
}

func newMemoryContentStore() *memoryContentStore {
	return &memoryContentStore{programsByKey: map[string]content.Program{}, programs: map[string]content.Program{}, episodes: map[string]content.Episode{}, intakeRuns: map[string]content.IntakeRun{}}
}

func (m *memoryContentStore) UpsertProgramFromSource(_ context.Context, in content.UpsertProgramInput) (content.Program, error) {
	key := in.SourceID + ":" + in.ExternalProgramID
	item, ok := m.programsByKey[key]
	if !ok {
		item = content.Program{ID: "program-" + in.SourceID, CanonicalKey: key, CreatedFromSourceID: in.SourceID, CreatedFromJobID: in.JobID, CreatedAt: time.Now()}
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

func (m *memoryContentStore) UpsertEpisode(_ context.Context, in content.UpsertEpisodeInput) (content.Episode, error) {
	key := in.ProgramID + ":" + in.ExternalEpisodeID
	item := m.episodes[key]
	if item.ID == "" {
		item.ID = "episode-" + in.ExternalEpisodeID
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

func (m *memoryContentStore) CreateOrKeepPendingReview(_ context.Context, targetType string, targetID string, reviewKind string, jobID string) (content.ReviewItem, error) {
	for _, item := range m.reviews {
		if item.TargetType == targetType && item.TargetID == targetID && item.ReviewKind == reviewKind && item.Status == content.ReviewStatusPending {
			return item, nil
		}
	}
	item := content.ReviewItem{ID: targetType + "-" + targetID, TargetType: targetType, TargetID: targetID, ReviewKind: reviewKind, Status: content.ReviewStatusPending, RequestedByJobID: &jobID, CreatedAt: time.Now()}
	m.reviews = append(m.reviews, item)
	return item, nil
}

func (m *memoryContentStore) CreateMediaAsset(_ context.Context, in content.CreateMediaAssetInput) (content.MediaAsset, error) {
	item := content.MediaAsset{ID: "media-" + in.ArtifactID, OwnerType: in.OwnerType, OwnerID: in.OwnerID, ImportJobID: in.ImportJobID, ArtifactID: in.ArtifactID, MediaKind: in.MediaKind, StagedStorageKey: in.StagedStorageKey, SizeBytes: in.SizeBytes, SHA256: in.SHA256, Status: content.MediaStatusStaged, CreatedAt: time.Now()}
	m.media = append(m.media, item)
	return item, nil
}
func (m *memoryContentStore) InsertPublicationEvent(_ context.Context, event content.PublicationEvent) error {
	m.events = append(m.events, event)
	return nil
}
func (m *memoryContentStore) GetIntakeRun(_ context.Context, jobID string) (content.IntakeRun, bool, error) {
	item, ok := m.intakeRuns[jobID]
	return item, ok, nil
}
func (m *memoryContentStore) UpsertIntakeRun(_ context.Context, run content.IntakeRun) (content.IntakeRun, error) {
	if existing, ok := m.intakeRuns[run.ImportJobID]; ok && existing.Status == "succeeded" {
		return existing, nil
	}
	m.intakeRuns[run.ImportJobID] = run
	return run, nil
}
func (m *memoryContentStore) ListStagingPrograms(context.Context) ([]content.Program, error) {
	var items []content.Program
	for _, item := range m.programs {
		items = append(items, item)
	}
	return items, nil
}
func (m *memoryContentStore) GetProgram(_ context.Context, id string) (content.Program, bool, error) {
	item, ok := m.programs[id]
	return item, ok, nil
}
func (m *memoryContentStore) ListStagingEpisodes(context.Context) ([]content.Episode, error) {
	var items []content.Episode
	for _, item := range m.episodes {
		items = append(items, item)
	}
	return items, nil
}
func (m *memoryContentStore) GetEpisode(_ context.Context, id string) (content.Episode, bool, error) {
	for _, item := range m.episodes {
		if item.ID == id {
			return item, true, nil
		}
	}
	return content.Episode{}, false, nil
}
func (m *memoryContentStore) ListAdminPrograms(ctx context.Context) ([]content.Program, error) {
	return m.ListStagingPrograms(ctx)
}
func (m *memoryContentStore) GetAdminProgram(ctx context.Context, id string) (content.Program, bool, error) {
	return m.GetProgram(ctx, id)
}
func (m *memoryContentStore) GetAdminEpisode(ctx context.Context, id string) (content.Episode, bool, error) {
	return m.GetEpisode(ctx, id)
}
func (m *memoryContentStore) ListProgramEpisodes(_ context.Context, programID string) ([]content.Episode, error) {
	var items []content.Episode
	for _, item := range m.episodes {
		if item.ProgramID == programID {
			items = append(items, item)
		}
	}
	return items, nil
}
func (m *memoryContentStore) ListReviews(context.Context) ([]content.ReviewItem, error) {
	return m.reviews, nil
}
func (m *memoryContentStore) GetReview(_ context.Context, id string) (content.ReviewItem, bool, error) {
	for _, item := range m.reviews {
		if item.ID == id {
			return item, true, nil
		}
	}
	return content.ReviewItem{}, false, nil
}
func (m *memoryContentStore) SetReviewDecision(_ context.Context, id string, status content.ReviewStatus, actorID string, note string) (content.ReviewItem, error) {
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
func (m *memoryContentStore) SetProgramStatus(ctx context.Context, id string, status content.ProgramStatus) (content.Program, error) {
	item, _, _ := m.GetProgram(ctx, id)
	item.Status = status
	m.programs[id] = item
	return item, nil
}
func (m *memoryContentStore) SetEpisodeStatus(ctx context.Context, id string, status content.EpisodeStatus) (content.Episode, error) {
	item, _, _ := m.GetEpisode(ctx, id)
	item.Status = status
	m.episodes[item.ProgramID+":"+item.ExternalEpisodeID] = item
	return item, nil
}
func (m *memoryContentStore) UpdateProgram(ctx context.Context, id string, in content.UpdateProgramInput) (content.Program, error) {
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
func (m *memoryContentStore) UpdateEpisode(ctx context.Context, id string, in content.UpdateEpisodeInput) (content.Episode, error) {
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
func (m *memoryContentStore) CountPendingReviews(_ context.Context, targetType string, targetID string) (int, error) {
	count := 0
	for _, item := range m.reviews {
		if item.TargetType == targetType && item.TargetID == targetID && item.Status == content.ReviewStatusPending {
			count++
		}
	}
	return count, nil
}
func (m *memoryContentStore) HasApprovedMedia(_ context.Context, episodeID string) (bool, error) {
	for _, item := range m.media {
		if item.OwnerID == episodeID && item.MediaKind == "audio" && item.Status == content.MediaStatusApproved {
			return true, nil
		}
	}
	return false, nil
}
func (m *memoryContentStore) ApproveMediaForEpisode(_ context.Context, episodeID string) error {
	for i, item := range m.media {
		if item.OwnerID == episodeID {
			item.Status = content.MediaStatusApproved
			item.DeliveryStatus = content.MediaStatusApproved
			m.media[i] = item
		}
	}
	return nil
}
func (m *memoryContentStore) PromoteEpisodeMedia(_ context.Context, episodeID string) error {
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
