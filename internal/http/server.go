package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/connectors"
	"github.com/mdxz2048/podcast-hub/internal/content"
	"github.com/mdxz2048/podcast-hub/internal/intake"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/publication"
	"github.com/mdxz2048/podcast-hub/internal/security"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

type Server struct {
	cfg              config.Config
	auth             *auth.Service
	turnstile        security.TurnstileVerifier
	health           HealthDependencies
	connectors       *connectors.Service
	sources          *sources.Service
	jobs             *jobs.Service
	content          *content.Service
	intake           *intake.Service
	publication      *publication.Service
	resolveSessionFn func(ctx context.Context, token string) (auth.Session, auth.User, error)
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

type RedisPinger interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

type HealthDependencies struct {
	DB       DBPinger
	Redis    RedisPinger
	SMTPHost string
	SMTPPort int
}

func NewServer(cfg config.Config, authService *auth.Service, turnstile security.TurnstileVerifier, health HealthDependencies, connectorService *connectors.Service, sourceService *sources.Service, jobService *jobs.Service, extras ...any) *Server {
	server := &Server{
		cfg:        cfg,
		auth:       authService,
		turnstile:  turnstile,
		health:     health,
		connectors: connectorService,
		sources:    sourceService,
		jobs:       jobService,
	}
	for _, extra := range extras {
		switch value := extra.(type) {
		case *content.Service:
			server.content = value
		case *intake.Service:
			server.intake = value
		case *publication.Service:
			server.publication = value
		}
	}
	return server
}

func (s *Server) Router() stdhttp.Handler {
	router := chi.NewRouter()
	router.Use(s.secureRequestIDMiddleware)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(s.corsMiddleware)
	router.Use(s.securityHeadersMiddleware)
	router.Use(s.csrfOriginMiddleware)

	router.Options("/*", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})
	router.Get("/healthz", s.handleHealthz)
	router.Get("/readyz", s.handleReadyz)
	router.Get("/rss/private/{opaqueToken}.xml", s.handlePrivateRSS)
	router.Get("/rss/private/{opaqueToken}/episodes/{episodeId}/media", s.handlePrivateRSSMedia)
	router.Head("/rss/private/{opaqueToken}/episodes/{episodeId}/media", s.handlePrivateRSSMedia)
	router.Group(func(r chi.Router) {
		r.Use(s.RequireAuth)
		r.Get("/programs", s.handleUserPrograms)
		r.Get("/programs/{programId}", s.handleUserProgram)
		r.Get("/programs/{programId}/episodes", s.handleUserProgramEpisodes)
		r.Get("/episodes/{episodeId}", s.handleUserEpisode)
		r.Get("/media/episodes/{episodeId}", s.handleUserMedia)
		r.Head("/media/episodes/{episodeId}", s.handleUserMedia)
		r.Get("/me/collections", s.handleUserCollections)
		r.Post("/me/collections", s.handleUserCollectionCreate)
		r.Patch("/me/collections/{collectionId}", s.handleUserCollectionPatch)
		r.Delete("/me/collections/{collectionId}", s.handleUserCollectionDelete)
		r.Post("/me/collections/{collectionId}/programs", s.handleUserCollectionAddProgram)
		r.Delete("/me/collections/{collectionId}/programs/{programId}", s.handleUserCollectionRemoveProgram)
		r.Get("/me/rss-feeds", s.handleUserRSSFeeds)
		r.Post("/me/rss-feeds", s.handleUserRSSFeedCreate)
		r.Post("/me/rss-feeds/{feedId}/rotate", s.handleUserRSSFeedRotate)
		r.Post("/me/rss-feeds/{feedId}/revoke", s.handleUserRSSFeedRevoke)
		r.Delete("/me/rss-feeds/{feedId}", s.handleUserRSSFeedDelete)
	})

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register/request-code", s.handleRegisterRequestCode)
		r.Post("/register/verify-code", s.handleRegisterVerifyCode)
		r.Post("/login", s.handleLogin)
		r.Post("/logout", s.handleLogout)
		r.Post("/password-reset/request", s.handlePasswordResetRequest)
		r.Post("/password-reset/verify", s.handlePasswordResetVerify)
		r.Get("/me", s.handleMe)
	})
	router.Route("/admin", func(r chi.Router) {
		r.Use(s.RequireAuth)
		r.Use(s.RequireAdmin)
		r.Get("/me", s.handleAdminMe)
		r.Get("/system/status", s.handleAdminSystemStatus)
		r.Get("/connectors", s.handleAdminConnectors)
		r.Post("/connectors/upload", s.handleAdminConnectorUpload)
		r.Get("/connectors/{connectorId}", s.handleAdminConnector)
		r.Get("/connectors/{connectorId}/versions", s.handleAdminConnectorVersions)
		r.Post("/connectors/{connectorId}/disable", s.handleAdminConnectorDisable)
		r.Post("/connectors/{connectorId}/enable", s.handleAdminConnectorEnable)
		r.Get("/connector-versions/{versionId}", s.handleAdminConnectorVersion)
		r.Post("/connector-versions/{versionId}/approve", s.handleAdminConnectorVersionApprove)
		r.Post("/connector-versions/{versionId}/reject", s.handleAdminConnectorVersionReject)
		r.Post("/connector-versions/{versionId}/disable", s.handleAdminConnectorVersionDisable)
		r.Get("/sources", s.handleAdminSources)
		r.Post("/sources", s.handleAdminSourceCreate)
		r.Get("/sources/{sourceId}", s.handleAdminSource)
		r.Patch("/sources/{sourceId}", s.handleAdminSourceUpdate)
		r.Post("/sources/{sourceId}/enable", s.handleAdminSourceEnable)
		r.Post("/sources/{sourceId}/disable", s.handleAdminSourceDisable)
		r.Get("/secrets", s.handleAdminSecrets)
		r.Post("/secrets/text", s.handleAdminSecretTextCreate)
		r.Post("/secrets/file", s.handleAdminSecretFileCreate)
		r.Post("/secrets/{secretId}/revoke", s.handleAdminSecretRevoke)
		r.Post("/sources/{sourceId}/secret-bindings", s.handleAdminSourceSecretBind)
		r.Delete("/sources/{sourceId}/secret-bindings/{bindingId}", s.handleAdminSourceSecretUnbind)
		r.Get("/import-jobs", s.handleAdminImportJobs)
		r.Post("/sources/{sourceId}/import-jobs", s.handleAdminImportJobCreate)
		r.Get("/import-jobs/{jobId}", s.handleAdminImportJob)
		r.Get("/import-jobs/{jobId}/events", s.handleAdminImportJobEvents)
		r.Get("/import-jobs/{jobId}/artifacts", s.handleAdminImportJobArtifacts)
		r.Post("/import-jobs/{jobId}/cancel", s.handleAdminImportJobCancel)
		r.Post("/import-jobs/{jobId}/intake", s.handleAdminImportJobIntake)
		r.Get("/import-jobs/{jobId}/intake-status", s.handleAdminImportJobIntakeStatus)
		r.Get("/review", s.handleAdminReviews)
		r.Get("/review/{reviewId}", s.handleAdminReview)
		r.Post("/review/{reviewId}/approve", s.handleAdminReviewApprove)
		r.Post("/review/{reviewId}/reject", s.handleAdminReviewReject)
		r.Get("/programs", s.handleAdminPrograms)
		r.Get("/rss-feeds", s.handleAdminRSSFeeds)
		r.Post("/rss-feeds/{feedId}/revoke", s.handleAdminRSSFeedRevoke)
		r.Get("/programs/{programId}/access-grants", s.handleAdminProgramAccessGrants)
		r.Post("/programs/{programId}/access-grants", s.handleAdminProgramAccessGrant)
		r.Post("/program-access/{grantId}/revoke", s.handleAdminProgramAccessRevoke)
		r.Get("/programs/{programId}", s.handleAdminProgram)
		r.Patch("/programs/{programId}", s.handleAdminProgramPatch)
		r.Post("/programs/{programId}/submit-review", s.handleAdminProgramSubmitReview)
		r.Post("/programs/{programId}/publish", s.handleAdminProgramPublish)
		r.Post("/programs/{programId}/archive", s.handleAdminProgramArchive)
		r.Get("/episodes/{episodeId}", s.handleAdminEpisode)
		r.Patch("/episodes/{episodeId}", s.handleAdminEpisodePatch)
		r.Post("/episodes/{episodeId}/submit-review", s.handleAdminEpisodeSubmitReview)
		r.Post("/episodes/{episodeId}/publish", s.handleAdminEpisodePublish)
		r.Post("/episodes/{episodeId}/archive", s.handleAdminEpisodeArchive)
		r.Get("/staging/programs", s.handleAdminStagingPrograms)
		r.Get("/staging/programs/{programId}", s.handleAdminStagingProgram)
		r.Get("/staging/episodes", s.handleAdminStagingEpisodes)
		r.Get("/staging/episodes/{episodeId}", s.handleAdminStagingEpisode)
	})
	return router
}

type errorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

func writeJSON(w stdhttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w stdhttp.ResponseWriter, r *stdhttp.Request, status int, code string, message string) {
	response := errorResponse{}
	response.Error.Code = code
	response.Error.Message = message
	response.Error.RequestID = middleware.GetReqID(r.Context())
	writeJSON(w, status, response)
}

func mapAuthError(err error) (int, string, string) {
	switch {
	case errors.Is(err, auth.ErrInvalidEmail):
		return stdhttp.StatusBadRequest, "invalid_email", "邮箱格式无效。"
	case errors.Is(err, auth.ErrWeakPassword):
		return stdhttp.StatusBadRequest, "weak_password", "密码强度不足。"
	case errors.Is(err, auth.ErrPasswordMismatch):
		return stdhttp.StatusBadRequest, "password_mismatch", "两次输入的密码不一致。"
	case errors.Is(err, auth.ErrRateLimited):
		return stdhttp.StatusTooManyRequests, "rate_limited", "请求过于频繁，请稍后再试。"
	case errors.Is(err, auth.ErrTurnstileRequired):
		return stdhttp.StatusBadRequest, "turnstile_required", "请先完成人机验证。"
	case errors.Is(err, auth.ErrTurnstileFailed):
		return stdhttp.StatusBadRequest, "turnstile_failed", "人机验证失败，请重试。"
	case errors.Is(err, auth.ErrInvalidOrExpiredCode):
		return stdhttp.StatusBadRequest, "invalid_or_expired_code", "验证码错误或已过期。"
	case errors.Is(err, auth.ErrVerificationCodeExpired):
		return stdhttp.StatusBadRequest, "verification_code_expired", "验证码已过期，请重新获取。"
	case errors.Is(err, auth.ErrTooManyAttempts):
		return stdhttp.StatusTooManyRequests, "too_many_attempts", "尝试次数过多，请稍后再试。"
	case errors.Is(err, auth.ErrInvalidCredentials):
		return stdhttp.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误。"
	case errors.Is(err, auth.ErrAccountUnavailable):
		return stdhttp.StatusForbidden, "account_unavailable", "账号当前不可用。"
	case errors.Is(err, auth.ErrNotAuthenticated):
		return stdhttp.StatusUnauthorized, "not_authenticated", "当前未登录。"
	case errors.Is(err, auth.ErrInvalidOrExpiredProof):
		return stdhttp.StatusBadRequest, "invalid_or_expired_proof", "重置凭证错误或已过期。"
	case errors.Is(err, auth.ErrInvalidResetProof):
		return stdhttp.StatusBadRequest, "invalid_reset_proof", "重置验证码错误。"
	case errors.Is(err, auth.ErrResetProofExpired):
		return stdhttp.StatusBadRequest, "reset_proof_expired", "重置验证码已过期。"
	case errors.Is(err, auth.ErrTemporarilyUnavailable):
		return stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "服务暂时不可用，请稍后重试。"
	case errors.Is(err, auth.ErrForbiddenCSRF):
		return stdhttp.StatusForbidden, "csrf_invalid", "请求未通过安全校验。"
	case errors.Is(err, auth.ErrUnsupportedContentType):
		return stdhttp.StatusUnsupportedMediaType, "unsupported_content_type", "请求内容类型必须为 application/json。"
	case errors.Is(err, auth.ErrInvalidJSONBody):
		return stdhttp.StatusBadRequest, "invalid_json_body", "请求体格式无效。"
	default:
		return stdhttp.StatusInternalServerError, "internal_error", "请求暂时无法完成。"
	}
}

func (s *Server) writeAuthError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	status, code, message := mapAuthError(err)
	writeError(w, r, status, code, message)
}

func (s *Server) resolveSession(ctx context.Context, token string) (auth.Session, auth.User, error) {
	if s.resolveSessionFn != nil {
		return s.resolveSessionFn(ctx, token)
	}
	return s.auth.ResolveSession(ctx, token)
}

func (s *Server) parseJSONBody(r *stdhttp.Request, target any) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return auth.ErrUnsupportedContentType
	}
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: %s", auth.ErrInvalidJSONBody, err.Error())
	}
	return nil
}

func (s *Server) setSessionCookie(w stdhttp.ResponseWriter, token string) {
	cookie := &stdhttp.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		Expires:  time.Now().Add(s.cfg.SessionTTL),
	}
	if strings.TrimSpace(s.cfg.SessionCookieDomain) != "" {
		cookie.Domain = s.cfg.SessionCookieDomain
	}
	stdhttp.SetCookie(w, cookie)
}

func (s *Server) clearSessionCookie(w stdhttp.ResponseWriter) {
	stdhttp.SetCookie(w, &stdhttp.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (s *Server) setCSRFCookie(w stdhttp.ResponseWriter, token string) {
	stdhttp.SetCookie(w, &stdhttp.Cookie{
		Name:     "podcast_hub_csrf",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		Expires:  time.Now().Add(s.cfg.SessionTTL),
	})
}
