package publication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/media"
	"github.com/mdxz2048/podcast-hub/internal/security"
	"github.com/mdxz2048/podcast-hub/internal/users"
)

type Service struct {
	store       Store
	audit       AuditLogger
	mediaStore  *media.LocalStore
	tokenPepper string
}

func NewService(store Store, audit AuditLogger, mediaStore *media.LocalStore, tokenPepper string) *Service {
	return &Service{store: store, audit: audit, mediaStore: mediaStore, tokenPepper: tokenPepper}
}

func (s *Service) GrantProgramAccess(ctx context.Context, programID string, email string, actorID string, reason string) (ProgramAccessGrant, error) {
	normalized, err := users.NormalizeEmail(email)
	if err != nil {
		return ProgramAccessGrant{}, err
	}
	user, found, err := s.store.FindUserByEmail(ctx, normalized)
	if err != nil {
		return ProgramAccessGrant{}, err
	}
	if !found || user.Status != auth.StatusActive {
		return ProgramAccessGrant{}, ErrUserNotEligible
	}
	now := time.Now()
	grant, err := s.store.GrantProgramAccess(ctx, ProgramAccessGrant{ID: uuid.NewString(), UserID: user.ID, ProgramID: programID, Status: ProgramAccessActive, Reason: strings.TrimSpace(reason), CreatedAt: now, UpdatedAt: now}, actorID)
	if err != nil {
		return ProgramAccessGrant{}, err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &actorID, TargetUserID: &user.ID, EventType: "publication.program_access_granted", Result: "success", Metadata: `{"program_id":"` + programID + `"}`})
	return grant, nil
}

func (s *Service) ListProgramAccessGrants(ctx context.Context, programID string) ([]ProgramAccessGrant, error) {
	return s.store.ListProgramAccessGrants(ctx, programID)
}

func (s *Service) RevokeProgramAccess(ctx context.Context, grantID string, actorID string, reason string) (ProgramAccessGrant, error) {
	grant, err := s.store.RevokeProgramAccess(ctx, grantID, actorID, strings.TrimSpace(reason), time.Now())
	if err != nil {
		return ProgramAccessGrant{}, err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &actorID, TargetUserID: &grant.UserID, EventType: "publication.program_access_revoked", Result: "success", Metadata: `{"program_id":"` + grant.ProgramID + `"}`})
	return grant, nil
}

func (s *Service) ListAuthorizedPrograms(ctx context.Context, userID string) ([]UserProgram, error) {
	return s.store.ListAuthorizedPrograms(ctx, userID)
}

func (s *Service) GetAuthorizedProgram(ctx context.Context, userID string, programID string) (UserProgram, error) {
	program, found, err := s.store.GetAuthorizedProgram(ctx, userID, programID)
	if err != nil {
		return UserProgram{}, err
	}
	if !found {
		return UserProgram{}, ErrProgramNotAvailable
	}
	return program, nil
}

func (s *Service) ListAuthorizedEpisodes(ctx context.Context, userID string, programID string) ([]UserEpisode, error) {
	if _, err := s.GetAuthorizedProgram(ctx, userID, programID); err != nil {
		return nil, err
	}
	return s.store.ListAuthorizedEpisodes(ctx, userID, programID)
}

func (s *Service) GetAuthorizedEpisode(ctx context.Context, userID string, episodeID string) (UserEpisode, error) {
	episode, found, err := s.store.GetAuthorizedEpisode(ctx, userID, episodeID)
	if err != nil {
		return UserEpisode{}, err
	}
	if !found {
		return UserEpisode{}, ErrProgramNotAvailable
	}
	return episode, nil
}

func (s *Service) ListUserCollections(ctx context.Context, userID string) ([]UserCollection, error) {
	return s.store.ListUserCollections(ctx, userID)
}

func (s *Service) CreateUserCollection(ctx context.Context, userID string, title string, description string) (UserCollection, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" || len([]rune(title)) > 120 || len([]rune(description)) > 1000 {
		return UserCollection{}, ErrInvalidCollection
	}
	now := time.Now()
	return s.store.CreateUserCollection(ctx, UserCollection{ID: uuid.NewString(), UserID: userID, Title: title, Description: description, CreatedAt: now, UpdatedAt: now})
}

func (s *Service) UpdateUserCollection(ctx context.Context, userID string, collectionID string, title *string, description *string) (UserCollection, error) {
	if title != nil {
		next := strings.TrimSpace(*title)
		if next == "" || len([]rune(next)) > 120 {
			return UserCollection{}, ErrInvalidCollection
		}
		*title = next
	}
	if description != nil {
		next := strings.TrimSpace(*description)
		if len([]rune(next)) > 1000 {
			return UserCollection{}, ErrInvalidCollection
		}
		*description = next
	}
	return s.store.UpdateUserCollection(ctx, userID, collectionID, title, description, time.Now())
}

func (s *Service) DeleteUserCollection(ctx context.Context, userID string, collectionID string) error {
	return s.store.DeleteUserCollection(ctx, userID, collectionID)
}

func (s *Service) AddProgramToCollection(ctx context.Context, userID string, collectionID string, programID string) (UserCollection, error) {
	if _, err := s.GetAuthorizedProgram(ctx, userID, programID); err != nil {
		return UserCollection{}, err
	}
	if err := s.store.AddProgramToCollection(ctx, userID, collectionID, programID, time.Now()); err != nil {
		return UserCollection{}, err
	}
	return s.getUserCollection(ctx, userID, collectionID)
}

func (s *Service) RemoveProgramFromCollection(ctx context.Context, userID string, collectionID string, programID string) (UserCollection, error) {
	if err := s.store.RemoveProgramFromCollection(ctx, userID, collectionID, programID); err != nil {
		return UserCollection{}, err
	}
	return s.getUserCollection(ctx, userID, collectionID)
}

func (s *Service) getUserCollection(ctx context.Context, userID string, collectionID string) (UserCollection, error) {
	collections, err := s.store.ListUserCollections(ctx, userID)
	if err != nil {
		return UserCollection{}, err
	}
	for _, collection := range collections {
		if collection.ID == collectionID {
			return collection, nil
		}
	}
	return UserCollection{}, ErrCollectionNotFound
}

func (s *Service) ListUserFeeds(ctx context.Context, userID string) ([]RSSFeed, error) {
	return s.store.ListRSSFeedsByUser(ctx, userID)
}

func (s *Service) CreateFeed(ctx context.Context, userID string, name string, baseURL string, expiresAt *time.Time) (FeedTokenResult, error) {
	feedName := strings.TrimSpace(name)
	if feedName == "" {
		feedName = "My private feed"
	}
	if len([]rune(feedName)) > 120 {
		return FeedTokenResult{}, ErrInvalidFeedName
	}
	rawToken, err := security.NewOpaqueToken()
	if err != nil {
		return FeedTokenResult{}, err
	}
	now := time.Now()
	feed, err := s.store.CreateRSSFeed(ctx, RSSFeed{ID: uuid.NewString(), UserID: userID, Name: feedName, TokenPrefix: tokenPrefix(rawToken), Status: FeedStatusActive, CreatedAt: now, ExpiresAt: expiresAt}, security.HashWithPepper(rawToken, s.tokenPepper))
	if err != nil {
		return FeedTokenResult{}, err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &userID, TargetUserID: &userID, EventType: "publication.rss_feed_created", Result: "success", Metadata: `{"feed_id":"` + feed.ID + `","token_prefix":"` + feed.TokenPrefix + `"}`})
	return FeedTokenResult{Feed: feed, Token: rawToken, FeedURL: feedURL(baseURL, rawToken)}, nil
}

func (s *Service) RotateFeed(ctx context.Context, userID string, feedID string, baseURL string, expiresAt *time.Time) (FeedTokenResult, error) {
	rawToken, err := security.NewOpaqueToken()
	if err != nil {
		return FeedTokenResult{}, err
	}
	now := time.Now()
	feed, err := s.store.RotateRSSFeed(ctx, feedID, userID, security.HashWithPepper(rawToken, s.tokenPepper), tokenPrefix(rawToken), now, expiresAt)
	if err != nil {
		return FeedTokenResult{}, err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &userID, TargetUserID: &userID, EventType: "publication.rss_feed_rotated", Result: "success", Metadata: `{"feed_id":"` + feed.ID + `","token_prefix":"` + feed.TokenPrefix + `"}`})
	return FeedTokenResult{Feed: feed, Token: rawToken, FeedURL: feedURL(baseURL, rawToken)}, nil
}

func (s *Service) RevokeFeed(ctx context.Context, userID string, feedID string) (RSSFeed, error) {
	feed, found, err := s.store.GetRSSFeedForUser(ctx, feedID, userID)
	if err != nil {
		return RSSFeed{}, err
	}
	if !found {
		return RSSFeed{}, ErrFeedForbidden
	}
	revoked, err := s.store.RevokeRSSFeed(ctx, feed.ID, time.Now())
	if err != nil {
		return RSSFeed{}, err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &userID, TargetUserID: &userID, EventType: "publication.rss_feed_revoked", Result: "success", Metadata: `{"feed_id":"` + feed.ID + `"}`})
	return revoked, nil
}

func (s *Service) DeleteFeed(ctx context.Context, userID string, feedID string) error {
	if err := s.store.DeleteRSSFeed(ctx, feedID, userID); err != nil {
		return err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &userID, TargetUserID: &userID, EventType: "publication.rss_feed_deleted", Result: "success", Metadata: `{"feed_id":"` + feedID + `"}`})
	return nil
}

func (s *Service) ListAdminFeeds(ctx context.Context) ([]AdminRSSFeed, error) {
	return s.store.ListAdminRSSFeeds(ctx)
}

func (s *Service) AdminRevokeFeed(ctx context.Context, feedID string, actorID string) (RSSFeed, error) {
	feed, found, err := s.store.GetRSSFeed(ctx, feedID)
	if err != nil {
		return RSSFeed{}, err
	}
	if !found {
		return RSSFeed{}, ErrFeedNotFound
	}
	revoked, err := s.store.RevokeRSSFeed(ctx, feedID, time.Now())
	if err != nil {
		return RSSFeed{}, err
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{ActorUserID: &actorID, TargetUserID: &feed.UserID, EventType: "publication.rss_feed_admin_revoked", Result: "success", Metadata: `{"feed_id":"` + feedID + `"}`})
	return revoked, nil
}

func (s *Service) ResolveUserMedia(ctx context.Context, userID string, episodeID string) (AuthorizedMedia, error) {
	asset, found, err := s.store.GetAuthorizedMediaForUser(ctx, userID, episodeID)
	if err != nil {
		return AuthorizedMedia{}, err
	}
	if !found {
		return AuthorizedMedia{}, ErrMediaNotAvailable
	}
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{TargetUserID: &userID, EventType: "publication.media_accessed", Result: "success", Metadata: `{"episode_id":"` + episodeID + `"}`})
	return asset, nil
}

func (s *Service) ResolveFeedMedia(ctx context.Context, token string, episodeID string) (AuthorizedMedia, error) {
	asset, found, err := s.store.GetAuthorizedMediaForFeed(ctx, security.HashWithPepper(token, s.tokenPepper), episodeID, time.Now())
	if err != nil {
		return AuthorizedMedia{}, err
	}
	if !found {
		return AuthorizedMedia{}, ErrMediaNotAvailable
	}
	if asset.FeedID != nil {
		_ = s.store.TouchRSSFeed(ctx, *asset.FeedID, time.Now())
	}
	if asset.FeedOwnerUserID != nil {
		_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{TargetUserID: asset.FeedOwnerUserID, EventType: "publication.feed_media_accessed", Result: "success", Metadata: `{"episode_id":"` + episodeID + `","token":"` + RedactOpaqueToken(token) + `"}`})
	}
	return asset, nil
}

func (s *Service) OpenPublishedMedia(ctx context.Context, publishedKey string) (*os.File, error) {
	if s.mediaStore == nil {
		return nil, fmt.Errorf("media store is not configured")
	}
	file, err := s.mediaStore.OpenPublished(ctx, publishedKey)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *Service) BuildPrivateFeed(ctx context.Context, token string, baseURL string) (FeedDocument, []byte, error) {
	feed, user, items, err := s.store.ListAuthorizedFeedEpisodes(ctx, security.HashWithPepper(token, s.tokenPepper), time.Now())
	if err != nil {
		return FeedDocument{}, nil, err
	}
	if feed.Status == FeedStatusRevoked {
		return FeedDocument{}, nil, ErrFeedRevoked
	}
	if feed.Status == FeedStatusExpired {
		return FeedDocument{}, nil, ErrFeedExpired
	}
	updatedAt := feed.CreatedAt
	for _, item := range items {
		if item.UpdatedAt.After(updatedAt) {
			updatedAt = item.UpdatedAt
		}
	}
	doc := FeedDocument{Feed: feed, UserID: user.ID, Items: items, UpdatedAt: updatedAt, ETag: buildFeedETag(feed, items, updatedAt)}
	body, err := marshalRSS(doc, token, baseURL)
	if err != nil {
		return FeedDocument{}, nil, err
	}
	_ = s.store.TouchRSSFeed(ctx, feed.ID, time.Now())
	_ = s.audit.InsertAuditLog(ctx, auth.AuditEvent{TargetUserID: &user.ID, EventType: "publication.rss_feed_accessed", Result: "success", Metadata: `{"feed_id":"` + feed.ID + `","item_count":` + fmt.Sprintf("%d", len(items)) + `,"token":"` + RedactOpaqueToken(token) + `"}`})
	return doc, body, nil
}

func RedactOpaqueToken(token string) string {
	if len(token) <= 8 {
		return "[redacted]"
	}
	return token[:4] + "...[redacted]"
}

func tokenPrefix(rawToken string) string {
	if len(rawToken) >= 8 {
		return rawToken[:8]
	}
	return rawToken
}

func feedURL(baseURL string, rawToken string) string {
	return strings.TrimRight(baseURL, "/") + "/rss/private/" + rawToken + ".xml"
}

func buildFeedETag(feed RSSFeed, items []FeedEpisode, updatedAt time.Time) string {
	hasher := sha256.New()
	hasher.Write([]byte(feed.ID))
	hasher.Write([]byte(feed.Status))
	hasher.Write([]byte(updatedAt.UTC().Format(time.RFC3339Nano)))
	for _, item := range items {
		hasher.Write([]byte(item.EpisodeID))
		hasher.Write([]byte(item.SHA256))
	}
	return `"` + hex.EncodeToString(hasher.Sum(nil)) + `"`
}

type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language,omitempty"`
	Author        string    `xml:"itunes:author,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Items         []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	GUID        string       `xml:"guid"`
	PubDate     string       `xml:"pubDate"`
	Author      string       `xml:"author,omitempty"`
	Enclosure   rssEnclosure `xml:"enclosure"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func marshalRSS(doc FeedDocument, token string, baseURL string) ([]byte, error) {
	channel := rssChannel{
		Title:         doc.Feed.Name,
		Description:   "Private Podcast Hub feed",
		LastBuildDate: doc.UpdatedAt.UTC().Format(time.RFC1123Z),
	}
	if len(doc.Items) > 0 {
		channel.Language = doc.Items[0].Language
		channel.Author = doc.Items[0].ProgramAuthor
	}
	for _, item := range doc.Items {
		channel.Items = append(channel.Items, rssItem{
			Title:       item.Title,
			Description: item.Description,
			GUID:        item.EpisodeID,
			PubDate:     item.PublishedAt.UTC().Format(time.RFC1123Z),
			Author:      item.ProgramAuthor,
			Enclosure: rssEnclosure{
				URL:    strings.TrimRight(baseURL, "/") + "/rss/private/" + token + "/episodes/" + item.EpisodeID + "/media",
				Length: item.SizeBytes,
				Type:   item.ContentType,
			},
		})
	}
	root := rssRoot{Version: "2.0", Channel: channel}
	body, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}
