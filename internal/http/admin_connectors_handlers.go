package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/connectors"
)

const (
	adminUploadMaxBodyBytes = 25 * 1024 * 1024
)

func (s *Server) handleAdminConnectors(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
		return
	}
	items, err := s.connectors.ListConnectors(r.Context())
	if err != nil {
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "获取 Connector 列表失败。")
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"connectors": items})
}

func (s *Server) handleAdminConnectorUpload(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
		return
	}
	if err := s.validateCSRFFromCookie(r); err != nil {
		s.writeAuthError(w, r, err)
		return
	}
	if !strings.Contains(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))), "multipart/form-data") {
		writeError(w, r, stdhttp.StatusUnsupportedMediaType, "unsupported_content_type", "上传必须使用 multipart/form-data。")
		return
	}
	r.Body = stdhttp.MaxBytesReader(w, r.Body, adminUploadMaxBodyBytes)
	if err := r.ParseMultipartForm(adminUploadMaxBodyBytes); err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_upload", "上传内容无效或超过大小限制。")
		return
	}
	connectorID := strings.TrimSpace(r.FormValue("connector_id"))
	version := strings.TrimSpace(r.FormValue("version"))
	file, header, err := r.FormFile("package")
	if err != nil {
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_upload", "缺少 package 文件。")
		return
	}
	defer file.Close()
	user, ok := authUserFromContext(r.Context())
	if !ok {
		s.writeAuthError(w, r, auth.ErrNotAuthenticated)
		return
	}
	userID := user.ID
	result, uploadErr := s.connectors.Upload(r.Context(), connectors.UploadInput{
		ConnectorID: connectorID,
		Version:     version,
		PackageName: header.Filename,
		Content:     file,
		UploadedBy:  &userID,
	})
	if uploadErr != nil {
		s.writeConnectorError(w, r, uploadErr)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{
		"connector":          result.Connector,
		"version":            sanitizeConnectorVersion(result.Version),
		"validation_summary": result.Summary,
		"note":               "仅登记与静态校验，未执行 Connector。",
	})
}

func (s *Server) handleAdminConnector(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
		return
	}
	connectorID := chi.URLParam(r, "connectorId")
	connector, err := s.connectors.GetConnector(r.Context(), connectorID)
	if err != nil {
		s.writeConnectorError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"connector": connector})
}

func (s *Server) handleAdminConnectorVersions(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
		return
	}
	connectorID := chi.URLParam(r, "connectorId")
	versions, err := s.connectors.ListVersions(r.Context(), connectorID)
	if err != nil {
		s.writeConnectorError(w, r, err)
		return
	}
	sanitized := make([]map[string]any, 0, len(versions))
	for _, version := range versions {
		sanitized = append(sanitized, sanitizeConnectorVersion(version))
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"versions": sanitized})
}

func (s *Server) handleAdminConnectorVersion(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
		return
	}
	versionID := chi.URLParam(r, "versionId")
	version, err := s.connectors.GetVersion(r.Context(), versionID)
	if err != nil {
		s.writeConnectorError(w, r, err)
		return
	}
	var validationSummary any
	_ = json.Unmarshal([]byte(version.ValidationSummaryJSON), &validationSummary)
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"version":            sanitizeConnectorVersion(version),
		"validation_summary": validationSummary,
	})
}

func (s *Server) handleAdminConnectorVersionApprove(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleConnectorVersionStateChange(w, r, "approve")
}

func (s *Server) handleAdminConnectorVersionReject(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleConnectorVersionStateChange(w, r, "reject")
}

func (s *Server) handleAdminConnectorVersionDisable(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleConnectorVersionStateChange(w, r, "disable")
}

func (s *Server) handleConnectorVersionStateChange(w stdhttp.ResponseWriter, r *stdhttp.Request, action string) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
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
	versionID := chi.URLParam(r, "versionId")
	actorID := user.ID
	var (
		version connectors.ConnectorVersion
		err     error
	)
	switch action {
	case "approve":
		version, err = s.connectors.ApproveVersion(r.Context(), versionID, &actorID)
	case "reject":
		version, err = s.connectors.RejectVersion(r.Context(), versionID, &actorID)
	case "disable":
		version, err = s.connectors.DisableVersion(r.Context(), versionID, &actorID)
	default:
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_action", "无效操作。")
		return
	}
	if err != nil {
		s.writeConnectorError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"version": sanitizeConnectorVersion(version)})
}

func (s *Server) handleAdminConnectorDisable(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleConnectorStatusChange(w, r, connectors.ConnectorStatusDisabled)
}

func (s *Server) handleAdminConnectorEnable(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.handleConnectorStatusChange(w, r, connectors.ConnectorStatusActive)
}

func (s *Server) handleConnectorStatusChange(w stdhttp.ResponseWriter, r *stdhttp.Request, status connectors.ConnectorStatus) {
	if s.connectors == nil {
		writeError(w, r, stdhttp.StatusServiceUnavailable, "temporarily_unavailable", "Connector 服务尚未启用。")
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
	connectorID := chi.URLParam(r, "connectorId")
	actorID := user.ID
	connector, err := s.connectors.SetConnectorStatus(r.Context(), connectorID, status, &actorID)
	if err != nil {
		s.writeConnectorError(w, r, err)
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{"connector": connector})
}

func (s *Server) writeConnectorError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	switch {
	case errors.Is(err, connectors.ErrInvalidConnectorID):
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_connector_id", "Connector id 无效。")
	case errors.Is(err, connectors.ErrInvalidVersion):
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_version", "Connector version 无效。")
	case errors.Is(err, connectors.ErrInvalidUpload):
		writeError(w, r, stdhttp.StatusBadRequest, "invalid_upload", "上传内容无效。")
	case errors.Is(err, connectors.ErrConnectorNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "connector_not_found", "Connector 不存在。")
	case errors.Is(err, connectors.ErrVersionNotFound):
		writeError(w, r, stdhttp.StatusNotFound, "connector_version_not_found", "Connector 版本不存在。")
	case errors.Is(err, connectors.ErrVersionAlreadyExists):
		writeError(w, r, stdhttp.StatusConflict, "version_already_exists", "该 Connector 下版本已存在，不可覆盖。")
	case errors.Is(err, connectors.ErrConnectorDisabled):
		writeError(w, r, stdhttp.StatusConflict, "connector_disabled", "Connector 已禁用，不能启用其版本。")
	case errors.Is(err, connectors.ErrVersionNotPendingReview):
		writeError(w, r, stdhttp.StatusConflict, "invalid_review_state", "版本当前状态不允许此操作。")
	case errors.Is(err, connectors.ErrVersionNotApproved):
		writeError(w, r, stdhttp.StatusConflict, "invalid_review_state", "仅 approved 版本可执行 disable。")
	default:
		writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "Connector 请求暂时无法完成。")
	}
}

func sanitizeConnectorVersion(version connectors.ConnectorVersion) map[string]any {
	manifest := map[string]any{}
	_ = json.Unmarshal([]byte(version.ManifestJSON), &manifest)
	return map[string]any{
		"id":                      version.ID,
		"connector_id":            version.ConnectorID,
		"version":                 version.Version,
		"review_status":           version.ReviewStatus,
		"runtime_profile":         version.RuntimeProfile,
		"entrypoint":              version.Entrypoint,
		"manifest":                manifest,
		"package_sha256":          version.PackageSHA256,
		"package_size_bytes":      version.PackageSizeBytes,
		"validation_summary_json": version.ValidationSummaryJSON,
		"uploaded_by":             version.UploadedBy,
		"reviewed_by":             version.ReviewedBy,
		"reviewed_at":             version.ReviewedAt,
		"created_at":              version.CreatedAt,
	}
}
