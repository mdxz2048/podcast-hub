package http

import (
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

const adminSecretMaxBodyBytes = sources.MaxSecretBytes + 4096

func (s *Server) handleAdminSources(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
		return
	}
	items, err := s.sources.ListSources(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "获取 Source 列表失败。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"sources": items})
}

func marshalConfig(config map[string]any) (string, error) {
	if config == nil {
		return "{}", nil
	}
	body, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type sourcePayload struct {
	ConnectorVersionID string         `json:"connector_version_id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	TriggerType        string         `json:"trigger_type"`
	AuthMode           string         `json:"auth_mode"`
	ExecutionMode      string         `json:"execution_mode"`
	Config             map[string]any `json:"config"`
	NetworkMode        string         `json:"network_mode"`
}

func (s *Server) handleAdminSourceCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	var payload sourcePayload
	if err := s.parseJSONBody(r, &payload); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	configJSON, err := marshalConfig(payload.Config)
	if err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_source_input", "Source config 必须是 JSON 对象。")
		return
	}
	userID := user.ID
	detail, err := s.sources.CreateSource(r.Context(), sources.CreateSourceInput{
		ConnectorVersionID: strings.TrimSpace(payload.ConnectorVersionID),
		Name:               payload.Name,
		Description:        payload.Description,
		TriggerType:        payload.TriggerType,
		AuthMode:           payload.AuthMode,
		ExecutionMode:      payload.ExecutionMode,
		ConfigJSON:         configJSON,
		NetworkMode:        payload.NetworkMode,
		CreatedBy:          &userID,
	})
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, detail)
}

func (s *Server) handleAdminSource(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
		return
	}
	detail, err := s.sources.GetSourceDetail(r.Context(), chi.URLParam(r, "sourceId"))
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, detail)
}

func (s *Server) handleAdminSourceUpdate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	var payload sourcePayload
	if err := s.parseJSONBody(r, &payload); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	configJSON, err := marshalConfig(payload.Config)
	if err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_source_input", "Source config 必须是 JSON 对象。")
		return
	}
	userID := user.ID
	detail, err := s.sources.UpdateSource(r.Context(), chi.URLParam(r, "sourceId"), sources.UpdateSourceInput{
		Name:        strings.TrimSpace(payload.Name),
		Description: strings.TrimSpace(payload.Description),
		ConfigJSON:  configJSON,
		NetworkMode: payload.NetworkMode,
	}, &userID)
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, detail)
}

func (s *Server) handleAdminSourceEnable(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleAdminSourceStatus(w, r, true)
}

func (s *Server) handleAdminSourceDisable(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleAdminSourceStatus(w, r, false)
}

func (s *Server) handleAdminSourceStatus(w stdhttp.ResponseWriter, r *stdhttp.Request, enable bool) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	var (
		detail sources.ConnectorSourceDetail
		err    error
	)
	if enable {
		detail, err = s.sources.EnableSource(r.Context(), chi.URLParam(r, "sourceId"), &userID)
	} else {
		detail, err = s.sources.DisableSource(r.Context(), chi.URLParam(r, "sourceId"), &userID)
	}
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, detail)
}

func (s *Server) handleAdminSecrets(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
		return
	}
	items, err := s.sources.ListSecrets(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "获取 Secret 列表失败。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"secrets": items})
}

func (s *Server) handleAdminSecretTextCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	var payload struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := s.parseJSONBody(r, &payload); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	userID := user.ID
	secret, err := s.sources.CreateSecret(r.Context(), payload.Name, sources.SecretTypeText, []byte(payload.Value), &userID)
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"secret": secret})
}

func (s *Server) handleAdminSecretFileCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	r.Body = stdhttp.MaxBytesReader(w, r.Body, adminSecretMaxBodyBytes)
	if err := r.ParseMultipartForm(adminSecretMaxBodyBytes); err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_secret_input", "Secret 文件无效或超过大小限制。")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_secret_input", "缺少 Secret 文件。")
		return
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, sources.MaxSecretBytes+1))
	if err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_secret_input", "读取 Secret 文件失败。")
		return
	}
	userID := user.ID
	secret, err := s.sources.CreateSecret(r.Context(), name, sources.SecretTypeFile, payload, &userID)
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"secret": secret})
}

func (s *Server) handleAdminSecretRevoke(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	secret, err := s.sources.RevokeSecret(r.Context(), chi.URLParam(r, "secretId"), &userID)
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"secret": secret})
}

func (s *Server) handleAdminSourceSecretBind(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	var payload struct {
		SecretName string `json:"secret_name"`
		SecretID   string `json:"secret_record_id"`
	}
	if err := s.parseJSONBody(r, &payload); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	userID := user.ID
	detail, err := s.sources.BindSecret(r.Context(), chi.URLParam(r, "sourceId"), payload.SecretName, payload.SecretID, &userID)
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, detail)
}

func (s *Server) handleAdminSourceSecretUnbind(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.sources == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Source 服务尚未启用。")
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
	detail, err := s.sources.DeleteBinding(r.Context(), chi.URLParam(r, "sourceId"), chi.URLParam(r, "bindingId"), &userID)
	if err != nil {
		s.writeSourceError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, detail)
}

func (s *Server) writeSourceError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	switch {
	case errors.Is(err, sources.ErrSourceNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "source_not_found", "Source 不存在。")
	case errors.Is(err, sources.ErrSecretNotFound), errors.Is(err, sources.ErrBindingNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "secret_not_found", "Secret 或绑定不存在。")
	case errors.Is(err, sources.ErrUnsupportedAlphaMode):
		writeError(w, r, stdhttp.StatusBadRequest, "unsupported_alpha_mode", "该 Source 能力将在 M3 interactive / QR Connector 阶段提供。")
	case errors.Is(err, sources.ErrConnectorUnavailable), errors.Is(err, sources.ErrConnectorVersionInvalid):
		writeError(w, r, stdhttp.StatusConflict, "connector_unavailable", "只能使用 active Connector 下 approved 且未 disabled 的版本创建 Source。")
	case errors.Is(err, sources.ErrMissingRequiredSecrets):
		writeError(w, r, stdhttp.StatusConflict, "missing_required_secrets", "Source 缺少 manifest 声明的 required Secret 绑定。")
	case errors.Is(err, sources.ErrSecretRevoked):
		writeError(w, r, stdhttp.StatusConflict, "secret_revoked", "已撤销的 Secret 不能用于启用 Source。")
	case errors.Is(err, sources.ErrSecretTooLarge):
		writeError(w, r, stdhttp.StatusBadRequest, "secret_too_large", "Secret 内容为空或超过大小限制。")
	case errors.Is(err, sources.ErrInvalidInput):
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_source_input", "Source 或 Secret 输入无效。")
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Source 请求暂时无法完成。")
	}
}
