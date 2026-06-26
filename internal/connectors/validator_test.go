package connectors

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"
)

func TestValidatePackageZipValid(t *testing.T) {
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(validManifestYAML)},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	if !result.Summary.IsValid {
		t.Fatalf("expected valid package, got issues: %+v", result.Summary.Issues)
	}
}

func TestValidatePackageZipRejectsZipSlip(t *testing.T) {
	zipData := buildZip(t, []zipEntry{
		{name: "../manifest.yaml", body: []byte(validManifestYAML)},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "zip_path_invalid")
}

func TestValidatePackageZipRejectsSymlink(t *testing.T) {
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(validManifestYAML)},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
		{name: "src/link.py", body: []byte("ignored"), mode: os.ModeSymlink},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "zip_symlink_forbidden")
}

func TestValidatePackageZipRejectsCompressionBomb(t *testing.T) {
	limits := DefaultValidationLimits()
	limits.MaxCompressionRatio = 2
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(validManifestYAML)},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: bytes.Repeat([]byte("A"), 1024)},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", limits)
	assertHasIssue(t, result.Summary, "zip_compression_ratio_too_high")
}

func TestValidatePackageZipManifestUnknownFieldRejected(t *testing.T) {
	manifest := validManifestYAML + "\nunknown_root_field: true\n"
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(manifest)},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "manifest_invalid")
}

func TestValidatePackageZipMissingRequirementsRejected(t *testing.T) {
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: []byte(validManifestYAML)},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "requirements_lock_missing")
}

func TestValidatePackageZipEntrypointNotFoundRejected(t *testing.T) {
	manifest := bytes.ReplaceAll([]byte(validManifestYAML), []byte("src/connector.py"), []byte("src/missing.py"))
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: manifest},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "entrypoint_not_found")
}

func TestValidatePackageZipRejectsNonPythonRuntime(t *testing.T) {
	manifest := bytes.ReplaceAll([]byte(validManifestYAML), []byte("language: python"), []byte("language: go"))
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: manifest},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "runtime_language_invalid")
}

func TestValidatePackageZipRejectsInvalidAuthTriggerExecutionCombo(t *testing.T) {
	manifest := bytes.ReplaceAll([]byte(validManifestYAML), []byte("mode: none"), []byte("mode: qr_each_run"))
	manifest = bytes.ReplaceAll(manifest, []byte("- manual"), []byte("- scheduled"))
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: manifest},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "qr_scheduled_forbidden")
}

func TestValidatePackageZipRejectsInvalidNetworkAllowlist(t *testing.T) {
	manifest := bytes.ReplaceAll([]byte(validManifestYAML), []byte("- api.example.invalid"), []byte("- https://api.example.invalid/path"))
	zipData := buildZip(t, []zipEntry{
		{name: "manifest.yaml", body: manifest},
		{name: "requirements.lock", body: []byte("pkg==1.0.0")},
		{name: "README.md", body: []byte("# readme")},
		{name: "src/connector.py", body: []byte("print('ok')")},
	})
	result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
	assertHasIssue(t, result.Summary, "network_host_invalid")
}

func TestValidatePackageZipRejectsForbiddenFiles(t *testing.T) {
	for _, entry := range []zipEntry{
		{name: "Dockerfile", body: []byte("FROM python:3.12")},
		{name: "cookies.json", body: []byte("{}")},
		{name: "media/audio.mp3", body: []byte("x")},
		{name: "src/bin.py", body: []byte{0x00, 0x01}},
	} {
		zipData := buildZip(t, []zipEntry{
			{name: "manifest.yaml", body: []byte(validManifestYAML)},
			{name: "requirements.lock", body: []byte("pkg==1.0.0")},
			{name: "README.md", body: []byte("# readme")},
			{name: "src/connector.py", body: []byte("print('ok')")},
			entry,
		})
		result := ValidatePackageZip(zipData, "example-connector", "1.0.0", DefaultValidationLimits())
		if result.Summary.IsValid {
			t.Fatalf("expected forbidden file package to be invalid: %s", entry.name)
		}
	}
}

type zipEntry struct {
	name string
	body []byte
	mode os.FileMode
}

func buildZip(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		fw, err := w.CreateHeader(header)
		if err != nil {
			t.Fatalf("create zip header: %v", err)
		}
		if len(entry.body) > 0 {
			if _, err := fw.Write(entry.body); err != nil {
				t.Fatalf("write zip body: %v", err)
			}
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func assertHasIssue(t *testing.T, summary ValidationSummary, code string) {
	t.Helper()
	for _, issue := range summary.Issues {
		if issue.Code == code {
			return
		}
	}
	t.Fatalf("expected issue %s, got %+v", code, summary.Issues)
}

const validManifestYAML = `spec_version: 1
id: example-connector
name: 示例连接器
version: 1.0.0
runtime:
  language: python
  profile: python-basic
  entrypoint: src/connector.py
ingestion_type: connector
trigger:
  allowed:
    - manual
auth:
  mode: none
execution:
  mode: unattended
  timeout_seconds: 900
  memory_mb: 512
  max_download_size_mb: 2048
inputs: []
secrets: []
network:
  allowlist:
    - api.example.invalid
outputs:
  type: podcast_episode_bundle
`
