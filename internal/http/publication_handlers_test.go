package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/media"
	"github.com/mdxz2048/podcast-hub/internal/publication"
	"github.com/mdxz2048/podcast-hub/internal/security"
)

func TestUserMediaRequiresAuthorization(t *testing.T) {
	server, state := newPublicationTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/media/episodes/episode-1", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	state.sessionUser = state.users["user-2"]
	req = httptest.NewRequest(http.MethodGet, "/media/episodes/episode-1", nil)
	req.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unauthorized access, got %d", rec.Code)
	}
	assertNoSensitiveLeak(t, rec.Body.String())
}

func TestUserMediaSupportsHeadRangeETagAndRevocation(t *testing.T) {
	server, state := newPublicationTestServer(t)
	state.sessionUser = state.users["user-1"]

	headReq := httptest.NewRequest(http.MethodHead, "/media/episodes/episode-1", nil)
	headReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	headRec := httptest.NewRecorder()
	server.Router().ServeHTTP(headRec, headReq)
	if headRec.Code != http.StatusOK {
		t.Fatalf("expected HEAD 200, got %d", headRec.Code)
	}
	if headRec.Header().Get("Content-Type") != "audio/mpeg" {
		t.Fatalf("expected audio content type, got %q", headRec.Header().Get("Content-Type"))
	}
	if headRec.Header().Get("Accept-Ranges") != "bytes" {
		t.Fatalf("expected range support, got %q", headRec.Header().Get("Accept-Ranges"))
	}
	etag := headRec.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected etag header")
	}

	rangeReq := httptest.NewRequest(http.MethodGet, "/media/episodes/episode-1", nil)
	rangeReq.Header.Set("Range", "bytes=0-3")
	rangeReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rangeRec := httptest.NewRecorder()
	server.Router().ServeHTTP(rangeRec, rangeReq)
	if rangeRec.Code != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d", rangeRec.Code)
	}
	if rangeRec.Body.String() != "FAKE" {
		t.Fatalf("expected partial media body, got %q", rangeRec.Body.String())
	}

	invalidRangeReq := httptest.NewRequest(http.MethodGet, "/media/episodes/episode-1", nil)
	invalidRangeReq.Header.Set("Range", "bytes=999-1000")
	invalidRangeReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	invalidRangeRec := httptest.NewRecorder()
	server.Router().ServeHTTP(invalidRangeRec, invalidRangeReq)
	if invalidRangeRec.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("expected 416, got %d", invalidRangeRec.Code)
	}
	assertNoSensitiveLeak(t, invalidRangeRec.Body.String())

	cacheReq := httptest.NewRequest(http.MethodGet, "/media/episodes/episode-1", nil)
	cacheReq.Header.Set("If-None-Match", etag)
	cacheReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	cacheRec := httptest.NewRecorder()
	server.Router().ServeHTTP(cacheRec, cacheReq)
	if cacheRec.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", cacheRec.Code)
	}

	state.revokeGrant("grant-1")
	revokedReq := httptest.NewRequest(http.MethodGet, "/media/episodes/episode-1", nil)
	revokedReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	revokedRec := httptest.NewRecorder()
	server.Router().ServeHTTP(revokedRec, revokedReq)
	if revokedRec.Code != http.StatusNotFound {
		t.Fatalf("expected revoked access to fail with 404, got %d", revokedRec.Code)
	}
}

func TestPrivateRSSFeedAndEnclosureSecurity(t *testing.T) {
	server, state := newPublicationTestServer(t)

	invalidReq := httptest.NewRequest(http.MethodGet, "/rss/private/not-a-token.xml", nil)
	invalidRec := httptest.NewRecorder()
	server.Router().ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusNotFound {
		t.Fatalf("expected invalid token 404, got %d", invalidRec.Code)
	}

	feedReq := httptest.NewRequest(http.MethodGet, "/rss/private/token-active.xml", nil)
	feedRec := httptest.NewRecorder()
	server.Router().ServeHTTP(feedRec, feedReq)
	if feedRec.Code != http.StatusOK {
		t.Fatalf("expected feed 200, got %d body=%s", feedRec.Code, feedRec.Body.String())
	}
	body := feedRec.Body.String()
	if !strings.Contains(body, "&amp;") {
		t.Fatalf("expected xml escaping, got %s", body)
	}
	if strings.Contains(strings.ToLower(body), "user@example.invalid") || strings.Contains(body, state.mediaByEpisode["episode-1"].PublishedKey) {
		t.Fatalf("feed leaked private data: %s", body)
	}
	if feedRec.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("expected no-referrer policy")
	}
	if feedRec.Header().Get("X-Robots-Tag") != "noindex, nofollow, noarchive" {
		t.Fatalf("expected robots tag")
	}

	oldETag := feedRec.Header().Get("ETag")
	cacheReq := httptest.NewRequest(http.MethodGet, "/rss/private/token-active.xml", nil)
	cacheReq.Header.Set("If-None-Match", oldETag)
	cacheRec := httptest.NewRecorder()
	server.Router().ServeHTTP(cacheRec, cacheReq)
	if cacheRec.Code != http.StatusNotModified {
		t.Fatalf("expected 304 for cached feed, got %d", cacheRec.Code)
	}

	mediaReq := httptest.NewRequest(http.MethodGet, "/rss/private/token-active/episodes/episode-1/media", nil)
	mediaReq.Header.Set("Range", "bytes=5-8")
	mediaRec := httptest.NewRecorder()
	server.Router().ServeHTTP(mediaRec, mediaReq)
	if mediaRec.Code != http.StatusPartialContent {
		t.Fatalf("expected enclosure 206, got %d", mediaRec.Code)
	}
	if mediaRec.Body.String() != "EDIA" {
		t.Fatalf("expected partial enclosure body, got %q", mediaRec.Body.String())
	}

	state.feedByToken[hashToken("token-active", "pepper")] = publication.RSSFeed{ID: "feed-1", UserID: "user-1", Name: "My Feed", TokenPrefix: "token-ac", Status: publication.FeedStatusRevoked, CreatedAt: time.Now()}
	revokedReq := httptest.NewRequest(http.MethodGet, "/rss/private/token-active.xml", nil)
	revokedRec := httptest.NewRecorder()
	server.Router().ServeHTTP(revokedRec, revokedReq)
	if revokedRec.Code != http.StatusNotFound {
		t.Fatalf("expected revoked token 404, got %d", revokedRec.Code)
	}
}

func TestRSSManagementAPIsReturnOneTimeTokenAndEnforceOwnership(t *testing.T) {
	server, state := newPublicationTestServer(t)
	state.sessionUser = state.users["user-1"]

	createReq := httptest.NewRequest(http.MethodPost, "/me/rss-feeds", strings.NewReader(`{"name":"Roadmap Feed"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Origin", "http://127.0.0.1:5173")
	createReq.Header.Set("X-CSRF-Token", "csrf-token")
	createReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	createReq.AddCookie(&http.Cookie{Name: "podcast_hub_csrf", Value: "csrf-token"})
	createRec := httptest.NewRecorder()
	server.Router().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		Feed  publication.RSSFeed `json:"feed"`
		Token string              `json:"token"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Token == "" {
		t.Fatalf("expected one-time token in create response")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/me/rss-feeds", nil)
	listReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	listRec := httptest.NewRecorder()
	server.Router().ServeHTTP(listRec, listReq)
	if strings.Contains(listRec.Body.String(), created.Token) {
		t.Fatalf("list response must not expose plaintext token")
	}

	rotateReq := httptest.NewRequest(http.MethodPost, "/me/rss-feeds/feed-1/rotate", strings.NewReader(`{}`))
	rotateReq.Header.Set("Content-Type", "application/json")
	rotateReq.Header.Set("Origin", "http://127.0.0.1:5173")
	rotateReq.Header.Set("X-CSRF-Token", "csrf-token")
	rotateReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rotateReq.AddCookie(&http.Cookie{Name: "podcast_hub_csrf", Value: "csrf-token"})
	rotateRec := httptest.NewRecorder()
	server.Router().ServeHTTP(rotateRec, rotateReq)
	if rotateRec.Code != http.StatusOK {
		t.Fatalf("expected rotate 200, got %d body=%s", rotateRec.Code, rotateRec.Body.String())
	}
	var rotated struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(rotateRec.Body.Bytes(), &rotated)
	if rotated.Token == "" {
		t.Fatalf("expected rotated token")
	}

	oldFeedReq := httptest.NewRequest(http.MethodGet, "/rss/private/token-active.xml", nil)
	oldFeedRec := httptest.NewRecorder()
	server.Router().ServeHTTP(oldFeedRec, oldFeedReq)
	if oldFeedRec.Code != http.StatusNotFound {
		t.Fatalf("expected old token invalid after rotate, got %d", oldFeedRec.Code)
	}

	state.sessionUser = state.users["user-2"]
	forbiddenReq := httptest.NewRequest(http.MethodPost, "/me/rss-feeds/feed-1/revoke", strings.NewReader(`{}`))
	forbiddenReq.Header.Set("Content-Type", "application/json")
	forbiddenReq.Header.Set("Origin", "http://127.0.0.1:5173")
	forbiddenReq.Header.Set("X-CSRF-Token", "csrf-token")
	forbiddenReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	forbiddenReq.AddCookie(&http.Cookie{Name: "podcast_hub_csrf", Value: "csrf-token"})
	forbiddenRec := httptest.NewRecorder()
	server.Router().ServeHTTP(forbiddenRec, forbiddenReq)
	if forbiddenRec.Code != http.StatusNotFound {
		t.Fatalf("expected ownership check to return 404, got %d", forbiddenRec.Code)
	}
}

func TestAdminRSSAPIsRequireAdminAndCanRevoke(t *testing.T) {
	server, state := newPublicationTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/rss-feeds", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	state.sessionUser = state.users["user-1"]
	req = httptest.NewRequest(http.MethodGet, "/admin/rss-feeds", nil)
	req.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	state.sessionUser = state.users["admin-1"]
	revokeReq := httptest.NewRequest(http.MethodPost, "/admin/rss-feeds/feed-1/revoke", strings.NewReader(`{}`))
	revokeReq.Header.Set("Content-Type", "application/json")
	revokeReq.Header.Set("Origin", "http://127.0.0.1:5173")
	revokeReq.Header.Set("X-CSRF-Token", "csrf-token")
	revokeReq.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	revokeReq.AddCookie(&http.Cookie{Name: "podcast_hub_csrf", Value: "csrf-token"})
	revokeRec := httptest.NewRecorder()
	server.Router().ServeHTTP(revokeRec, revokeReq)
	if revokeRec.Code != http.StatusOK {
		t.Fatalf("expected admin revoke 200, got %d body=%s", revokeRec.Code, revokeRec.Body.String())
	}

	feedReq := httptest.NewRequest(http.MethodGet, "/rss/private/token-active.xml", nil)
	feedRec := httptest.NewRecorder()
	server.Router().ServeHTTP(feedRec, feedReq)
	if feedRec.Code != http.StatusNotFound {
		t.Fatalf("expected revoked feed to fail, got %d", feedRec.Code)
	}
}

type publicationTestState struct {
	users          map[string]auth.User
	sessionUser    auth.User
	grants         map[string]publication.ProgramAccessGrant
	feeds          map[string]publication.RSSFeed
	feedByToken    map[string]publication.RSSFeed
	feedEpisodes   map[string][]publication.FeedEpisode
	mediaByEpisode map[string]publication.AuthorizedMedia
	auditEvents    []auth.AuditEvent
	pepper         string
}

func newPublicationTestServer(t *testing.T) (*Server, *publicationTestState) {
	t.Helper()
	stagingRoot := t.TempDir()
	publishedRoot := t.TempDir()
	mediaStore := media.NewLocalStore(stagingRoot, publishedRoot)
	publishedKey := filepath.ToSlash(filepath.Join("episodes", "episode-1", "media-1.bin"))
	publishedPath := filepath.Join(publishedRoot, filepath.FromSlash(publishedKey))
	if err := os.MkdirAll(filepath.Dir(publishedPath), 0o755); err != nil {
		t.Fatalf("mkdir media dir: %v", err)
	}
	if err := os.WriteFile(publishedPath, []byte("FAKEMEDIA-CONTENT"), 0o644); err != nil {
		t.Fatalf("write media: %v", err)
	}
	state := &publicationTestState{
		pepper: "pepper",
		users: map[string]auth.User{
			"user-1":  {ID: "user-1", Email: "user@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive},
			"user-2":  {ID: "user-2", Email: "other@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive},
			"admin-1": {ID: "admin-1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive},
		},
		grants: map[string]publication.ProgramAccessGrant{
			"grant-1": {ID: "grant-1", UserID: "user-1", ProgramID: "program-1", Status: publication.ProgramAccessActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		feeds: map[string]publication.RSSFeed{
			"feed-1": {ID: "feed-1", UserID: "user-1", Name: "My Feed", TokenPrefix: "token-ac", Status: publication.FeedStatusActive, CreatedAt: time.Now()},
		},
		feedByToken:  map[string]publication.RSSFeed{},
		feedEpisodes: map[string][]publication.FeedEpisode{},
		mediaByEpisode: map[string]publication.AuthorizedMedia{
			"episode-1": {EpisodeID: "episode-1", ProgramID: "program-1", ContentType: "audio/mpeg", SizeBytes: int64(len("FAKEMEDIA-CONTENT")), SHA256: "sha256-media-1", PublishedAt: time.Now().Add(-time.Hour), PublishedKey: publishedKey, EpisodeTitle: "Hello & Goodbye", ProgramTitle: "Program Title"},
		},
	}
	tokenHash := hashToken("token-active", state.pepper)
	state.feedByToken[tokenHash] = state.feeds["feed-1"]
	state.feedEpisodes[tokenHash] = []publication.FeedEpisode{{EpisodeID: "episode-1", ProgramID: "program-1", ProgramTitle: "Program Title", ProgramAuthor: "Author Name", Language: "zh-CN", Title: "Hello & Goodbye", Description: "Fish & Chips <Test>", PublishedAt: time.Now().Add(-time.Hour), ContentType: "audio/mpeg", SizeBytes: int64(len("FAKEMEDIA-CONTENT")), SHA256: "sha256-media-1", UpdatedAt: time.Now().Add(-time.Minute)}}
	service := publication.NewService(state, state, mediaStore, state.pepper)
	server := NewServer(config.Config{AppEnv: "development", FrontendOrigin: "http://127.0.0.1:5173", SessionCookieName: "podcast_hub_session", CSRFHeaderName: "X-CSRF-Token"}, nil, nil, HealthDependencies{}, nil, nil, nil, nil, nil, service)
	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		if state.sessionUser.ID == "" {
			return auth.Session{}, auth.User{}, auth.ErrNotAuthenticated
		}
		return auth.Session{ID: "session-1", UserID: state.sessionUser.ID}, state.sessionUser, nil
	}
	return server, state
}

func (s *publicationTestState) FindUserByEmail(_ context.Context, email string) (auth.User, bool, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, true, nil
		}
	}
	return auth.User{}, false, nil
}

func (s *publicationTestState) GrantProgramAccess(_ context.Context, grant publication.ProgramAccessGrant, _ string) (publication.ProgramAccessGrant, error) {
	for _, item := range s.grants {
		if item.UserID == grant.UserID && item.ProgramID == grant.ProgramID && item.Status == publication.ProgramAccessActive {
			return item, nil
		}
	}
	s.grants[grant.ID] = grant
	return grant, nil
}

func (s *publicationTestState) ListProgramAccessGrants(_ context.Context, programID string) ([]publication.ProgramAccessGrant, error) {
	items := []publication.ProgramAccessGrant{}
	for _, item := range s.grants {
		if item.ProgramID == programID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *publicationTestState) RevokeProgramAccess(_ context.Context, grantID string, _ string, reason string, revokedAt time.Time) (publication.ProgramAccessGrant, error) {
	item, ok := s.grants[grantID]
	if !ok || item.Status != publication.ProgramAccessActive {
		return publication.ProgramAccessGrant{}, publication.ErrProgramAccessNotFound
	}
	item.Status = publication.ProgramAccessRevoked
	item.Reason = reason
	item.RevokedAt = &revokedAt
	item.UpdatedAt = revokedAt
	s.grants[grantID] = item
	return item, nil
}

func (s *publicationTestState) GetRSSFeed(_ context.Context, feedID string) (publication.RSSFeed, bool, error) {
	item, ok := s.feeds[feedID]
	return item, ok, nil
}

func (s *publicationTestState) GetRSSFeedForUser(_ context.Context, feedID string, userID string) (publication.RSSFeed, bool, error) {
	item, ok := s.feeds[feedID]
	if !ok || item.UserID != userID {
		return publication.RSSFeed{}, false, nil
	}
	return item, true, nil
}

func (s *publicationTestState) ListRSSFeedsByUser(_ context.Context, userID string) ([]publication.RSSFeed, error) {
	items := []publication.RSSFeed{}
	for _, item := range s.feeds {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *publicationTestState) ListAdminRSSFeeds(_ context.Context) ([]publication.AdminRSSFeed, error) {
	items := []publication.AdminRSSFeed{}
	for _, item := range s.feeds {
		user := s.users[item.UserID]
		items = append(items, publication.AdminRSSFeed{RSSFeed: item, UserEmailHint: user.Email[:1] + "***"})
	}
	return items, nil
}

func (s *publicationTestState) CreateRSSFeed(_ context.Context, feed publication.RSSFeed, tokenHash string) (publication.RSSFeed, error) {
	s.feeds[feed.ID] = feed
	s.feedByToken[tokenHash] = feed
	return feed, nil
}

func (s *publicationTestState) RotateRSSFeed(_ context.Context, feedID string, userID string, tokenHash string, tokenPrefix string, rotatedAt time.Time, expiresAt *time.Time) (publication.RSSFeed, error) {
	item, ok := s.feeds[feedID]
	if !ok || item.UserID != userID {
		return publication.RSSFeed{}, publication.ErrFeedForbidden
	}
	for key, feed := range s.feedByToken {
		if feed.ID == feedID {
			delete(s.feedByToken, key)
			delete(s.feedEpisodes, key)
		}
	}
	item.TokenPrefix = tokenPrefix
	item.RotatedAt = &rotatedAt
	item.ExpiresAt = expiresAt
	item.Status = publication.FeedStatusActive
	s.feeds[feedID] = item
	s.feedByToken[tokenHash] = item
	s.feedEpisodes[tokenHash] = []publication.FeedEpisode{{EpisodeID: "episode-1", ProgramID: "program-1", ProgramTitle: "Program Title", ProgramAuthor: "Author Name", Language: "zh-CN", Title: "Hello & Goodbye", Description: "Fish & Chips <Test>", PublishedAt: time.Now().Add(-time.Hour), ContentType: "audio/mpeg", SizeBytes: int64(len("FAKEMEDIA-CONTENT")), SHA256: "sha256-media-1", UpdatedAt: time.Now()}}
	return item, nil
}

func (s *publicationTestState) RevokeRSSFeed(_ context.Context, feedID string, revokedAt time.Time) (publication.RSSFeed, error) {
	item, ok := s.feeds[feedID]
	if !ok {
		return publication.RSSFeed{}, publication.ErrFeedNotFound
	}
	item.Status = publication.FeedStatusRevoked
	item.RevokedAt = &revokedAt
	s.feeds[feedID] = item
	for key, feed := range s.feedByToken {
		if feed.ID == feedID {
			s.feedByToken[key] = item
		}
	}
	return item, nil
}

func (s *publicationTestState) DeleteRSSFeed(_ context.Context, feedID string, userID string) error {
	item, ok := s.feeds[feedID]
	if !ok || item.UserID != userID {
		return publication.ErrFeedForbidden
	}
	delete(s.feeds, feedID)
	for key, feed := range s.feedByToken {
		if feed.ID == feedID {
			delete(s.feedByToken, key)
			delete(s.feedEpisodes, key)
		}
	}
	return nil
}

func (s *publicationTestState) GetRSSFeedByTokenHash(_ context.Context, tokenHash string, _ time.Time) (publication.RSSFeed, auth.User, bool, error) {
	feed, ok := s.feedByToken[tokenHash]
	if !ok {
		return publication.RSSFeed{}, auth.User{}, false, nil
	}
	return feed, s.users[feed.UserID], true, nil
}

func (s *publicationTestState) TouchRSSFeed(_ context.Context, feedID string, usedAt time.Time) error {
	item := s.feeds[feedID]
	item.LastUsedAt = &usedAt
	s.feeds[feedID] = item
	for key, feed := range s.feedByToken {
		if feed.ID == feedID {
			feed.LastUsedAt = &usedAt
			s.feedByToken[key] = feed
		}
	}
	return nil
}

func (s *publicationTestState) GetAuthorizedMediaForUser(_ context.Context, userID string, episodeID string) (publication.AuthorizedMedia, bool, error) {
	asset, ok := s.mediaByEpisode[episodeID]
	if !ok || !s.hasActiveGrant(userID, asset.ProgramID) {
		return publication.AuthorizedMedia{}, false, nil
	}
	return asset, true, nil
}

func (s *publicationTestState) GetAuthorizedMediaForFeed(_ context.Context, tokenHash string, episodeID string, _ time.Time) (publication.AuthorizedMedia, bool, error) {
	feed, ok := s.feedByToken[tokenHash]
	if !ok || feed.Status != publication.FeedStatusActive {
		return publication.AuthorizedMedia{}, false, nil
	}
	asset, ok := s.mediaByEpisode[episodeID]
	if !ok || !s.hasActiveGrant(feed.UserID, asset.ProgramID) {
		return publication.AuthorizedMedia{}, false, nil
	}
	asset.FeedID = &feed.ID
	asset.FeedOwnerUserID = &feed.UserID
	return asset, true, nil
}

func (s *publicationTestState) ListAuthorizedFeedEpisodes(_ context.Context, tokenHash string, _ time.Time) (publication.RSSFeed, auth.User, []publication.FeedEpisode, error) {
	feed, ok := s.feedByToken[tokenHash]
	if !ok {
		return publication.RSSFeed{}, auth.User{}, nil, publication.ErrFeedTokenInvalid
	}
	items := []publication.FeedEpisode{}
	for _, item := range s.feedEpisodes[tokenHash] {
		if s.hasActiveGrant(feed.UserID, item.ProgramID) {
			items = append(items, item)
		}
	}
	return feed, s.users[feed.UserID], items, nil
}

func (s *publicationTestState) InsertAuditLog(_ context.Context, event auth.AuditEvent) error {
	s.auditEvents = append(s.auditEvents, event)
	return nil
}

func (s *publicationTestState) hasActiveGrant(userID string, programID string) bool {
	for _, item := range s.grants {
		if item.UserID == userID && item.ProgramID == programID && item.Status == publication.ProgramAccessActive {
			return true
		}
	}
	return false
}

func (s *publicationTestState) revokeGrant(grantID string) {
	item := s.grants[grantID]
	now := time.Now()
	item.Status = publication.ProgramAccessRevoked
	item.RevokedAt = &now
	item.UpdatedAt = now
	s.grants[grantID] = item
}

func hashToken(token string, pepper string) string {
	return security.HashWithPepper(token, pepper)
}

func assertNoSensitiveLeak(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"storage_key", "published_storage_key", ".local", "/Users/", "token-active", "secret"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
}
