package filedata

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/filedata"
	filestorage "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/filestorage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository implements Repository interface for testing.
type MockRepository struct {
	SaveFunc func(ctx context.Context, params repository.SaveParams) error
	LoadFunc func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error)
}

func (m *MockRepository) Save(ctx context.Context, params repository.SaveParams) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, params)
	}
	return nil
}

func (m *MockRepository) Load(
	ctx context.Context,
	params repository.LoadParams,
) ([]*filedata.FileData, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(ctx, params)
	}
	return nil, nil
}

// MockFileStorageRepository implements FileStorageRepository interface for testing.
type MockFileStorageRepository struct {
	SaveFunc   func(ctx context.Context, params filestorage.SaveParams) error
	LoadFunc   func(ctx context.Context, params filestorage.LoadParams) ([]byte, error)
	DeleteFunc func(ctx context.Context, params filestorage.DeleteParams) error
}

func (m *MockFileStorageRepository) Save(ctx context.Context, params filestorage.SaveParams) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, params)
	}
	return nil
}

func (m *MockFileStorageRepository) Load(ctx context.Context, params filestorage.LoadParams) ([]byte, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(ctx, params)
	}
	return nil, nil
}

func (m *MockFileStorageRepository) Delete(ctx context.Context, params filestorage.DeleteParams) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, params)
	}
	return nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		repo Repository
		fs   FileStorageRepository
		want *Service
		name string
	}{
		{
			name: "success/creates_service_with_repositories",
			repo: &MockRepository{},
			fs:   &MockFileStorageRepository{},
			want: &Service{
				r:  &MockRepository{},
				fs: &MockFileStorageRepository{},
			},
		},
		{
			name: "success/nil_repositories",
			repo: nil,
			fs:   nil,
			want: &Service{r: nil, fs: nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewService(tt.repo, tt.fs)
			require.NotNil(t, got)
			assert.Equal(t, tt.repo, got.r)
			assert.Equal(t, tt.fs, got.fs)
		})
	}
}

func TestService_Pull(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testFileID := uuid.New()
	testTime := time.Now()
	testData := []byte("test file content")

	tests := []struct {
		setupRepoMock func(*MockRepository)
		setupFSMock   func(*MockFileStorageRepository)
		want          *FileData
		name          string
		wantErrText   string
		params        PullParams
		wantErr       bool
	}{
		{
			name: "success/file_found_and_loaded",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					assert.Equal(t, testFileID, params.ID)
					assert.Equal(t, testUserID, params.UserID)
					return []*filedata.FileData{{
						ID:          testFileID,
						UserID:      testUserID,
						StorageKey:  []byte("test/storage/key"),
						HashSum:     []byte("abc123"),
						Description: []byte("test description"),
						UpdatedAt:   testTime,
					}}, nil
				}
			},
			setupFSMock: func(m *MockFileStorageRepository) {
				m.LoadFunc = func(ctx context.Context, params filestorage.LoadParams) ([]byte, error) {
					assert.Equal(t, testUserID, params.UserID)
					assert.Equal(t, "test/storage/key", params.StorageKey)
					return testData, nil
				}
			},
			want: &FileData{
				ID:          testFileID,
				UserID:      testUserID,
				StorageKey:  "test/storage/key",
				HashSum:     "abc123",
				Description: "test description",
				UpdatedAt:   testTime,
				Data:        testData,
			},
			wantErr: false,
		},
		{
			name: "error/file_metadata_not_found",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{}, nil // Empty result
				}
			},
			setupFSMock: func(m *MockFileStorageRepository) {},
			want:        nil,
			wantErr:     true,
			wantErrText: "file not found",
		},
		{
			name: "error/repository_metadata_error",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return nil, errors.New("database error")
				}
			},
			setupFSMock: func(m *MockFileStorageRepository) {},
			want:        nil,
			wantErr:     true,
			wantErrText: "failed to load metadata",
		},
		{
			name: "error/file_storage_load_error",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{{
						ID:          testFileID,
						UserID:      testUserID,
						StorageKey:  []byte("test/storage/key"),
						HashSum:     []byte("abc123"),
						Description: []byte("test description"),
						UpdatedAt:   testTime,
					}}, nil
				}
			},
			setupFSMock: func(m *MockFileStorageRepository) {
				m.LoadFunc = func(ctx context.Context, params filestorage.LoadParams) ([]byte, error) {
					return nil, errors.New("storage error")
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "failed to load file data",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			mockFS := &MockFileStorageRepository{}

			if tt.setupRepoMock != nil {
				tt.setupRepoMock(mockRepo)
			}
			if tt.setupFSMock != nil {
				tt.setupFSMock(mockFS)
			}

			service := NewService(mockRepo, mockFS)
			got, err := service.Pull(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		setupRepoMock func(*MockRepository)
		name          string
		wantErrText   string
		want          []*FileData
		params        ListParams
		wantErr       bool
	}{
		{
			name: "success/files_found",
			params: ListParams{
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					assert.Equal(t, testUserID, params.UserID)
					return []*filedata.FileData{
						{
							ID:          uuid.New(),
							UserID:      testUserID,
							StorageKey:  []byte("file1.txt"),
							HashSum:     []byte("hash1"),
							Description: []byte("desc1"),
							UpdatedAt:   testTime,
						},
						{
							ID:          uuid.New(),
							UserID:      testUserID,
							StorageKey:  []byte("file2.txt"),
							HashSum:     []byte("hash2"),
							Description: []byte("desc2"),
							UpdatedAt:   testTime,
						},
					}, nil
				}
			},
			want: []*FileData{
				{
					ID:          uuid.UUID{},
					UserID:      testUserID,
					StorageKey:  "file1.txt",
					HashSum:     "hash1",
					Description: "desc1",
					UpdatedAt:   testTime,
				},
				{
					ID:          uuid.UUID{},
					UserID:      testUserID,
					StorageKey:  "file2.txt",
					HashSum:     "hash2",
					Description: "desc2",
					UpdatedAt:   testTime,
				},
			},
			wantErr: false,
		},
		{
			name: "success/empty_list",
			params: ListParams{
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{}, nil
				}
			},
			want:    []*FileData{},
			wantErr: false,
		},
		{
			name: "error/repository_error",
			params: ListParams{
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return nil, errors.New("database error")
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "failed to load files",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			mockFS := &MockFileStorageRepository{}

			if tt.setupRepoMock != nil {
				tt.setupRepoMock(mockRepo)
			}

			service := NewService(mockRepo, mockFS)
			got, err := service.List(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				// Compare length first
				require.Len(t, got, len(tt.want))
				// Compare individual elements (excluding generated IDs)
				for i := range got {
					assert.Equal(t, tt.want[i].UserID, got[i].UserID)
					assert.Equal(t, tt.want[i].StorageKey, got[i].StorageKey)
					assert.Equal(t, tt.want[i].HashSum, got[i].HashSum)
					assert.Equal(t, tt.want[i].Description, got[i].Description)
					assert.Equal(t, tt.want[i].UpdatedAt, got[i].UpdatedAt)
				}
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testData := []byte("test file content")

	tests := []struct {
		params        *PushParams
		setupRepoMock func(*MockRepository)
		setupFSMock   func(*MockFileStorageRepository)
		name          string
		wantErrText   string
		wantID        uuid.UUID
		wantErr       bool
	}{
		{
			name: "success/create_new_file",
			params: &PushParams{
				UserID:      testUserID,
				StorageKey:  "test/file.txt",
				Description: "test description",
				Data:        testData,
			},
			setupRepoMock: func(m *MockRepository) {
				m.SaveFunc = func(ctx context.Context, params repository.SaveParams) error {
					assert.NotNil(t, params.Entity)
					assert.Equal(t, testUserID, params.Entity.UserID)
					assert.Equal(t, []byte("test/file.txt"), params.Entity.StorageKey)
					assert.Equal(t, []byte("test description"), params.Entity.Description)
					return nil
				}
			},
			setupFSMock: func(m *MockFileStorageRepository) {
				m.SaveFunc = func(ctx context.Context, params filestorage.SaveParams) error {
					assert.Equal(t, testUserID, params.UserID)
					assert.Equal(t, "test/file.txt", params.StorageKey)
					assert.Equal(t, testData, params.Data)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "error/empty_data",
			params: &PushParams{
				UserID:      testUserID,
				StorageKey:  "test/file.txt",
				Description: "test description",
				Data:        []byte{}, // Empty data
			},
			setupRepoMock: func(m *MockRepository) {},
			setupFSMock:   func(m *MockFileStorageRepository) {},
			wantErr:       true,
			wantErrText:   "file data is required",
		},
		{
			name: "error/invalid_domain_file",
			params: &PushParams{
				UserID:      testUserID,
				StorageKey:  "", // Invalid empty storage key
				Description: "test description",
				Data:        testData,
			},
			setupRepoMock: func(m *MockRepository) {},
			setupFSMock:   func(m *MockFileStorageRepository) {},
			wantErr:       true,
			wantErrText:   "failed to create file",
		},
		{
			name: "error/file_storage_save_error",
			params: &PushParams{
				UserID:      testUserID,
				StorageKey:  "test/file.txt",
				Description: "test description",
				Data:        testData,
			},
			setupRepoMock: func(m *MockRepository) {},
			setupFSMock: func(m *MockFileStorageRepository) {
				m.SaveFunc = func(ctx context.Context, params filestorage.SaveParams) error {
					return errors.New("storage error")
				}
			},
			wantErr:     true,
			wantErrText: "failed to save file data",
		},
		{
			name: "error/metadata_save_error_with_rollback",
			params: &PushParams{
				UserID:      testUserID,
				StorageKey:  "test/file.txt",
				Description: "test description",
				Data:        testData,
			},
			setupRepoMock: func(m *MockRepository) {
				m.SaveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return errors.New("metadata save error")
				}
			},
			setupFSMock: func(m *MockFileStorageRepository) {
				m.SaveFunc = func(ctx context.Context, params filestorage.SaveParams) error {
					return nil // FS save succeeds
				}
				m.DeleteFunc = func(ctx context.Context, params filestorage.DeleteParams) error {
					return nil // Rollback succeeds
				}
			},
			wantErr:     true,
			wantErrText: "failed to save file metadata",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			mockFS := &MockFileStorageRepository{}

			if tt.setupRepoMock != nil {
				tt.setupRepoMock(mockRepo)
			}
			if tt.setupFSMock != nil {
				tt.setupFSMock(mockFS)
			}

			service := NewService(mockRepo, mockFS)
			gotID, err := service.Push(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Equal(t, uuid.Nil, gotID)
			} else {
				require.NoError(t, err)
				if tt.wantID != uuid.Nil {
					assert.Equal(t, tt.wantID, gotID)
				} else {
					assert.NotEqual(t, uuid.Nil, gotID)
				}
			}
		})
	}
}

func TestService_loadMetadata(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testFileID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		setupRepoMock func(*MockRepository)
		want          *filedata.FileData
		name          string
		wantErrText   string
		params        PullParams
		wantErr       bool
	}{
		{
			name: "success/metadata_found",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					assert.Equal(t, testFileID, params.ID)
					assert.Equal(t, testUserID, params.UserID)
					return []*filedata.FileData{{
						ID:          testFileID,
						UserID:      testUserID,
						StorageKey:  []byte("test/storage/key"),
						HashSum:     []byte("abc123"),
						Description: []byte("test description"),
						UpdatedAt:   testTime,
					}}, nil
				}
			},
			want: &filedata.FileData{
				ID:          testFileID,
				UserID:      testUserID,
				StorageKey:  []byte("test/storage/key"),
				HashSum:     []byte("abc123"),
				Description: []byte("test description"),
				UpdatedAt:   testTime,
			},
			wantErr: false,
		},
		{
			name: "error/file_not_found",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{}, nil
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "file not found",
		},
		{
			name: "error/repository_error",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			setupRepoMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return nil, errors.New("database error")
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "failed to load file metadata",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			mockFS := &MockFileStorageRepository{}

			if tt.setupRepoMock != nil {
				tt.setupRepoMock(mockRepo)
			}

			service := NewService(mockRepo, mockFS)
			got, err := service.loadMetadata(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_findFileForUpdate(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		params      *PushParams
		mockRepo    *MockRepository
		want        *filedata.FileData
		name        string
		wantErrText string
		wantErr     bool
	}{
		{
			name: "successful file access",
			params: &PushParams{
				ID:     fileID,
				UserID: userID,
			},
			mockRepo: &MockRepository{
				LoadFunc: func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{
						{
							ID:          fileID,
							UserID:      userID,
							Description: []byte("test file"),
						},
					}, nil
				},
			},
			want: &filedata.FileData{
				ID:          fileID,
				UserID:      userID,
				Description: []byte("test file"),
			},
		},
		{
			name: "access denied - different user",
			params: &PushParams{
				ID:     fileID,
				UserID: otherUserID,
			},
			mockRepo: &MockRepository{
				LoadFunc: func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{
						{
							ID:          fileID,
							UserID:      userID, // Different user owns the file
							Description: []byte("test file"),
						},
					}, nil
				},
			},
			wantErr:     true,
			wantErrText: "access denied to file",
		},
		{
			name: "file not found",
			params: &PushParams{
				ID:     fileID,
				UserID: userID,
			},
			mockRepo: &MockRepository{
				LoadFunc: func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return nil, ErrFileNotFound
				},
			},
			wantErr:     true,
			wantErrText: "access check for updating file failed",
		},
		{
			name: "repository error",
			params: &PushParams{
				ID:     fileID,
				UserID: userID,
			},
			mockRepo: &MockRepository{
				LoadFunc: func(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error) {
					return nil, errors.New("database error")
				},
			},
			wantErr:     true,
			wantErrText: "access check for updating file failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(tt.mockRepo, &MockFileStorageRepository{})
			got, err := service.findFileForUpdate(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_removeOldFileOnKeyChange(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()

	tests := []struct {
		existing      *filedata.FileData
		mockFS        *MockFileStorageRepository
		name          string
		newStorageKey string
		wantErrText   string
		wantErr       bool
	}{
		{
			name: "no key change - no deletion",
			existing: &filedata.FileData{
				ID:         fileID,
				UserID:     userID,
				StorageKey: []byte("same_key"),
			},
			newStorageKey: "same_key",
			mockFS:        &MockFileStorageRepository{},
		},
		{
			name: "key changed - successful deletion",
			existing: &filedata.FileData{
				ID:         fileID,
				UserID:     userID,
				StorageKey: []byte("old_key"),
			},
			newStorageKey: "new_key",
			mockFS: &MockFileStorageRepository{
				DeleteFunc: func(ctx context.Context, params filestorage.DeleteParams) error {
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, "old_key", params.StorageKey)
					return nil
				},
			},
		},
		{
			name: "key changed - deletion fails",
			existing: &filedata.FileData{
				ID:         fileID,
				UserID:     userID,
				StorageKey: []byte("old_key"),
			},
			newStorageKey: "new_key",
			mockFS: &MockFileStorageRepository{
				DeleteFunc: func(ctx context.Context, params filestorage.DeleteParams) error {
					return errors.New("storage deletion failed")
				},
			},
			wantErr:     true,
			wantErrText: "failed to delete old file data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(&MockRepository{}, tt.mockFS)
			err := service.removeOldFileOnKeyChange(context.Background(), tt.existing, tt.newStorageKey)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_rollbackFileSave(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()

	tests := []struct {
		fileData    *filedata.FileData
		mockFS      *MockFileStorageRepository
		name        string
		wantErrText string
		wantErr     bool
	}{
		{
			name: "successful rollback",
			fileData: &filedata.FileData{
				ID:         fileID,
				UserID:     userID,
				StorageKey: []byte("test_key"),
			},
			mockFS: &MockFileStorageRepository{
				DeleteFunc: func(ctx context.Context, params filestorage.DeleteParams) error {
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, "test_key", params.StorageKey)
					return nil
				},
			},
		},
		{
			name: "rollback deletion fails",
			fileData: &filedata.FileData{
				ID:         fileID,
				UserID:     userID,
				StorageKey: []byte("test_key"),
			},
			mockFS: &MockFileStorageRepository{
				DeleteFunc: func(ctx context.Context, params filestorage.DeleteParams) error {
					return errors.New("deletion failed")
				},
			},
			wantErr:     true,
			wantErrText: "rollback of file save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(&MockRepository{}, tt.mockFS)
			err := service.rollbackFileSave(context.Background(), tt.fileData)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
