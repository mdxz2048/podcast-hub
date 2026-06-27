package content

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReviewApproveRejectAndPublishWorkflow(t *testing.T) {
	ctx := context.Background()
	store := newContentMemoryStore()
	service := NewService(store)

	if _, err := service.RejectReview(ctx, ReviewDecisionInput{ReviewID: "review-program", ActorID: "admin"}); !errors.Is(err, ErrReviewReasonRequired) {
		t.Fatalf("expected reason required, got %v", err)
	}
	review, err := service.ApproveReview(ctx, ReviewDecisionInput{ReviewID: "review-program", ActorID: "admin"})
	if err != nil {
		t.Fatalf("approve program review: %v", err)
	}
	if review.Status != ReviewStatusApproved || store.programs["program-1"].Status != ProgramStatusApproved {
		t.Fatalf("expected approved program review/status, got review=%+v program=%+v", review, store.programs["program-1"])
	}
	if _, err := service.PublishProgram(ctx, "program-1", "admin"); err != nil {
		t.Fatalf("publish approved program: %v", err)
	}
	if store.programs["program-1"].Status != ProgramStatusPublished {
		t.Fatalf("expected published program, got %s", store.programs["program-1"].Status)
	}

	if _, err := service.PublishEpisode(ctx, "episode-1", "admin"); !errors.Is(err, ErrPublishPrecondition) {
		t.Fatalf("expected episode publish blocked before review, got %v", err)
	}
	if _, err := service.ApproveReview(ctx, ReviewDecisionInput{ReviewID: "review-episode", ActorID: "admin"}); err != nil {
		t.Fatalf("approve episode review: %v", err)
	}
	if !store.media["episode-1"] {
		t.Fatalf("expected episode media approved by review")
	}
	if _, err := service.PublishEpisode(ctx, "episode-1", "admin"); err != nil {
		t.Fatalf("publish approved episode: %v", err)
	}
	if store.episodes["episode-1"].Status != EpisodeStatusPublished {
		t.Fatalf("expected published episode, got %s", store.episodes["episode-1"].Status)
	}
	if len(store.eventsByType["published"]) != 2 {
		t.Fatalf("expected program and episode publish events, got %+v", store.eventsByType["published"])
	}
}

func TestPublishPreconditionsAndMetadataAudit(t *testing.T) {
	ctx := context.Background()
	store := newContentMemoryStore()
	service := NewService(store)

	if _, err := service.PublishProgram(ctx, "program-1", "admin"); !errors.Is(err, ErrPublishPrecondition) {
		t.Fatalf("expected review_pending program publish blocked, got %v", err)
	}
	store.programs["program-1"] = programWithStatus(ProgramStatusPublished)
	title := "Updated Title"
	if _, err := service.UpdateProgram(ctx, "program-1", UpdateProgramInput{Title: &title, ActorID: "admin"}); err != nil {
		t.Fatalf("update published program: %v", err)
	}
	if store.programs["program-1"].Title != title {
		t.Fatalf("expected updated title")
	}
	if store.pendingReviewCount("program", "program-1") == 0 || len(store.eventsByType["metadata_updated"]) == 0 {
		t.Fatalf("expected metadata update event and pending review")
	}

	store.programs["program-1"] = programWithStatus(ProgramStatusPublished)
	store.episodes["episode-1"] = episodeWithStatus(EpisodeStatusApproved)
	store.media["episode-1"] = false
	if _, err := service.PublishEpisode(ctx, "episode-1", "admin"); !errors.Is(err, ErrPublishPrecondition) {
		t.Fatalf("expected approved media precondition, got %v", err)
	}
	if _, err := service.ArchiveEpisode(ctx, "episode-1", "admin"); err != nil {
		t.Fatalf("archive episode: %v", err)
	}
	if store.episodes["episode-1"].Status != EpisodeStatusArchived || len(store.eventsByType["archived"]) != 1 {
		t.Fatalf("expected archived episode and event")
	}
}

func TestPublishEpisodeRequiresSuccessfulMediaPromotion(t *testing.T) {
	ctx := context.Background()
	store := newContentMemoryStore()
	service := NewService(store)

	store.programs["program-1"] = programWithStatus(ProgramStatusPublished)
	store.episodes["episode-1"] = episodeWithStatus(EpisodeStatusApproved)
	store.media["episode-1"] = true
	approvedBy := "admin"
	approvedAt := time.Now()
	store.reviews["review-episode"] = ReviewItem{ID: "review-episode", TargetType: "episode", TargetID: "episode-1", ReviewKind: "metadata", Status: ReviewStatusApproved, ReviewedBy: &approvedBy, ReviewedAt: &approvedAt, CreatedAt: approvedAt}
	store.promoteErr = errors.New("promote failed")

	if _, err := service.PublishEpisode(ctx, "episode-1", "admin"); err == nil || err.Error() != "promote failed" {
		t.Fatalf("expected promote failure, got %v", err)
	}
	if store.episodes["episode-1"].Status != EpisodeStatusApproved {
		t.Fatalf("expected episode to remain approved, got %s", store.episodes["episode-1"].Status)
	}
	if store.promoteCalls != 1 {
		t.Fatalf("expected one promote call, got %d", store.promoteCalls)
	}
	if len(store.eventsByType["published"]) != 0 {
		t.Fatalf("expected no published event on promote failure")
	}
}

type contentMemoryStore struct {
	programs     map[string]Program
	episodes     map[string]Episode
	reviews      map[string]ReviewItem
	media        map[string]bool
	eventsByType map[string][]PublicationEvent
	promoteErr   error
	promoteCalls int
}

func newContentMemoryStore() *contentMemoryStore {
	return &contentMemoryStore{
		programs: map[string]Program{"program-1": programWithStatus(ProgramStatusReviewPending)},
		episodes: map[string]Episode{"episode-1": episodeWithStatus(EpisodeStatusReviewPending)},
		reviews: map[string]ReviewItem{
			"review-program": {ID: "review-program", TargetType: "program", TargetID: "program-1", ReviewKind: "metadata", Status: ReviewStatusPending, CreatedAt: time.Now()},
			"review-episode": {ID: "review-episode", TargetType: "episode", TargetID: "episode-1", ReviewKind: "metadata", Status: ReviewStatusPending, CreatedAt: time.Now()},
		},
		media:        map[string]bool{"episode-1": false},
		eventsByType: map[string][]PublicationEvent{},
	}
}

func programWithStatus(status ProgramStatus) Program {
	return Program{ID: "program-1", CanonicalKey: "source:p1", Title: "Program", Description: "Desc", Author: "Author", Language: "zh-CN", Status: status, CreatedFromSourceID: "source-1", CreatedFromJobID: "job-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func episodeWithStatus(status EpisodeStatus) Episode {
	return Episode{ID: "episode-1", ProgramID: "program-1", ExternalEpisodeID: "ep-1", Title: "Episode", Description: "Desc", PublishedAt: time.Now(), DurationSeconds: 120, Status: status, SourceJobID: "job-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (s *contentMemoryStore) UpsertProgramFromSource(context.Context, UpsertProgramInput) (Program, error) {
	return Program{}, nil
}
func (s *contentMemoryStore) UpsertEpisode(context.Context, UpsertEpisodeInput) (Episode, error) {
	return Episode{}, nil
}
func (s *contentMemoryStore) CreateOrKeepPendingReview(_ context.Context, targetType string, targetID string, reviewKind string, jobID string) (ReviewItem, error) {
	for _, item := range s.reviews {
		if item.TargetType == targetType && item.TargetID == targetID && item.ReviewKind == reviewKind && item.Status == ReviewStatusPending {
			return item, nil
		}
	}
	item := ReviewItem{ID: "review-" + targetType + "-" + targetID, TargetType: targetType, TargetID: targetID, ReviewKind: reviewKind, Status: ReviewStatusPending, CreatedAt: time.Now()}
	s.reviews[item.ID] = item
	return item, nil
}
func (s *contentMemoryStore) CreateMediaAsset(context.Context, CreateMediaAssetInput) (MediaAsset, error) {
	return MediaAsset{}, nil
}
func (s *contentMemoryStore) InsertPublicationEvent(_ context.Context, event PublicationEvent) error {
	s.eventsByType[event.EventType] = append(s.eventsByType[event.EventType], event)
	return nil
}
func (s *contentMemoryStore) GetIntakeRun(context.Context, string) (IntakeRun, bool, error) {
	return IntakeRun{}, false, nil
}
func (s *contentMemoryStore) UpsertIntakeRun(context.Context, IntakeRun) (IntakeRun, error) {
	return IntakeRun{}, nil
}
func (s *contentMemoryStore) ListStagingPrograms(context.Context) ([]Program, error) { return nil, nil }
func (s *contentMemoryStore) GetProgram(_ context.Context, id string) (Program, bool, error) {
	item, ok := s.programs[id]
	return item, ok, nil
}
func (s *contentMemoryStore) ListStagingEpisodes(context.Context) ([]Episode, error) { return nil, nil }
func (s *contentMemoryStore) GetEpisode(_ context.Context, id string) (Episode, bool, error) {
	item, ok := s.episodes[id]
	return item, ok, nil
}
func (s *contentMemoryStore) ListAdminPrograms(context.Context) ([]Program, error) { return nil, nil }
func (s *contentMemoryStore) GetAdminProgram(ctx context.Context, id string) (Program, bool, error) {
	return s.GetProgram(ctx, id)
}
func (s *contentMemoryStore) GetAdminEpisode(ctx context.Context, id string) (Episode, bool, error) {
	return s.GetEpisode(ctx, id)
}
func (s *contentMemoryStore) ListProgramEpisodes(context.Context, string) ([]Episode, error) {
	return nil, nil
}
func (s *contentMemoryStore) ListReviews(context.Context) ([]ReviewItem, error) {
	items := make([]ReviewItem, 0, len(s.reviews))
	for _, item := range s.reviews {
		items = append(items, item)
	}
	return items, nil
}
func (s *contentMemoryStore) GetReview(_ context.Context, id string) (ReviewItem, bool, error) {
	item, ok := s.reviews[id]
	return item, ok, nil
}
func (s *contentMemoryStore) SetReviewDecision(_ context.Context, id string, status ReviewStatus, actorID string, note string) (ReviewItem, error) {
	item := s.reviews[id]
	item.Status = status
	item.ReviewedBy = &actorID
	item.ReviewNote = note
	now := time.Now()
	item.ReviewedAt = &now
	s.reviews[id] = item
	return item, nil
}
func (s *contentMemoryStore) SetProgramStatus(_ context.Context, id string, status ProgramStatus) (Program, error) {
	item := s.programs[id]
	item.Status = status
	s.programs[id] = item
	return item, nil
}
func (s *contentMemoryStore) SetEpisodeStatus(_ context.Context, id string, status EpisodeStatus) (Episode, error) {
	item := s.episodes[id]
	item.Status = status
	s.episodes[id] = item
	return item, nil
}
func (s *contentMemoryStore) UpdateProgram(_ context.Context, id string, in UpdateProgramInput) (Program, error) {
	item := s.programs[id]
	if in.Title != nil {
		item.Title = *in.Title
	}
	s.programs[id] = item
	return item, nil
}
func (s *contentMemoryStore) UpdateEpisode(_ context.Context, id string, in UpdateEpisodeInput) (Episode, error) {
	item := s.episodes[id]
	if in.Title != nil {
		item.Title = *in.Title
	}
	s.episodes[id] = item
	return item, nil
}
func (s *contentMemoryStore) CountPendingReviews(_ context.Context, targetType string, targetID string) (int, error) {
	return s.pendingReviewCount(targetType, targetID), nil
}
func (s *contentMemoryStore) pendingReviewCount(targetType string, targetID string) int {
	count := 0
	for _, item := range s.reviews {
		if item.TargetType == targetType && item.TargetID == targetID && item.Status == ReviewStatusPending {
			count++
		}
	}
	return count
}
func (s *contentMemoryStore) HasApprovedMedia(_ context.Context, episodeID string) (bool, error) {
	return s.media[episodeID], nil
}
func (s *contentMemoryStore) ApproveMediaForEpisode(_ context.Context, episodeID string) error {
	s.media[episodeID] = true
	return nil
}
func (s *contentMemoryStore) PromoteEpisodeMedia(_ context.Context, episodeID string) error {
	s.promoteCalls++
	return s.promoteErr
}
