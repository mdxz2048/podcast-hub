package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/content"
)

func TestReviewAPIRequiresAuthAndAdmin(t *testing.T) {
	server := newReviewAPITestServer()
	server.resolveSessionFn = nil
	req := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "u1", Email: "user@example.invalid", Role: auth.RoleUser, Status: auth.StatusActive}, nil
	}
	req = httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	req.AddCookie(&http.Cookie{Name: "podcast_hub_session", Value: "user-token"})
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestReviewApproveRejectPublishArchiveAPI(t *testing.T) {
	server := newReviewAPITestServer()

	req := adminJobRequest(http.MethodPost, "/admin/review/review-reject/reject")
	req.Body = nopBody(`{"reason":""}`)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected reject without reason 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = adminJobRequest(http.MethodPost, "/admin/programs/program-1/publish")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected unreviewed program publish conflict, got %d", rec.Code)
	}

	req = adminJobRequest(http.MethodPost, "/admin/review/review-program/approve")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"approved"`) {
		t.Fatalf("expected approve 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertContentResponseSafe(t, rec.Body.String())

	req = adminJobRequest(http.MethodPost, "/admin/programs/program-1/publish")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"published"`) {
		t.Fatalf("expected program publish 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = adminJobRequest(http.MethodPost, "/admin/episodes/episode-1/publish")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected unapproved episode publish conflict, got %d", rec.Code)
	}

	req = adminJobRequest(http.MethodPost, "/admin/review/review-episode/approve")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected episode approve 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = adminJobRequest(http.MethodPost, "/admin/episodes/episode-1/publish")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"published"`) {
		t.Fatalf("expected episode publish 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertContentResponseSafe(t, rec.Body.String())

	req = adminJobRequest(http.MethodPost, "/admin/episodes/episode-1/archive")
	rec = httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"archived"`) {
		t.Fatalf("expected archive 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func newReviewAPITestServer() *Server {
	cfg := config.Config{
		AppEnv:            "development",
		FrontendOrigin:    "http://127.0.0.1:5173",
		SessionCookieName: "podcast_hub_session",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	store := newStagingMemoryContentStore()
	store.programs["program-1"] = content.Program{ID: "program-1", CanonicalKey: "source:p1", Title: "Program", Description: "Desc", Author: "Author", Language: "zh-CN", Status: content.ProgramStatusReviewPending, CreatedFromSourceID: "source-1", CreatedFromJobID: "job-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	store.episodes["program-1:ep-1"] = content.Episode{ID: "episode-1", ProgramID: "program-1", ExternalEpisodeID: "ep-1", Title: "Episode", Description: "Desc", PublishedAt: time.Now(), DurationSeconds: 120, Status: content.EpisodeStatusReviewPending, SourceJobID: "job-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	store.reviews = []content.ReviewItem{
		{ID: "review-program", TargetType: "program", TargetID: "program-1", ReviewKind: "metadata", Status: content.ReviewStatusPending, CreatedAt: time.Now()},
		{ID: "review-episode", TargetType: "episode", TargetID: "episode-1", ReviewKind: "metadata", Status: content.ReviewStatusPending, CreatedAt: time.Now()},
		{ID: "review-reject", TargetType: "episode", TargetID: "episode-other", ReviewKind: "rights", Status: content.ReviewStatusPending, CreatedAt: time.Now()},
	}
	store.media = []content.MediaAsset{{ID: "media-1", OwnerType: "episode", OwnerID: "episode-1", MediaKind: "audio", Status: content.MediaStatusStaged, StagedStorageKey: "/Users/hidden/audio.mp3"}}
	server := NewServer(cfg, nil, nil, HealthDependencies{}, nil, nil, nil, content.NewService(store))
	server.resolveSessionFn = func(context.Context, string) (auth.Session, auth.User, error) {
		return auth.Session{}, auth.User{ID: "a1", Email: "admin@example.invalid", Role: auth.RoleAdmin, Status: auth.StatusActive}, nil
	}
	return server
}

func assertContentResponseSafe(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"/Users/", ".local/", "storage_key", "staged_storage_key", "secret", "token", "raw log"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("content response leaked %q: %s", forbidden, body)
		}
	}
}

func nopBody(body string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(body))
}
