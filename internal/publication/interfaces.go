package publication

import (
	"context"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/auth"
)

type Store interface {
	FindUserByEmail(ctx context.Context, email string) (auth.User, bool, error)
	GrantProgramAccess(ctx context.Context, grant ProgramAccessGrant, actorID string) (ProgramAccessGrant, error)
	ListProgramAccessGrants(ctx context.Context, programID string) ([]ProgramAccessGrant, error)
	RevokeProgramAccess(ctx context.Context, grantID string, actorID string, reason string, revokedAt time.Time) (ProgramAccessGrant, error)
	ListAuthorizedPrograms(ctx context.Context, userID string) ([]UserProgram, error)
	GetAuthorizedProgram(ctx context.Context, userID string, programID string) (UserProgram, bool, error)
	ListAuthorizedEpisodes(ctx context.Context, userID string, programID string) ([]UserEpisode, error)
	GetAuthorizedEpisode(ctx context.Context, userID string, episodeID string) (UserEpisode, bool, error)
	ListUserCollections(ctx context.Context, userID string) ([]UserCollection, error)
	CreateUserCollection(ctx context.Context, collection UserCollection) (UserCollection, error)
	UpdateUserCollection(ctx context.Context, userID string, collectionID string, title *string, description *string, updatedAt time.Time) (UserCollection, error)
	DeleteUserCollection(ctx context.Context, userID string, collectionID string) error
	AddProgramToCollection(ctx context.Context, userID string, collectionID string, programID string, addedAt time.Time) error
	RemoveProgramFromCollection(ctx context.Context, userID string, collectionID string, programID string) error
	GetRSSFeed(ctx context.Context, feedID string) (RSSFeed, bool, error)
	GetRSSFeedForUser(ctx context.Context, feedID string, userID string) (RSSFeed, bool, error)
	ListRSSFeedsByUser(ctx context.Context, userID string) ([]RSSFeed, error)
	ListAdminRSSFeeds(ctx context.Context) ([]AdminRSSFeed, error)
	CreateRSSFeed(ctx context.Context, feed RSSFeed, tokenHash string) (RSSFeed, error)
	RotateRSSFeed(ctx context.Context, feedID string, userID string, tokenHash string, tokenPrefix string, rotatedAt time.Time, expiresAt *time.Time) (RSSFeed, error)
	RevokeRSSFeed(ctx context.Context, feedID string, revokedAt time.Time) (RSSFeed, error)
	DeleteRSSFeed(ctx context.Context, feedID string, userID string) error
	GetRSSFeedByTokenHash(ctx context.Context, tokenHash string, now time.Time) (RSSFeed, auth.User, bool, error)
	TouchRSSFeed(ctx context.Context, feedID string, usedAt time.Time) error
	GetAuthorizedMediaForUser(ctx context.Context, userID string, episodeID string) (AuthorizedMedia, bool, error)
	GetAuthorizedMediaForFeed(ctx context.Context, tokenHash string, episodeID string, now time.Time) (AuthorizedMedia, bool, error)
	ListAuthorizedFeedEpisodes(ctx context.Context, tokenHash string, now time.Time) (RSSFeed, auth.User, []FeedEpisode, error)
}

type AuditLogger interface {
	InsertAuditLog(ctx context.Context, event auth.AuditEvent) error
}
