package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/content"
)

func (s *Server) handleAdminReviews(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Review 服务尚未启用。")
		return
	}
	items, err := s.content.ListReviews(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "无法读取审核队列。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"reviews": items})
}

func (s *Server) handleAdminReview(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	item, ok := s.readReview(w, r)
	if !ok {
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"review": item})
}

func (s *Server) handleAdminReviewApprove(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	item, err := s.content.ApproveReview(r.Context(), content.ReviewDecisionInput{ReviewID: chi.URLParam(r, "reviewId"), ActorID: user.ID})
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"review": item})
}

func (s *Server) handleAdminReviewReject(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	item, err := s.content.RejectReview(r.Context(), content.ReviewDecisionInput{ReviewID: chi.URLParam(r, "reviewId"), ActorID: user.ID, Reason: input.Reason})
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"review": item})
}

func (s *Server) handleAdminPrograms(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Program 服务尚未启用。")
		return
	}
	items, err := s.content.ListAdminPrograms(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "无法读取 Program。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"programs": items})
}

func (s *Server) handleAdminProgram(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Program 服务尚未启用。")
		return
	}
	program, err := s.content.GetAdminProgram(r.Context(), chi.URLParam(r, "programId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusNotFound, "program_not_found", "Program 不存在。")
		return
	}
	episodes, err := s.content.ListProgramEpisodes(r.Context(), program.ID)
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "无法读取 Episode。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"program": program, "episodes": episodes})
}

func (s *Server) handleAdminEpisode(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Episode 服务尚未启用。")
		return
	}
	episode, err := s.content.GetAdminEpisode(r.Context(), chi.URLParam(r, "episodeId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusNotFound, "episode_not_found", "Episode 不存在。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episode": episode})
}

func (s *Server) handleAdminProgramPatch(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	var input struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Author      *string `json:"author"`
		Language    *string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_json", "请求 JSON 无效。")
		return
	}
	program, err := s.content.UpdateProgram(r.Context(), chi.URLParam(r, "programId"), content.UpdateProgramInput{Title: input.Title, Description: input.Description, Author: input.Author, Language: input.Language, ActorID: user.ID})
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"program": program})
}

func (s *Server) handleAdminEpisodePatch(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	var input struct {
		Title           *string `json:"title"`
		Description     *string `json:"description"`
		DurationSeconds *int    `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_json", "请求 JSON 无效。")
		return
	}
	episode, err := s.content.UpdateEpisode(r.Context(), chi.URLParam(r, "episodeId"), content.UpdateEpisodeInput{Title: input.Title, Description: input.Description, DurationSeconds: input.DurationSeconds, ActorID: user.ID})
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episode": episode})
}

func (s *Server) handleAdminProgramSubmitReview(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if _, ok := requireAdminAction(w, r, s); !ok {
		return
	}
	item, err := s.content.SubmitProgramReview(r.Context(), chi.URLParam(r, "programId"))
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"review": item})
}

func (s *Server) handleAdminEpisodeSubmitReview(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if _, ok := requireAdminAction(w, r, s); !ok {
		return
	}
	item, err := s.content.SubmitEpisodeReview(r.Context(), chi.URLParam(r, "episodeId"))
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"review": item})
}

func (s *Server) handleAdminProgramPublish(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	program, err := s.content.PublishProgram(r.Context(), chi.URLParam(r, "programId"), user.ID)
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"program": program})
}

func (s *Server) handleAdminEpisodePublish(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	episode, err := s.content.PublishEpisode(r.Context(), chi.URLParam(r, "episodeId"), user.ID)
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episode": episode})
}

func (s *Server) handleAdminProgramArchive(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	program, err := s.content.ArchiveProgram(r.Context(), chi.URLParam(r, "programId"), user.ID)
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"program": program})
}

func (s *Server) handleAdminEpisodeArchive(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := requireAdminAction(w, r, s)
	if !ok {
		return
	}
	episode, err := s.content.ArchiveEpisode(r.Context(), chi.URLParam(r, "episodeId"), user.ID)
	if err != nil {
		s.writeContentError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episode": episode})
}

func (s *Server) readReview(w stdhttp.ResponseWriter, r *stdhttp.Request) (content.ReviewItem, bool) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Review 服务尚未启用。")
		return content.ReviewItem{}, false
	}
	item, err := s.content.GetReview(r.Context(), chi.URLParam(r, "reviewId"))
	if err != nil {
		writeError(w, r, stdhttp.StatusNotFound, "review_not_found", "ReviewItem 不存在。")
		return content.ReviewItem{}, false
	}
	return item, true
}

func requireAdminAction(w stdhttp.ResponseWriter, r *stdhttp.Request, s *Server) (auth.User, bool) {
	if s.content == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Content 服务尚未启用。")
		return auth.User{}, false
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return auth.User{}, false
	}
	user, ok := authUserFromContext(r.Context())
	if !ok {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return auth.User{}, false
	}
	return user, true
}

func (s *Server) writeContentError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	switch {
	case errors.Is(err, content.ErrReviewReasonRequired):
		writeError(w, r, stdhttp.StatusBadRequest, "review_reason_required", "拒绝审核必须提供原因。")
	case errors.Is(err, content.ErrPublishPrecondition):
		writeError(w, r, stdhttp.StatusConflict, "publish_precondition_failed", "发布前置条件未满足。")
	case errors.Is(err, content.ErrInvalidTransition):
		writeError(w, r, stdhttp.StatusConflict, "invalid_transition", "当前状态不允许该操作。")
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "请求暂时无法完成。")
	}
}
