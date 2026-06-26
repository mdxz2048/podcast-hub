package runner

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

var (
	ErrArtifactPathEscape  = errors.New("artifact path escapes output directory")
	ErrArtifactInvalidType = errors.New("artifact is not a regular file")
	ErrArtifactTooMany     = errors.New("too many artifacts")
	ErrArtifactTooLarge    = errors.New("artifact size exceeds maximum")
	ErrArtifactDuplicate   = errors.New("duplicate artifact")
	ErrArtifactUndeclared  = errors.New("undeclared artifact")
)

type ArtifactLimits struct {
	MaxArtifacts       int
	MaxSingleFileBytes int64
	MaxTotalBytes      int64
}

func DefaultArtifactLimits() ArtifactLimits {
	return ArtifactLimits{MaxArtifacts: 100, MaxSingleFileBytes: 50 * 1024 * 1024, MaxTotalBytes: 250 * 1024 * 1024}
}

func ValidateArtifacts(jobID string, outputDir string, declared []DeclaredArtifact, limits ArtifactLimits) ([]jobs.ImportJobArtifact, error) {
	if limits.MaxArtifacts <= 0 || limits.MaxSingleFileBytes <= 0 || limits.MaxTotalBytes <= 0 {
		limits = DefaultArtifactLimits()
	}
	if len(declared) > limits.MaxArtifacts {
		return nil, ErrArtifactTooMany
	}
	seen := map[string]struct{}{}
	total := int64(0)
	artifacts := make([]jobs.ImportJobArtifact, 0, len(declared))
	for _, declaration := range declared {
		rel, err := cleanArtifactPath(declaration.RelativePath)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[rel]; exists {
			return nil, ErrArtifactDuplicate
		}
		seen[rel] = struct{}{}
		fullPath := filepath.Join(outputDir, rel)
		info, err := os.Lstat(fullPath)
		if err != nil {
			return nil, fmt.Errorf("stat artifact: %w", err)
		}
		if !info.Mode().IsRegular() {
			return nil, ErrArtifactInvalidType
		}
		if info.Size() > limits.MaxSingleFileBytes {
			return nil, ErrArtifactTooLarge
		}
		total += info.Size()
		if total > limits.MaxTotalBytes {
			return nil, ErrArtifactTooLarge
		}
		sum, err := sha256File(fullPath)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, jobs.ImportJobArtifact{
			ID:           uuid.NewString(),
			ImportJobID:  jobID,
			ArtifactType: declaration.ArtifactType,
			RelativePath: rel,
			SizeBytes:    info.Size(),
			SHA256:       sum,
			CreatedAt:    time.Now(),
		})
	}
	if err := rejectUndeclaredFiles(outputDir, seen); err != nil {
		return nil, err
	}
	return artifacts, nil
}

func cleanArtifactPath(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || filepath.IsAbs(value) {
		return "", ErrArtifactPathEscape
	}
	cleaned := filepath.Clean(value)
	if cleaned == "." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return "", ErrArtifactPathEscape
	}
	return cleaned, nil
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open artifact: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash artifact: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func rejectUndeclaredFiles(outputDir string, declared map[string]struct{}) error {
	return filepath.WalkDir(outputDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == outputDir {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			rel, err := filepath.Rel(outputDir, path)
			if err != nil {
				return err
			}
			if _, ok := declared[rel]; !ok {
				return ErrArtifactUndeclared
			}
		}
		return nil
	})
}
