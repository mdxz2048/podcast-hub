package runner

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateArtifactsRecordsHashAndSize(t *testing.T) {
	outputDir := t.TempDir()
	body := []byte("episode metadata")
	writeArtifact(t, outputDir, "episodes/episode-001.json", body)

	artifacts, err := ValidateArtifacts("job1", outputDir, []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "episodes/episode-001.json"}}, ArtifactLimits{})
	if err != nil {
		t.Fatalf("validate artifact: %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("expected one artifact, got %d", len(artifacts))
	}
	sum := sha256.Sum256(body)
	if artifacts[0].SizeBytes != int64(len(body)) || artifacts[0].SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("unexpected artifact metadata: %+v", artifacts[0])
	}
	if artifacts[0].RelativePath != "episodes/episode-001.json" {
		t.Fatalf("expected relative path only, got %s", artifacts[0].RelativePath)
	}
}

func TestValidateArtifactsRejectsUnsafeOutputs(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string)
		declared []DeclaredArtifact
		limits   ArtifactLimits
		want     error
	}{
		{name: "path escape", declared: []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "../outside.json"}}, want: ErrArtifactPathEscape},
		{name: "absolute path", declared: []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "/tmp/outside.json"}}, want: ErrArtifactPathEscape},
		{name: "symlink", setup: func(dir string) {
			writeArtifact(t, dir, "target.json", []byte("ok"))
			if err := os.Symlink(filepath.Join(dir, "target.json"), filepath.Join(dir, "link.json")); err != nil {
				t.Fatalf("symlink: %v", err)
			}
		}, declared: []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "link.json"}}, want: ErrArtifactInvalidType},
		{name: "directory", setup: func(dir string) {
			if err := os.MkdirAll(filepath.Join(dir, "folder"), 0o700); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
		}, declared: []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "folder"}}, want: ErrArtifactInvalidType},
		{name: "too many", declared: []DeclaredArtifact{{ArtifactType: "a", RelativePath: "one.json"}, {ArtifactType: "a", RelativePath: "two.json"}}, limits: ArtifactLimits{MaxArtifacts: 1, MaxSingleFileBytes: 100, MaxTotalBytes: 100}, want: ErrArtifactTooMany},
		{name: "too large", setup: func(dir string) {
			writeArtifact(t, dir, "big.json", []byte("0123456789"))
		}, declared: []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "big.json"}}, limits: ArtifactLimits{MaxArtifacts: 10, MaxSingleFileBytes: 2, MaxTotalBytes: 100}, want: ErrArtifactTooLarge},
		{name: "undeclared", setup: func(dir string) {
			writeArtifact(t, dir, "extra.json", []byte("hidden"))
		}, declared: nil, want: ErrArtifactUndeclared},
		{name: "duplicate", setup: func(dir string) {
			writeArtifact(t, dir, "dup.json", []byte("ok"))
		}, declared: []DeclaredArtifact{{ArtifactType: "episode_metadata", RelativePath: "dup.json"}, {ArtifactType: "episode_metadata", RelativePath: "dup.json"}}, want: ErrArtifactDuplicate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()
			if tt.setup != nil {
				tt.setup(outputDir)
			}
			_, err := ValidateArtifacts("job1", outputDir, tt.declared, tt.limits)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func writeArtifact(t *testing.T, outputDir string, rel string, body []byte) {
	t.Helper()
	path := filepath.Join(outputDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir artifact: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}
