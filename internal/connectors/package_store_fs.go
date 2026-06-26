package connectors

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalPackageStore struct {
	rootDir string
}

func NewLocalPackageStore(rootDir string) *LocalPackageStore {
	return &LocalPackageStore{rootDir: rootDir}
}

func (s *LocalPackageStore) PutQuarantine(_ context.Context, packageName string, content io.Reader) (PackageRef, error) {
	_ = packageName
	quarantineDir, err := s.safeDir("quarantine")
	if err != nil {
		return PackageRef{}, err
	}
	if err := os.MkdirAll(quarantineDir, 0o755); err != nil {
		return PackageRef{}, fmt.Errorf("create quarantine directory: %w", err)
	}
	randomName, err := randomPackageName()
	if err != nil {
		return PackageRef{}, err
	}
	storageKey := filepath.ToSlash(filepath.Join("quarantine", randomName))
	finalPath, err := s.safePath(storageKey)
	if err != nil {
		return PackageRef{}, err
	}
	tempFile, err := os.CreateTemp(quarantineDir, ".upload-*.tmp")
	if err != nil {
		return PackageRef{}, fmt.Errorf("create temporary package file: %w", err)
	}
	tempPath := tempFile.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tempFile, hasher), content)
	if err != nil {
		_ = tempFile.Close()
		return PackageRef{}, fmt.Errorf("write package file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return PackageRef{}, fmt.Errorf("close package file: %w", err)
	}
	if err := ensureWithinRoot(s.rootDir, tempPath); err != nil {
		return PackageRef{}, err
	}
	if _, err := os.Stat(finalPath); err == nil {
		return PackageRef{}, fmt.Errorf("package storage key collision")
	} else if !os.IsNotExist(err) {
		return PackageRef{}, fmt.Errorf("check package destination: %w", err)
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		return PackageRef{}, fmt.Errorf("store package atomically: %w", err)
	}
	removeTemp = false
	return PackageRef{
		StorageKey: storageKey,
		SizeBytes:  written,
		SHA256:     hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

func (s *LocalPackageStore) Read(_ context.Context, ref PackageRef) (io.ReadCloser, error) {
	path, err := s.safePath(ref.StorageKey)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open package file: %w", err)
	}
	return file, nil
}

func (s *LocalPackageStore) PromoteApproved(_ context.Context, ref PackageRef) (PackageRef, error) {
	approvedDir, err := s.safeDir("approved")
	if err != nil {
		return PackageRef{}, err
	}
	if err := os.MkdirAll(approvedDir, 0o755); err != nil {
		return PackageRef{}, fmt.Errorf("create approved directory: %w", err)
	}
	sourcePath, err := s.safePath(ref.StorageKey)
	if err != nil {
		return PackageRef{}, err
	}
	fileName := filepath.Base(ref.StorageKey)
	targetKey := filepath.ToSlash(filepath.Join("approved", fileName))
	targetPath, err := s.safePath(targetKey)
	if err != nil {
		return PackageRef{}, err
	}
	if _, err := os.Stat(targetPath); err == nil {
		return PackageRef{}, fmt.Errorf("approved package already exists")
	} else if !os.IsNotExist(err) {
		return PackageRef{}, fmt.Errorf("check approved destination: %w", err)
	}
	if err := os.Rename(sourcePath, targetPath); err != nil {
		return PackageRef{}, fmt.Errorf("promote package: %w", err)
	}
	ref.StorageKey = targetKey
	return ref, nil
}

func (s *LocalPackageStore) Delete(_ context.Context, ref PackageRef) error {
	path, err := s.safePath(ref.StorageKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete package: %w", err)
	}
	return nil
}

func (s *LocalPackageStore) safeDir(name string) (string, error) {
	return s.safePath(filepath.ToSlash(name))
}

func (s *LocalPackageStore) safePath(storageKey string) (string, error) {
	if filepath.IsAbs(storageKey) {
		return "", fmt.Errorf("package storage key must be relative")
	}
	cleanKey := filepath.Clean(filepath.FromSlash(storageKey))
	if cleanKey == "." || cleanKey == ".." || cleanKey == "" {
		return "", fmt.Errorf("package storage key invalid")
	}
	if cleanKey == ".." || len(cleanKey) >= 3 && cleanKey[:3] == "../" {
		return "", fmt.Errorf("package storage key escapes root")
	}
	path := filepath.Join(s.rootDir, cleanKey)
	if err := ensureWithinRoot(s.rootDir, path); err != nil {
		return "", err
	}
	return path, nil
}

func ensureWithinRoot(rootDir string, candidate string) error {
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("resolve package root: %w", err)
	}
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return fmt.Errorf("resolve package path: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, candidateAbs)
	if err != nil {
		return fmt.Errorf("relativize package path: %w", err)
	}
	if rel == ".." || len(rel) >= 3 && rel[:3] == "../" || filepath.IsAbs(rel) {
		return fmt.Errorf("package path escapes storage root")
	}
	return nil
}

func randomPackageName() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate package storage key: %w", err)
	}
	return hex.EncodeToString(bytes[:]) + ".zip", nil
}
