package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/intake"
)

func (s *Server) handleAdminImportJobIntake(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.intake == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Intake 服务尚未启用。")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	result, err := s.intake.Run(r.Context(), chi.URLParam(r, "jobId"))
	if err != nil {
		s.writeIntakeError(w, r, err, result.Issues)
		return
	}
	writeJSON(w, stdhttp.StatusOK, result)
}

func (s *Server) handleAdminImportJobIntakeStatus(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.intake == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Intake 服务尚未启用。")
		return
	}
	run, found, err := s.intake.Status(r.Context(), chi.URLParam(r, "jobId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Intake 状态暂时无法读取。")
		return
	}
	if !found {
		writeJSON(w, stdhttp.StatusOK, map[string]any{"intake_run": nil})
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"intake_run": run})
}

func (s *Server) handleAdminStagingPrograms(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Staging 服务尚未启用。")
		return
	}
	items, err := s.content.ListStagingPrograms(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "无法读取 staging Program。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"programs": items})
}

func (s *Server) handleAdminStagingProgram(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Staging 服务尚未启用。")
		return
	}
	item, err := s.content.GetProgram(r.Context(), chi.URLParam(r, "programId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusNotFound, "program_not_found", "Program 不存在。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"program": item})
}

func (s *Server) handleAdminStagingEpisodes(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Staging 服务尚未启用。")
		return
	}
	items, err := s.content.ListStagingEpisodes(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "无法读取 staging Episode。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episodes": items})
}

func (s *Server) handleAdminStagingEpisode(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Staging 服务尚未启用。")
		return
	}
	item, err := s.content.GetEpisode(r.Context(), chi.URLParam(r, "episodeId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusNotFound, "episode_not_found", "Episode 不存在。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episode": item})
}

func (s *Server) writeIntakeError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error, issues []string) {
	switch {
	case errors.Is(err, intake.ErrJobNotCompleted):
		writeError(w, r, stdhttp.StatusConflict, "job_not_completed", "只有 completed Import Job 可以导入待审核区。")
	case errors.Is(err, intake.ErrBundleMissing):
		writeJSON(w, stdhttp.StatusUnprocessableEntity, map[string]any{"error": map[string]any{"code": "metadata_bundle_missing", "message": "Import Job 缺少 metadata bundle。", "validation_issues": issues}})
	case errors.Is(err, intake.ErrBundleInvalid):
		writeJSON(w, stdhttp.StatusUnprocessableEntity, map[string]any{"error": map[string]any{"code": "metadata_bundle_invalid", "message": "metadata bundle 校验失败。", "validation_issues": issues}})
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Intake 请求暂时无法完成。")
	}
}
