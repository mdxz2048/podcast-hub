package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

func (s *Server) handleAdminImportJobs(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.jobs == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Import Job 服务尚未启用。")
		return
	}
	items, err := s.jobs.ListJobs(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "获取 Import Job 列表失败。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"jobs": items})
}

func (s *Server) handleAdminImportJobCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.jobs == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Import Job 服务尚未启用。")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	user, ok := authUserFromContext(r.Context())
	if !ok {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return
	}
	userID := user.ID
	job, err := s.jobs.CreateManualJob(r.Context(), chi.URLParam(r, "sourceId"), &userID)
	if err != nil {
		s.writeJobError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"job": job})
}

func (s *Server) handleAdminImportJob(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.jobs == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Import Job 服务尚未启用。")
		return
	}
	job, err := s.jobs.GetJob(r.Context(), chi.URLParam(r, "jobId"))
	if err != nil {
		s.writeJobError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"job": job})
}

func (s *Server) handleAdminImportJobEvents(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.jobs == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Import Job 服务尚未启用。")
		return
	}
	events, err := s.jobs.ListEvents(r.Context(), chi.URLParam(r, "jobId"))
	if err != nil {
		s.writeJobError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"events": events})
}

func (s *Server) handleAdminImportJobArtifacts(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.jobs == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Import Job 服务尚未启用。")
		return
	}
	artifacts, err := s.jobs.ListArtifacts(r.Context(), chi.URLParam(r, "jobId"))
	if err != nil {
		s.writeJobError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"artifacts": artifacts})
}

func (s *Server) handleAdminImportJobCancel(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.jobs == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Import Job 服务尚未启用。")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	job, err := s.jobs.CancelJob(r.Context(), chi.URLParam(r, "jobId"))
	if err != nil {
		s.writeJobError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"job": job})
}

func (s *Server) writeJobError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	switch {
	case errors.Is(err, jobs.ErrJobNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "job_not_found", "Import Job 不存在。")
	case errors.Is(err, sources.ErrSourceNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "source_not_found", "Source 不存在。")
	case errors.Is(err, jobs.ErrActiveJobExists):
		writeError(w, r, stdhttp.StatusConflict, "active_job_exists", "该 Source 已有 queued 或 running Import Job。")
	case errors.Is(err, jobs.ErrSourceNotRunnable), errors.Is(err, sources.ErrUnsupportedAlphaMode), errors.Is(err, sources.ErrConnectorUnavailable), errors.Is(err, sources.ErrConnectorVersionInvalid):
		writeError(w, r, stdhttp.StatusConflict, "source_not_runnable", "Source 当前不可运行。")
	case errors.Is(err, sources.ErrMissingRequiredSecrets):
		writeError(w, r, stdhttp.StatusConflict, "missing_required_secrets", "Source 缺少 required Secret 绑定。")
	case errors.Is(err, sources.ErrSecretRevoked):
		writeError(w, r, stdhttp.StatusConflict, "secret_revoked", "Source 绑定的 Secret 已撤销。")
	case errors.Is(err, jobs.ErrInvalidJobState):
		writeError(w, r, stdhttp.StatusConflict, "invalid_job_state", "Import Job 当前状态不允许该操作。")
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Import Job 请求暂时无法完成。")
	}
}
