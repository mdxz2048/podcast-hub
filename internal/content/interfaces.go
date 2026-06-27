package content

import "context"

type Store interface {
	UpsertProgramFromSource(ctx context.Context, in UpsertProgramInput) (Program, error)
	UpsertEpisode(ctx context.Context, in UpsertEpisodeInput) (Episode, error)
	CreateOrKeepPendingReview(ctx context.Context, targetType string, targetID string, reviewKind string, jobID string) (ReviewItem, error)
	CreateMediaAsset(ctx context.Context, in CreateMediaAssetInput) (MediaAsset, error)
	InsertPublicationEvent(ctx context.Context, event PublicationEvent) error
	GetIntakeRun(ctx context.Context, jobID string) (IntakeRun, bool, error)
	UpsertIntakeRun(ctx context.Context, run IntakeRun) (IntakeRun, error)
	ListStagingPrograms(ctx context.Context) ([]Program, error)
	GetProgram(ctx context.Context, programID string) (Program, bool, error)
	ListStagingEpisodes(ctx context.Context) ([]Episode, error)
	GetEpisode(ctx context.Context, episodeID string) (Episode, bool, error)
	ListAdminPrograms(ctx context.Context) ([]Program, error)
	GetAdminProgram(ctx context.Context, programID string) (Program, bool, error)
	GetAdminEpisode(ctx context.Context, episodeID string) (Episode, bool, error)
	ListProgramEpisodes(ctx context.Context, programID string) ([]Episode, error)
	ListReviews(ctx context.Context) ([]ReviewItem, error)
	GetReview(ctx context.Context, reviewID string) (ReviewItem, bool, error)
	SetReviewDecision(ctx context.Context, reviewID string, status ReviewStatus, actorID string, note string) (ReviewItem, error)
	SetProgramStatus(ctx context.Context, programID string, status ProgramStatus) (Program, error)
	SetEpisodeStatus(ctx context.Context, episodeID string, status EpisodeStatus) (Episode, error)
	UpdateProgram(ctx context.Context, programID string, in UpdateProgramInput) (Program, error)
	UpdateEpisode(ctx context.Context, episodeID string, in UpdateEpisodeInput) (Episode, error)
	CountPendingReviews(ctx context.Context, targetType string, targetID string) (int, error)
	HasApprovedMedia(ctx context.Context, episodeID string) (bool, error)
	ApproveMediaForEpisode(ctx context.Context, episodeID string) error
	PromoteEpisodeMedia(ctx context.Context, episodeID string) error
}
