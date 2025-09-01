package filestorage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
)

// mockKeyProvider is a simple mock implementation for testing.
type mockKeyProvider struct {
	err error
	key []byte
}

func (m *mockKeyProvider) UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.key, nil
}

func TestNewRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keyProvider keyprv.UserKeyProvider
		name        string
		basePath    string
	}{
		{
			name:        "valid parameters",
			basePath:    "/tmp/test",
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
		},
		{
			name:        "empty base path",
			basePath:    "",
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository(tt.basePath, tt.keyProvider)

			assert.NotNil(t, repo)
			assert.NotNil(t, repo.save)
			assert.NotNil(t, repo.load)
			assert.NotNil(t, repo.delete)
		})
	}
}

func TestRepository_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keyProvider keyprv.UserKeyProvider
		name        string
		errContains string
		params      SaveParams
		wantErr     bool
	}{
		{
			name: "successful save",
			params: SaveParams{
				UserID:     uuid.New(),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			wantErr:     false,
		},
		{
			name: "key provider error",
			params: SaveParams{
				UserID:     uuid.New(),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			keyProvider: &mockKeyProvider{err: assert.AnError},
			wantErr:     true,
			errContains: "failed to save file to storage",
		},
		{
			name: "invalid storage key with path traversal",
			params: SaveParams{
				UserID:     uuid.New(),
				StorageKey: "../../../etc/passwd",
				Data:       []byte("malicious data"),
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			wantErr:     true,
			errContains: "failed to save file to storage",
		},
		{
			name: "empty storage key",
			params: SaveParams{
				UserID:     uuid.New(),
				StorageKey: "",
				Data:       []byte("test data"),
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			wantErr:     true, // Empty storage key normalizes to "." which is a directory
			errContains: "failed to save file to storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use temporary directory for testing
			basePath := t.TempDir()
			repo := NewRepository(basePath, tt.keyProvider)

			err := repo.Save(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepository_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keyProvider keyprv.UserKeyProvider
		setupData   *SaveParams
		name        string
		errContains string
		want        []byte
		loadParams  LoadParams
		wantErr     bool
	}{
		{
			name: "successful load",
			setupData: &SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			loadParams: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			want:        []byte("test data"),
			wantErr:     false,
		},
		{
			name:      "file not found",
			setupData: nil,
			loadParams: LoadParams{
				UserID:     uuid.New(),
				StorageKey: "nonexistent.txt",
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			want:        nil,
			wantErr:     true,
			errContains: "failed to load file from storage",
		},
		{
			name: "key provider error",
			setupData: &SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			loadParams: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
			},
			keyProvider: &mockKeyProvider{err: assert.AnError},
			want:        nil,
			wantErr:     true,
			errContains: "failed to load file from storage",
		},
		{
			name: "invalid storage key with path traversal",
			setupData: &SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			loadParams: LoadParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "../../../etc/passwd",
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			want:        nil,
			wantErr:     true,
			errContains: "failed to load file from storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			repo := NewRepository(basePath, tt.keyProvider)

			// Setup test data if provided
			if tt.setupData != nil {
				setupKeyProvider := &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")}
				setupRepo := NewRepository(basePath, setupKeyProvider)
				err := setupRepo.Save(context.Background(), *tt.setupData)
				require.NoError(t, err)
			}

			data, err := repo.Load(context.Background(), tt.loadParams)

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

func TestRepository_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keyProvider  keyprv.UserKeyProvider
		setupData    *SaveParams
		name         string
		errContains  string
		deleteParams DeleteParams
		wantErr      bool
	}{
		{
			name: "successful delete",
			setupData: &SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			deleteParams: DeleteParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			wantErr:     false,
		},
		{
			name:      "delete nonexistent file",
			setupData: nil,
			deleteParams: DeleteParams{
				UserID:     uuid.New(),
				StorageKey: "nonexistent.txt",
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			wantErr:     false, // Delete of nonexistent file should not error
		},
		{
			name: "invalid storage key with path traversal",
			setupData: &SaveParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "test-file.txt",
				Data:       []byte("test data"),
			},
			deleteParams: DeleteParams{
				UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StorageKey: "../../../etc/passwd",
			},
			keyProvider: &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")},
			wantErr:     true,
			errContains: "failed to delete file from storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			repo := NewRepository(basePath, tt.keyProvider)

			// Setup test data if provided
			if tt.setupData != nil {
				setupKeyProvider := &mockKeyProvider{key: []byte("test-key-32-bytes-for-encryption")}
				setupRepo := NewRepository(basePath, setupKeyProvider)
				err := setupRepo.Save(context.Background(), *tt.setupData)
				require.NoError(t, err)
			}

			err := repo.Delete(context.Background(), tt.deleteParams)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
