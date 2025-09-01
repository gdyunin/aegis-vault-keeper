package filestorage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DirectoryPermission is the permission for creating directories.
	DirectoryPermission = 0o750
	// FilePermission is the permission for creating files.
	FilePermission = 0o600
)

// rawSave creates a function that performs raw filesystem save operations.
func rawSave(basePath string) saveFunc {
	return func(ctx context.Context, p SaveParams) error {
		userDir := filepath.Join(basePath, p.UserID.String())
		if err := os.MkdirAll(userDir, DirectoryPermission); err != nil {
			return fmt.Errorf("failed to create user directory: %w", err)
		}

		normalizedKey := normalizeStorageKey(p.StorageKey)
		fullPath := filepath.Join(userDir, normalizedKey)

		if !strings.HasPrefix(fullPath, userDir) {
			return errors.New("invalid storage key: path traversal detected")
		}

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, DirectoryPermission); err != nil {
			return fmt.Errorf("failed to create file directory: %w", err)
		}

		if err := os.WriteFile(fullPath, p.Data, FilePermission); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		return nil
	}
}

// rawLoad creates a function that performs raw filesystem load operations.
func rawLoad(basePath string) func(ctx context.Context, p LoadParams) ([]byte, error) {
	return func(ctx context.Context, p LoadParams) ([]byte, error) {
		userDir := filepath.Join(basePath, p.UserID.String())
		normalizedKey := normalizeStorageKey(p.StorageKey)
		fullPath := filepath.Join(userDir, normalizedKey)

		if !strings.HasPrefix(fullPath, userDir) {
			return nil, errors.New("invalid storage key: path traversal detected")
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %w", err)
			}
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		return data, nil
	}
}

// rawDelete creates a function that performs raw filesystem delete operations.
func rawDelete(basePath string) func(ctx context.Context, p DeleteParams) error {
	return func(ctx context.Context, p DeleteParams) error {
		userDir := filepath.Join(basePath, p.UserID.String())
		normalizedKey := normalizeStorageKey(p.StorageKey)
		fullPath := filepath.Join(userDir, normalizedKey)

		if !strings.HasPrefix(fullPath, userDir) {
			return errors.New("invalid storage key: path traversal detected")
		}

		if err := os.Remove(fullPath); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("failed to delete file: %w", err)
		}

		// Try to remove empty directories (best effort)
		dir := filepath.Dir(fullPath)
		for dir != userDir && dir != "." {
			if err := os.Remove(dir); err != nil {
				// Directory is not empty or other error, stop cleanup but don't fail
				// This is expected behavior for non-empty directories
				break // nolint:nilerr // Ignoring error in cleanup is intentional
			}
			dir = filepath.Dir(dir)
		}

		return nil // nolint:nilerr // Ignoring error in cleanup is intentional
	}
}

// normalizeStorageKey sanitizes storage keys to prevent path traversal attacks.
func normalizeStorageKey(key string) string {
	key = strings.ReplaceAll(key, `\`, `/`)
	key = strings.TrimPrefix(key, "./")
	key = strings.TrimPrefix(key, "/")
	return filepath.Clean(key)
}
