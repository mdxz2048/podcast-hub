package media

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStore struct {
	stagingRoot   string
	publishedRoot string
}

func NewLocalStore(stagingRoot string, publishedRoot string) *LocalStore {
	return &LocalStore{stagingRoot: stagingRoot, publishedRoot: publishedRoot}
}

func (s *LocalStore) Promote(_ context.Context, stagedKey string, publishedKey string) error {
	sourcePath, err := safePath(s.stagingRoot, stagedKey)
	if err != nil {
		return err
	}
	targetPath, err := safePath(s.publishedRoot, publishedKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create media target directory: %w", err)
	}
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("published media already exists")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check published media destination: %w", err)
	}
	if err := os.Rename(sourcePath, targetPath); err == nil {
		return nil
	}
	if err := copyFile(sourcePath, targetPath); err != nil {
		return err
	}
	if err := os.Remove(sourcePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove staged media after copy: %w", err)
	}
	return nil
}

func (s *LocalStore) OpenPublished(_ context.Context, publishedKey string) (*os.File, error) {
	path, err := safePath(s.publishedRoot, publishedKey)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open published media: %w", err)
	}
	return file, nil
}

func (s *LocalStore) DeletePublished(_ context.Context, publishedKey string) error {
	path, err := safePath(s.publishedRoot, publishedKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete published media: %w", err)
	}
	return nil
}

func safePath(root string, storageKey string) (string, error) {
	if filepath.IsAbs(storageKey) {
		return "", fmt.Errorf("media storage key must be relative")
	}
	cleanKey := filepath.Clean(filepath.FromSlash(storageKey))
	if cleanKey == "." || cleanKey == ".." || cleanKey == "" {
		return "", fmt.Errorf("media storage key invalid")
	}
	if strings.HasPrefix(cleanKey, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("media storage key escapes root")
	}
	path := filepath.Join(root, cleanKey)
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve media root: %w", err)
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve media path: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return "", fmt.Errorf("relativize media path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("media path escapes storage root")
	}
	return path, nil
}

func copyFile(sourcePath string, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open staged media: %w", err)
	}
	defer source.Close()
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create published media: %w", err)
	}
	defer func() {
		_ = target.Close()
	}()
	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy staged media: %w", err)
	}
	if err := target.Close(); err != nil {
		return fmt.Errorf("close published media: %w", err)
	}
	return nil
}
