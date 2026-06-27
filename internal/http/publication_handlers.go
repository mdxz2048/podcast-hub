package http

import (
	"errors"
	"fmt"
	"io"
	"net"
	stdhttp "net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/publication"
)

func (s *Server) handleUserMedia(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Media service is unavailable.")
		return
	}
	user, ok := authUserFromContext(r.Context())
	if !ok {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return
	}
	asset, err := s.publication.ResolveUserMedia(r.Context(), user.ID, chi.URLParam(r, "episodeId"))
	if err != nil {
		s.writePublicationAccessError(w, r, err)
		return
	}
	s.serveAuthorizedMedia(w, r, asset)
}

func (s *Server) handlePrivateRSSMedia(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	token := strings.TrimSpace(chi.URLParam(r, "opaqueToken"))
	asset, err := s.publication.ResolveFeedMedia(r.Context(), token, chi.URLParam(r, "episodeId"))
	if err != nil {
		s.writePublicationAccessError(w, r, err)
		return
	}
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	s.serveAuthorizedMedia(w, r, asset)
}

func (s *Server) handlePrivateRSS(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	token := strings.TrimSpace(chi.URLParam(r, "opaqueToken"))
	baseURL := requestBaseURL(r)
	doc, body, err := s.publication.BuildPrivateFeed(r.Context(), token, baseURL)
	if err != nil {
		s.writePublicationAccessError(w, r, err)
		return
	}
	setPrivateCacheHeaders(w)
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("ETag", doc.ETag)
	w.Header().Set("Last-Modified", doc.UpdatedAt.UTC().Format(stdhttp.TimeFormat))
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	if ifNoneMatch := strings.TrimSpace(r.Header.Get("If-None-Match")); ifNoneMatch != "" && ifNoneMatch == doc.ETag {
		w.WriteHeader(stdhttp.StatusNotModified)
		return
	}
	if isModifiedSinceFresh(r, doc.UpdatedAt) {
		w.WriteHeader(stdhttp.StatusNotModified)
		return
	}
	w.WriteHeader(stdhttp.StatusOK)
	_, _ = w.Write(body)
}

func (s *Server) handleUserRSSFeeds(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	user, _ := authUserFromContext(r.Context())
	feeds, err := s.publication.ListUserFeeds(r.Context(), user.ID)
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Unable to load RSS feeds.")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"feeds": feeds})
}

func (s *Server) handleUserRSSFeedCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	var req struct {
		Name             string `json:"name"`
		ExpiresAtRFC3339 string `json:"expires_at"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	result, err := s.publication.CreateFeed(r.Context(), user.ID, req.Name, requestBaseURL(r), parseOptionalTime(req.ExpiresAtRFC3339))
	if err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"feed": result.Feed, "token": result.Token, "feed_url": result.FeedURL})
}

func (s *Server) handleUserRSSFeedRotate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	var req struct {
		ExpiresAtRFC3339 string `json:"expires_at"`
	}
	if err := s.parseJSONBody(r, &req); err != nil && !errors.Is(err, stdhttp.ErrBodyNotAllowed) {
		s.writeAuthError(w, r, err)
		return
	}
	result, err := s.publication.RotateFeed(r.Context(), user.ID, chi.URLParam(r, "feedId"), requestBaseURL(r), parseOptionalTime(req.ExpiresAtRFC3339))
	if err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"feed": result.Feed, "token": result.Token, "feed_url": result.FeedURL})
}

func (s *Server) handleUserRSSFeedRevoke(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	feed, err := s.publication.RevokeFeed(r.Context(), user.ID, chi.URLParam(r, "feedId"))
	if err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"feed": feed})
}

func (s *Server) handleUserRSSFeedDelete(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	if err := s.publication.DeleteFeed(r.Context(), user.ID, chi.URLParam(r, "feedId")); err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	w.WriteHeader(stdhttp.StatusNoContent)
}

func (s *Server) handleAdminRSSFeeds(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	feeds, err := s.publication.ListAdminFeeds(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Unable to load RSS feeds.")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"feeds": feeds})
}

func (s *Server) handleAdminRSSFeedRevoke(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "RSS service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	feed, err := s.publication.AdminRevokeFeed(r.Context(), chi.URLParam(r, "feedId"), user.ID)
	if err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"feed": feed})
}

func (s *Server) handleAdminProgramAccessGrants(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Publication service is unavailable.")
		return
	}
	items, err := s.publication.ListProgramAccessGrants(r.Context(), chi.URLParam(r, "programId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Unable to load access grants.")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"grants": items})
}

func (s *Server) handleAdminProgramAccessGrant(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Publication service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	var req struct {
		Email  string `json:"email"`
		Reason string `json:"reason"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	grant, err := s.publication.GrantProgramAccess(r.Context(), chi.URLParam(r, "programId"), req.Email, user.ID, req.Reason)
	if err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"grant": grant})
}

func (s *Server) handleAdminProgramAccessRevoke(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Publication service is unavailable.")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, _ := authUserFromContext(r.Context())
	var req struct {
		Reason string `json:"reason"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	grant, err := s.publication.RevokeProgramAccess(r.Context(), chi.URLParam(r, "grantId"), user.ID, req.Reason)
	if err != nil {
		s.writePublicationMutationError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"grant": grant})
}

func (s *Server) serveAuthorizedMedia(w stdhttp.ResponseWriter, r *stdhttp.Request, asset publication.AuthorizedMedia) {
	file, err := s.publication.OpenPublishedMedia(r.Context(), asset.PublishedKey)
	if err != nil {
		s.writePublicationAccessError(w, r, publication.ErrMediaNotAvailable)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Unable to read media.")
		return
	}
	etag := buildMediaETag(asset)
	setPrivateCacheHeaders(w)
	w.Header().Set("Content-Type", asset.ContentType)
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", asset.PublishedAt.UTC().Format(stdhttp.TimeFormat))
	if ifNoneMatch := strings.TrimSpace(r.Header.Get("If-None-Match")); ifNoneMatch != "" && ifNoneMatch == etag {
		w.WriteHeader(stdhttp.StatusNotModified)
		return
	}
	if isModifiedSinceFresh(r, asset.PublishedAt) {
		w.WriteHeader(stdhttp.StatusNotModified)
		return
	}
	stdhttp.ServeContent(w, r, "media", asset.PublishedAt, fileWithSize{ReadSeeker: file, size: stat.Size()})
}

func (s *Server) writePublicationAccessError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	if errors.Is(err, publication.ErrMediaNotAvailable) || errors.Is(err, publication.ErrFeedTokenInvalid) || errors.Is(err, publication.ErrFeedRevoked) || errors.Is(err, publication.ErrFeedExpired) || errors.Is(err, publication.ErrFeedForbidden) {
		writeError(w, r, stdhttp.StatusNotFound, "resource_not_available", "Resource not available.")
		return
	}
	writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Request could not be completed.")
}

func (s *Server) writePublicationMutationError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	switch {
	case errors.Is(err, publication.ErrInvalidFeedName):
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_feed_name", "Feed name is invalid.")
	case errors.Is(err, publication.ErrFeedForbidden), errors.Is(err, publication.ErrFeedNotFound), errors.Is(err, publication.ErrProgramAccessNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "resource_not_found", "Resource not found.")
	case errors.Is(err, publication.ErrUserNotEligible):
		writeError(w, r, stdhttp.StatusConflict, "user_not_eligible", "User is not eligible for access.")
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Request could not be completed.")
	}
}

func setPrivateCacheHeaders(w stdhttp.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	appendVary(w, "Authorization")
	appendVary(w, "Cookie")
	appendVary(w, "Range")
	appendVary(w, "If-None-Match")
	appendVary(w, "If-Modified-Since")
}

func appendVary(w stdhttp.ResponseWriter, value string) {
	existing := w.Header().Values("Vary")
	for _, item := range existing {
		if item == value {
			return
		}
	}
	w.Header().Add("Vary", value)
}

func requestBaseURL(r *stdhttp.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	}
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	return scheme + "://" + host
}

func parseOptionalTime(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &parsed
}

func isModifiedSinceFresh(r *stdhttp.Request, updatedAt time.Time) bool {
	value := strings.TrimSpace(r.Header.Get("If-Modified-Since"))
	if value == "" {
		return false
	}
	parsed, err := time.Parse(stdhttp.TimeFormat, value)
	if err != nil {
		return false
	}
	return !updatedAt.After(parsed)
}

func buildMediaETag(asset publication.AuthorizedMedia) string {
	return fmt.Sprintf(`"%s-%d-%d"`, asset.SHA256, asset.SizeBytes, asset.PublishedAt.UTC().Unix())
}

type fileWithSize struct {
	io.ReadSeeker
	size int64
}

func (f fileWithSize) Seek(offset int64, whence int) (int64, error) {
	return f.ReadSeeker.Seek(offset, whence)
}

func (f fileWithSize) Stat() (os.FileInfo, error) {
	return fixedFileInfo{size: f.size}, nil
}

type fixedFileInfo struct{ size int64 }

func (f fixedFileInfo) Name() string       { return "media" }
func (f fixedFileInfo) Size() int64        { return f.size }
func (f fixedFileInfo) Mode() os.FileMode  { return 0 }
func (f fixedFileInfo) ModTime() time.Time { return time.Time{} }
func (f fixedFileInfo) IsDir() bool        { return false }
func (f fixedFileInfo) Sys() any           { return nil }

func redactOpaqueToken(token string) string {
	if len(token) <= 8 {
		return "[redacted]"
	}
	return token[:4] + "...[redacted]"
}

func sanitizeRemoteAddr(value string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(value))
	if err == nil {
		return host
	}
	return strings.TrimSpace(value)
}

func parseContentLength(value string) int64 {
	length, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return length
}
