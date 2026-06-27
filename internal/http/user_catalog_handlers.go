package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/publication"
)

func (s *Server) handleUserPrograms(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationUser(w, r)
	if !ok {
		return
	}
	items, err := s.publication.ListAuthorizedPrograms(r.Context(), user.ID)
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Unable to load programs.")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"programs": items})
}

func (s *Server) handleUserProgram(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationUser(w, r)
	if !ok {
		return
	}
	program, err := s.publication.GetAuthorizedProgram(r.Context(), user.ID, chi.URLParam(r, "programId"))
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"program": program})
}

func (s *Server) handleUserProgramEpisodes(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationUser(w, r)
	if !ok {
		return
	}
	episodes, err := s.publication.ListAuthorizedEpisodes(r.Context(), user.ID, chi.URLParam(r, "programId"))
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episodes": episodes})
}

func (s *Server) handleUserEpisode(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationUser(w, r)
	if !ok {
		return
	}
	episode, err := s.publication.GetAuthorizedEpisode(r.Context(), user.ID, chi.URLParam(r, "episodeId"))
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"episode": episode})
}

func (s *Server) handleUserCollections(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationUser(w, r)
	if !ok {
		return
	}
	collections, err := s.publication.ListUserCollections(r.Context(), user.ID)
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Unable to load collections.")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"collections": collections})
}

func (s *Server) handleUserCollectionCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationMutationUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	collection, err := s.publication.CreateUserCollection(r.Context(), user.ID, req.Title, req.Description)
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"collection": collection})
}

func (s *Server) handleUserCollectionPatch(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationMutationUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	collection, err := s.publication.UpdateUserCollection(r.Context(), user.ID, chi.URLParam(r, "collectionId"), req.Title, req.Description)
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"collection": collection})
}

func (s *Server) handleUserCollectionDelete(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationMutationUser(w, r)
	if !ok {
		return
	}
	if err := s.publication.DeleteUserCollection(r.Context(), user.ID, chi.URLParam(r, "collectionId")); err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	w.WriteHeader(stdhttp.StatusNoContent)
}

func (s *Server) handleUserCollectionAddProgram(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationMutationUser(w, r)
	if !ok {
		return
	}
	var req struct {
		ProgramID string `json:"program_id"`
	}
	if err := s.parseJSONBody(r, &req); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	collection, err := s.publication.AddProgramToCollection(r.Context(), user.ID, chi.URLParam(r, "collectionId"), req.ProgramID)
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"collection": collection})
}

func (s *Server) handleUserCollectionRemoveProgram(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.requirePublicationMutationUser(w, r)
	if !ok {
		return
	}
	collection, err := s.publication.RemoveProgramFromCollection(r.Context(), user.ID, chi.URLParam(r, "collectionId"), chi.URLParam(r, "programId"))
	if err != nil {
		s.writeUserCatalogError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"collection": collection})
}

func (s *Server) requirePublicationUser(w stdhttp.ResponseWriter, r *stdhttp.Request) (auth.User, bool) {
	if s.publication == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Publication service is unavailable.")
		return auth.User{}, false
	}
	user, ok := authUserFromContext(r.Context())
	if !ok {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return auth.User{}, false
	}
	return user, true
}

func (s *Server) requirePublicationMutationUser(w stdhttp.ResponseWriter, r *stdhttp.Request) (auth.User, bool) {
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return auth.User{}, false
	}
	return s.requirePublicationUser(w, r)
}

func (s *Server) writeUserCatalogError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	switch {
	case errors.Is(err, publication.ErrProgramNotAvailable), errors.Is(err, publication.ErrCollectionNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "resource_not_found", "Resource not found.")
	case errors.Is(err, publication.ErrInvalidCollection):
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_collection", "Collection input is invalid.")
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Request could not be completed.")
	}
}
