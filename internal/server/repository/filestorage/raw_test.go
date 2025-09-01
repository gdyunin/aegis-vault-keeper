package filestorage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeStorageKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean path",
			input:    "file.txt",
			expected: "file.txt",
		},
		{
			name:     "path with backslashes",
			input:    "folder\\subfolder\\file.txt",
			expected: "folder/subfolder/file.txt",
		},
		{
			name:     "path with leading dot slash",
			input:    "./folder/file.txt",
			expected: "folder/file.txt",
		},
		{
			name:     "path with leading slash",
			input:    "/folder/file.txt",
			expected: "folder/file.txt",
		},
		{
			name:     "path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "../../../etc/passwd",
		},
		{
			name:     "complex path with multiple issues",
			input:    "./\\folder\\..\\file.txt",
			expected: "file.txt",
		},
		{
			name:     "empty string",
			input:    "",
			expected: ".",
		},
		{
			name:     "dot only",
			input:    ".",
			expected: ".",
		},
		{
			name:     "double dot",
			input:    "..",
			expected: "..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := normalizeStorageKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRawSave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		basePath    string
		errContains string
		params      SaveParams
		wantErr     bool
	}{
		{
			name:     "successful save",
			basePath: "",
			params: SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			wantErr: false,
		},
		{
			name:     "save with nested directory",
			basePath: "",
			params: SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "folder/subfolder/test-file.txt",
				Data:       []byte("nested data"),
			},
			wantErr: false,
		},
		{
			name:     "path traversal attack",
			basePath: "",
			params: SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "../../../etc/passwd",
				Data:       []byte("malicious data"),
			},
			wantErr:     true,
			errContains: "path traversal detected",
		},
		{
			name:     "empty storage key",
			basePath: "",
			params: SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "",
				Data:       []byte("test data"),
			},
			wantErr:     true, // Empty storage key normalizes to "." which is a directory
			errContains: "is a directory",
		},
		{
			name:     "large data",
			basePath: "",
			params: SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "large-file.dat",
				Data:       make([]byte, 1024*1024), // 1MB
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			if tt.basePath != "" {
				basePath = tt.basePath
			}

			saveFunc := rawSave(basePath)
			err := saveFunc(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)

				// Verify file was created and contains correct data
				userDir := filepath.Join(basePath, tt.params.UserID.String())
				normalizedKey := normalizeStorageKey(tt.params.StorageKey)
				fullPath := filepath.Join(userDir, normalizedKey)

				data, err := os.ReadFile(fullPath)
				require.NoError(t, err)
				assert.Equal(t, tt.params.Data, data)

				// Verify directory permissions
				info, err := os.Stat(userDir)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(DirectoryPermission), info.Mode().Perm())

				// Verify file permissions
				info, err = os.Stat(fullPath)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(FilePermission), info.Mode().Perm())
			}
		})
	}
}

func TestRawLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		errContains string
		setupData   []byte
		want        []byte
		params      LoadParams
		setupFile   bool
		wantErr     bool
	}{
		{
			name:      "successful load",
			setupFile: true,
			setupData: []byte("test data"),
			params: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
			},
			want:    []byte("test data"),
			wantErr: false,
		},
		{
			name:      "file not found",
			setupFile: false,
			params: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "nonexistent.txt",
			},
			want:        nil,
			wantErr:     true,
			errContains: "file not found",
		},
		{
			name:      "path traversal attack",
			setupFile: false,
			params: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "../../../etc/passwd",
			},
			want:        nil,
			wantErr:     true,
			errContains: "path traversal detected",
		},
		{
			name:      "empty file",
			setupFile: true,
			setupData: []byte{},
			params: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "empty-file.txt",
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name:      "large file",
			setupFile: true,
			setupData: make([]byte, 1024*1024), // 1MB
			params: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "large-file.dat",
			},
			want:    make([]byte, 1024*1024),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			loadFunc := rawLoad(basePath)

			// Setup test file if needed
			if tt.setupFile {
				userDir := filepath.Join(basePath, tt.params.UserID.String())
				err := os.MkdirAll(userDir, DirectoryPermission)
				require.NoError(t, err)

				normalizedKey := normalizeStorageKey(tt.params.StorageKey)
				fullPath := filepath.Join(userDir, normalizedKey)

				// Create directory for file if needed
				dir := filepath.Dir(fullPath)
				err = os.MkdirAll(dir, DirectoryPermission)
				require.NoError(t, err)

				err = os.WriteFile(fullPath, tt.setupData, FilePermission)
				require.NoError(t, err)
			}

			data, err := loadFunc(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, data)
			}
		})
	}
}

func TestRawDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		errContains string
		setupData   []byte
		params      DeleteParams
		setupFile   bool
		wantErr     bool
	}{
		{
			name:      "successful delete",
			setupFile: true,
			setupData: []byte("test data"),
			params: DeleteParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
			},
			wantErr: false,
		},
		{
			name:      "delete nonexistent file",
			setupFile: false,
			params: DeleteParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "nonexistent.txt",
			},
			wantErr: false, // Should not error when deleting nonexistent file
		},
		{
			name:      "path traversal attack",
			setupFile: false,
			params: DeleteParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "../../../etc/passwd",
			},
			wantErr:     true,
			errContains: "path traversal detected",
		},
		{
			name:      "delete file in nested directory",
			setupFile: true,
			setupData: []byte("nested data"),
			params: DeleteParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "folder/subfolder/nested-file.txt",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			deleteFunc := rawDelete(basePath)

			// Setup test file if needed
			if tt.setupFile {
				userDir := filepath.Join(basePath, tt.params.UserID.String())
				err := os.MkdirAll(userDir, DirectoryPermission)
				require.NoError(t, err)

				normalizedKey := normalizeStorageKey(tt.params.StorageKey)
				fullPath := filepath.Join(userDir, normalizedKey)

				// Create directory for file if needed
				dir := filepath.Dir(fullPath)
				err = os.MkdirAll(dir, DirectoryPermission)
				require.NoError(t, err)

				err = os.WriteFile(fullPath, tt.setupData, FilePermission)
				require.NoError(t, err)
			}

			err := deleteFunc(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)

				// Verify file was deleted
				if tt.setupFile {
					userDir := filepath.Join(basePath, tt.params.UserID.String())
					normalizedKey := normalizeStorageKey(tt.params.StorageKey)
					fullPath := filepath.Join(userDir, normalizedKey)

					_, err := os.Stat(fullPath)
					assert.True(t, os.IsNotExist(err), "file should be deleted")
				}
			}
		})
	}
}
