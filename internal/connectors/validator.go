package connectors

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	slugPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}$`)
	semverPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)
	hostPattern   = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*)$`)
)

type ValidationLimits struct {
	MaxZipSizeBytes       int64
	MaxFileCount          int
	MaxSingleFileBytes    int64
	MaxUncompressedBytes  int64
	MaxCompressionRatio   float64
	MaxTimeoutSeconds     int
	MaxMemoryMB           int
	MaxDownloadSizeMB     int
	AllowedRuntimeProfile []string
}

func DefaultValidationLimits() ValidationLimits {
	return ValidationLimits{
		MaxZipSizeBytes:       25 * 1024 * 1024,
		MaxFileCount:          200,
		MaxSingleFileBytes:    5 * 1024 * 1024,
		MaxUncompressedBytes:  50 * 1024 * 1024,
		MaxCompressionRatio:   100,
		MaxTimeoutSeconds:     3600,
		MaxMemoryMB:           4096,
		MaxDownloadSizeMB:     8192,
		AllowedRuntimeProfile: []string{"python-basic", "python-telegram"},
	}
}

type PackageValidationResult struct {
	Manifest         ConnectorManifest
	ManifestJSON     string
	Summary          ValidationSummary
	NormalizedPaths  map[string]struct{}
	RequirementsPath string
	ReadmePath       string
}

func ValidatePackageZip(zipBytes []byte, connectorID string, version string, limits ValidationLimits) PackageValidationResult {
	result := PackageValidationResult{
		Summary:         ValidationSummary{IsValid: true},
		NormalizedPaths: map[string]struct{}{},
	}
	addIssue := func(code string, message string, p string) {
		result.Summary.IsValid = false
		result.Summary.Issues = append(result.Summary.Issues, ValidationIssue{Code: code, Message: message, Path: p})
	}
	if !slugPattern.MatchString(connectorID) {
		addIssue("connector_id_invalid", "connector id 必须为小写字母数字与短横线。", "connector_id")
	}
	if !semverPattern.MatchString(version) {
		addIssue("version_invalid", "version 必须为 semver。", "version")
	}
	if int64(len(zipBytes)) > limits.MaxZipSizeBytes {
		addIssue("zip_too_large", "ZIP 包超过允许大小。", "")
		return result
	}
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		addIssue("zip_invalid", "ZIP 文件无效。", "")
		return result
	}
	if len(reader.File) == 0 {
		addIssue("zip_empty", "ZIP 不能为空。", "")
		return result
	}
	if len(reader.File) > limits.MaxFileCount {
		addIssue("zip_too_many_files", "ZIP 文件数量超过限制。", "")
	}

	var manifestBytes []byte
	var totalUncompressed int64
	for _, file := range reader.File {
		cleanPath, pathOK := normalizeArchivePath(file.Name)
		if !pathOK {
			addIssue("zip_path_invalid", "ZIP 路径不安全。", file.Name)
			continue
		}
		if _, exists := result.NormalizedPaths[cleanPath]; exists {
			addIssue("zip_duplicate_path", "ZIP 中存在重复路径。", cleanPath)
			continue
		}
		result.NormalizedPaths[cleanPath] = struct{}{}
		if file.FileInfo().IsDir() {
			continue
		}
		mode := file.Mode()
		if mode&os.ModeSymlink != 0 {
			addIssue("zip_symlink_forbidden", "ZIP 中不允许 symlink。", cleanPath)
			continue
		}
		if !mode.IsRegular() {
			addIssue("zip_special_file_forbidden", "ZIP 中仅允许普通文件。", cleanPath)
			continue
		}
		if file.UncompressedSize64 > uint64(limits.MaxSingleFileBytes) {
			addIssue("file_too_large", "单文件超过限制。", cleanPath)
		}
		totalUncompressed += int64(file.UncompressedSize64)
		if totalUncompressed > limits.MaxUncompressedBytes {
			addIssue("zip_uncompressed_too_large", "解压后总大小超过限制。", "")
		}
		if file.CompressedSize64 > 0 {
			ratio := float64(file.UncompressedSize64) / float64(file.CompressedSize64)
			if ratio > limits.MaxCompressionRatio {
				addIssue("zip_compression_ratio_too_high", "压缩比超过安全限制。", cleanPath)
			}
		}
		if file.UncompressedSize64 > 0 && isForbiddenPath(cleanPath) {
			addIssue("forbidden_file", "ZIP 包含禁止文件或目录。", cleanPath)
		}
		if !isAllowedExtension(cleanPath) {
			addIssue("forbidden_extension", "文件扩展名不在允许列表。", cleanPath)
		}
		body, readErr := readZipFile(file)
		if readErr != nil {
			addIssue("zip_read_failed", "读取 ZIP 内容失败。", cleanPath)
			continue
		}
		if looksBinary(body) {
			addIssue("binary_forbidden", "仅允许文本类型文件。", cleanPath)
		}
		if strings.EqualFold(path.Base(cleanPath), "manifest.yaml") {
			manifestBytes = body
		}
		if strings.EqualFold(path.Base(cleanPath), "requirements.lock") {
			result.RequirementsPath = cleanPath
		}
		if strings.EqualFold(path.Base(cleanPath), "readme.md") {
			result.ReadmePath = cleanPath
		}
	}

	if len(manifestBytes) == 0 {
		addIssue("manifest_missing", "缺少 manifest.yaml。", "manifest.yaml")
		return result
	}
	if result.RequirementsPath == "" {
		addIssue("requirements_lock_missing", "缺少 requirements.lock。", "requirements.lock")
	}
	if result.ReadmePath == "" {
		addIssue("readme_missing", "缺少 README.md。", "README.md")
	}

	manifest, manifestJSON, manifestIssues := validateManifest(manifestBytes, connectorID, version, limits, result.NormalizedPaths)
	result.Manifest = manifest
	result.ManifestJSON = manifestJSON
	result.Summary.Issues = append(result.Summary.Issues, manifestIssues...)
	if len(manifestIssues) > 0 {
		result.Summary.IsValid = false
	}
	return result
}

func validateManifest(raw []byte, connectorID string, version string, limits ValidationLimits, paths map[string]struct{}) (ConnectorManifest, string, []ValidationIssue) {
	issues := make([]ValidationIssue, 0)
	addIssue := func(code string, message string, p string) {
		issues = append(issues, ValidationIssue{Code: code, Message: message, Path: p})
	}

	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	var manifest ConnectorManifest
	if err := decoder.Decode(&manifest); err != nil {
		addIssue("manifest_invalid", "manifest.yaml 结构无效或包含未知字段。", "manifest.yaml")
		return ConnectorManifest{}, "", issues
	}
	if manifest.SpecVersion != 1 {
		addIssue("manifest_spec_version_invalid", "spec_version 必须为 1。", "spec_version")
	}
	if manifest.ID != connectorID {
		addIssue("manifest_id_mismatch", "manifest id 必须与 API connector_id 一致。", "id")
	}
	if manifest.Version != version {
		addIssue("manifest_version_mismatch", "manifest version 必须与 API version 一致。", "version")
	}
	if manifest.Runtime.Language != "python" {
		addIssue("runtime_language_invalid", "runtime.language 必须为 python。", "runtime.language")
	}
	if !slices.Contains(limits.AllowedRuntimeProfile, manifest.Runtime.Profile) {
		addIssue("runtime_profile_invalid", "runtime.profile 不在允许列表。", "runtime.profile")
	}
	if manifest.IngestionType != "connector" {
		addIssue("ingestion_type_invalid", "ingestion_type 必须为 connector。", "ingestion_type")
	}
	if manifest.Runtime.Entrypoint == "" {
		addIssue("entrypoint_missing", "runtime.entrypoint 必填。", "runtime.entrypoint")
	} else {
		normalizedEntrypoint, ok := normalizeArchivePath(manifest.Runtime.Entrypoint)
		if !ok {
			addIssue("entrypoint_invalid", "runtime.entrypoint 路径非法。", "runtime.entrypoint")
		} else if _, exists := paths[normalizedEntrypoint]; !exists {
			addIssue("entrypoint_not_found", "runtime.entrypoint 文件不存在。", "runtime.entrypoint")
		}
	}
	if len(manifest.Trigger.Allowed) == 0 {
		addIssue("trigger_missing", "trigger.allowed 不能为空。", "trigger.allowed")
	}
	hasManual := false
	hasScheduled := false
	for _, trigger := range manifest.Trigger.Allowed {
		if trigger == "manual" {
			hasManual = true
		} else if trigger == "scheduled" {
			hasScheduled = true
		} else {
			addIssue("trigger_invalid", "仅支持 manual/scheduled。", "trigger.allowed")
		}
	}
	switch manifest.Auth.Mode {
	case "none", "reusable_session":
		if manifest.Execution.Mode != "unattended" {
			addIssue("execution_mode_invalid", "auth_mode none/reusable_session 必须为 unattended。", "execution.mode")
		}
	case "qr_each_run":
		if !hasManual || hasScheduled {
			addIssue("trigger_auth_combo_invalid", "qr_each_run 仅允许 manual。", "trigger.allowed")
		}
		if manifest.Execution.Mode != "interactive" {
			addIssue("execution_mode_invalid", "qr_each_run 必须为 interactive。", "execution.mode")
		}
	default:
		addIssue("auth_mode_invalid", "auth.mode 非法。", "auth.mode")
	}
	if manifest.Auth.Mode == "qr_each_run" && hasScheduled {
		addIssue("qr_scheduled_forbidden", "scheduled 不允许 qr_each_run。", "trigger.allowed")
	}
	if manifest.Execution.TimeoutSeconds <= 0 || manifest.Execution.TimeoutSeconds > limits.MaxTimeoutSeconds {
		addIssue("timeout_invalid", "timeout_seconds 超出平台允许范围。", "execution.timeout_seconds")
	}
	if manifest.Execution.MemoryMB <= 0 || manifest.Execution.MemoryMB > limits.MaxMemoryMB {
		addIssue("memory_invalid", "memory_mb 超出平台允许范围。", "execution.memory_mb")
	}
	if manifest.Execution.MaxDownloadSizeMB <= 0 || manifest.Execution.MaxDownloadSizeMB > limits.MaxDownloadSizeMB {
		addIssue("download_size_invalid", "max_download_size_mb 超出平台允许范围。", "execution.max_download_size_mb")
	}
	for i, secret := range manifest.Secrets {
		if strings.TrimSpace(secret.Name) == "" {
			addIssue("secret_name_required", "secret.name 不能为空。", fmt.Sprintf("secrets[%d].name", i))
		}
		if strings.TrimSpace(secret.Value) != "" {
			addIssue("secret_value_forbidden", "manifest secrets 仅允许声明，禁止包含值。", fmt.Sprintf("secrets[%d].value", i))
		}
	}
	for i, host := range manifest.Network.Allowlist {
		if !isValidAllowedHost(host) {
			addIssue("network_host_invalid", "network.allowlist 仅允许主机名，不允许 scheme/path/wildcard/ip。", fmt.Sprintf("network.allowlist[%d]", i))
		}
	}
	if manifest.Outputs.Type != "podcast_episode_bundle" {
		addIssue("output_type_invalid", "outputs.type 必须为 podcast_episode_bundle。", "outputs.type")
	}
	manifestJSONBytes, err := json.Marshal(manifest)
	if err != nil {
		addIssue("manifest_json_encode_failed", "manifest 转换失败。", "manifest.yaml")
		return manifest, "", issues
	}
	return manifest, string(manifestJSONBytes), issues
}

func normalizeArchivePath(p string) (string, bool) {
	p = strings.ReplaceAll(p, "\\", "/")
	clean := path.Clean(strings.TrimSpace(p))
	if clean == "." || strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", false
	}
	return clean, true
}

func isForbiddenPath(p string) bool {
	lowerPath := strings.ToLower(p)
	base := strings.ToLower(path.Base(p))
	if strings.HasPrefix(lowerPath, ".git/") || strings.Contains(lowerPath, "/.git/") || base == ".git" {
		return true
	}
	if strings.HasPrefix(base, ".env") {
		return true
	}
	switch base {
	case "dockerfile", "docker-compose.yml":
		return true
	}
	switch filepath.Ext(base) {
	case ".pem", ".key", ".p12", ".session", ".cookie":
		return true
	}
	if strings.HasPrefix(base, "cookies") && strings.HasSuffix(base, ".json") {
		return true
	}
	if strings.HasPrefix(lowerPath, "media/") || strings.Contains(lowerPath, "/media/") ||
		strings.HasPrefix(lowerPath, "uploads/") || strings.Contains(lowerPath, "/uploads/") ||
		strings.HasPrefix(lowerPath, "downloads/") || strings.Contains(lowerPath, "/downloads/") {
		return true
	}
	if mediaExt := filepath.Ext(base); mediaExt != "" {
		switch mediaExt {
		case ".mp3", ".m4a", ".aac", ".wav", ".flac", ".ogg", ".opus", ".mp4", ".mov", ".mkv", ".webm":
			return true
		}
	}
	return false
}

func isAllowedExtension(p string) bool {
	base := path.Base(p)
	if strings.EqualFold(base, "manifest.yaml") || strings.EqualFold(base, "requirements.lock") || strings.EqualFold(base, "readme.md") {
		return true
	}
	switch strings.ToLower(filepath.Ext(base)) {
	case ".py", ".yaml", ".yml", ".json", ".txt", ".md", ".toml", ".lock":
		return true
	default:
		return false
	}
}

func looksBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	limit := len(data)
	if limit > 1024 {
		limit = 1024
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func isValidAllowedHost(host string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(host))
	if trimmed == "" || strings.Contains(trimmed, "://") || strings.Contains(trimmed, "/") || strings.Contains(trimmed, "*") || strings.Contains(trimmed, ":") {
		return false
	}
	if net.ParseIP(trimmed) != nil {
		return false
	}
	return hostPattern.MatchString(trimmed)
}
