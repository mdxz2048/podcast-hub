package content

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrReviewReasonRequired = errors.New("review reject reason required")
	ErrInvalidTransition    = errors.New("invalid content state transition")
	ErrPublishPrecondition  = errors.New("publish precondition failed")
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) ListStagingPrograms(ctx context.Context) ([]Program, error) {
	return s.store.ListStagingPrograms(ctx)
}

func (s *Service) GetProgram(ctx context.Context, id string) (Program, error) {
	item, found, err := s.store.GetProgram(ctx, id)
	if err != nil {
		return Program{}, err
	}
	if !found {
		return Program{}, fmt.Errorf("program not found")
	}
	return item, nil
}

func (s *Service) ListStagingEpisodes(ctx context.Context) ([]Episode, error) {
	return s.store.ListStagingEpisodes(ctx)
}

func (s *Service) GetEpisode(ctx context.Context, id string) (Episode, error) {
	item, found, err := s.store.GetEpisode(ctx, id)
	if err != nil {
		return Episode{}, err
	}
	if !found {
		return Episode{}, fmt.Errorf("episode not found")
	}
	return item, nil
}

func (s *Service) ListAdminPrograms(ctx context.Context) ([]Program, error) {
	return s.store.ListAdminPrograms(ctx)
}

func (s *Service) GetAdminProgram(ctx context.Context, id string) (Program, error) {
	item, found, err := s.store.GetAdminProgram(ctx, id)
	if err != nil {
		return Program{}, err
	}
	if !found {
		return Program{}, fmt.Errorf("program not found")
	}
	return item, nil
}

func (s *Service) GetAdminEpisode(ctx context.Context, id string) (Episode, error) {
	item, found, err := s.store.GetAdminEpisode(ctx, id)
	if err != nil {
		return Episode{}, err
	}
	if !found {
		return Episode{}, fmt.Errorf("episode not found")
	}
	return item, nil
}

func (s *Service) ListProgramEpisodes(ctx context.Context, programID string) ([]Episode, error) {
	return s.store.ListProgramEpisodes(ctx, programID)
}

func (s *Service) ListReviews(ctx context.Context) ([]ReviewItem, error) {
	return s.store.ListReviews(ctx)
}

func (s *Service) GetReview(ctx context.Context, reviewID string) (ReviewItem, error) {
	item, found, err := s.store.GetReview(ctx, reviewID)
	if err != nil {
		return ReviewItem{}, err
	}
	if !found {
		return ReviewItem{}, fmt.Errorf("review not found")
	}
	return item, nil
}

func (s *Service) ApproveReview(ctx context.Context, in ReviewDecisionInput) (ReviewItem, error) {
	review, err := s.store.SetReviewDecision(ctx, in.ReviewID, ReviewStatusApproved, in.ActorID, "")
	if err != nil {
		return ReviewItem{}, err
	}
	switch review.TargetType {
	case "program":
		if _, err := s.store.SetProgramStatus(ctx, review.TargetID, ProgramStatusApproved); err != nil {
			return ReviewItem{}, err
		}
	case "episode":
		if _, err := s.store.SetEpisodeStatus(ctx, review.TargetID, EpisodeStatusApproved); err != nil {
			return ReviewItem{}, err
		}
		if err := s.store.ApproveMediaForEpisode(ctx, review.TargetID); err != nil {
			return ReviewItem{}, err
		}
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: review.TargetType, TargetID: review.TargetID, EventType: "approved", ActorID: &in.ActorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	return review, nil
}

func (s *Service) RejectReview(ctx context.Context, in ReviewDecisionInput) (ReviewItem, error) {
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return ReviewItem{}, ErrReviewReasonRequired
	}
	if len([]rune(reason)) > 1000 {
		reason = string([]rune(reason)[:1000])
	}
	review, err := s.store.SetReviewDecision(ctx, in.ReviewID, ReviewStatusRejected, in.ActorID, reason)
	if err != nil {
		return ReviewItem{}, err
	}
	switch review.TargetType {
	case "program":
		if _, err := s.store.SetProgramStatus(ctx, review.TargetID, ProgramStatusRejected); err != nil {
			return ReviewItem{}, err
		}
	case "episode":
		if _, err := s.store.SetEpisodeStatus(ctx, review.TargetID, EpisodeStatusRejected); err != nil {
			return ReviewItem{}, err
		}
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: review.TargetType, TargetID: review.TargetID, EventType: "rejected", ActorID: &in.ActorID, MetadataRedacted: `{"reason":"redacted"}`, CreatedAt: time.Now()})
	return review, nil
}

func (s *Service) UpdateProgram(ctx context.Context, programID string, in UpdateProgramInput) (Program, error) {
	program, err := s.store.UpdateProgram(ctx, programID, in)
	if err != nil {
		return Program{}, err
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: "program", TargetID: program.ID, EventType: "metadata_updated", ActorID: &in.ActorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	if program.Status == ProgramStatusPublished {
		_, _ = s.store.CreateOrKeepPendingReview(ctx, "program", program.ID, "metadata", "")
	}
	return program, nil
}

func (s *Service) UpdateEpisode(ctx context.Context, episodeID string, in UpdateEpisodeInput) (Episode, error) {
	episode, err := s.store.UpdateEpisode(ctx, episodeID, in)
	if err != nil {
		return Episode{}, err
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: "episode", TargetID: episode.ID, EventType: "metadata_updated", ActorID: &in.ActorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	if episode.Status == EpisodeStatusPublished {
		_, _ = s.store.CreateOrKeepPendingReview(ctx, "episode", episode.ID, "metadata", "")
	}
	return episode, nil
}

func (s *Service) SubmitProgramReview(ctx context.Context, programID string) (ReviewItem, error) {
	return s.store.CreateOrKeepPendingReview(ctx, "program", programID, "metadata", "")
}

func (s *Service) SubmitEpisodeReview(ctx context.Context, episodeID string) (ReviewItem, error) {
	return s.store.CreateOrKeepPendingReview(ctx, "episode", episodeID, "metadata", "")
}

func (s *Service) PublishProgram(ctx context.Context, programID string, actorID string) (Program, error) {
	program, found, err := s.store.GetAdminProgram(ctx, programID)
	if err != nil || !found {
		return Program{}, err
	}
	if program.Status != ProgramStatusApproved {
		return Program{}, ErrPublishPrecondition
	}
	pending, err := s.store.CountPendingReviews(ctx, "program", programID)
	if err != nil {
		return Program{}, err
	}
	if pending > 0 || strings.TrimSpace(program.Title) == "" || strings.TrimSpace(program.CreatedFromSourceID) == "" {
		return Program{}, ErrPublishPrecondition
	}
	program, err = s.store.SetProgramStatus(ctx, programID, ProgramStatusPublished)
	if err != nil {
		return Program{}, err
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: "program", TargetID: program.ID, EventType: "published", ActorID: &actorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	return program, nil
}

func (s *Service) PublishEpisode(ctx context.Context, episodeID string, actorID string) (Episode, error) {
	episode, found, err := s.store.GetAdminEpisode(ctx, episodeID)
	if err != nil || !found {
		return Episode{}, err
	}
	if episode.Status != EpisodeStatusApproved {
		return Episode{}, ErrPublishPrecondition
	}
	program, found, err := s.store.GetAdminProgram(ctx, episode.ProgramID)
	if err != nil || !found {
		return Episode{}, err
	}
	if program.Status != ProgramStatusPublished {
		return Episode{}, ErrPublishPrecondition
	}
	pending, err := s.store.CountPendingReviews(ctx, "episode", episodeID)
	if err != nil {
		return Episode{}, err
	}
	if pending > 0 {
		return Episode{}, ErrPublishPrecondition
	}
	mediaOK, err := s.store.HasApprovedMedia(ctx, episodeID)
	if err != nil {
		return Episode{}, err
	}
	if !mediaOK {
		return Episode{}, ErrPublishPrecondition
	}
	episode, err = s.store.SetEpisodeStatus(ctx, episodeID, EpisodeStatusPublished)
	if err != nil {
		return Episode{}, err
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: "episode", TargetID: episode.ID, EventType: "published", ActorID: &actorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	return episode, nil
}

func (s *Service) ArchiveProgram(ctx context.Context, programID string, actorID string) (Program, error) {
	program, err := s.store.SetProgramStatus(ctx, programID, ProgramStatusArchived)
	if err != nil {
		return Program{}, err
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: "program", TargetID: program.ID, EventType: "archived", ActorID: &actorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	return program, nil
}

func (s *Service) ArchiveEpisode(ctx context.Context, episodeID string, actorID string) (Episode, error) {
	episode, err := s.store.SetEpisodeStatus(ctx, episodeID, EpisodeStatusArchived)
	if err != nil {
		return Episode{}, err
	}
	_ = s.store.InsertPublicationEvent(ctx, PublicationEvent{ID: uuid.NewString(), TargetType: "episode", TargetID: episode.ID, EventType: "archived", ActorID: &actorID, MetadataRedacted: `{}`, CreatedAt: time.Now()})
	return episode, nil
}
