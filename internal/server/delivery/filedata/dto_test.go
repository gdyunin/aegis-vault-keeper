package filedata

import (
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileDataFromApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()
	updatedAt := time.Now()
	data := []byte("test file content")

	appFileData := &filedata.FileData{
		ID:          fileID,
		UserID:      userID,
		StorageKey:  "test.txt",
		HashSum:     "abcd1234",
		Description: "Test file",
		UpdatedAt:   updatedAt,
		Data:        data,
	}

	result := NewFileDataFromApp(appFileData)

	require.NotNil(t, result)
	assert.Equal(t, fileID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "test.txt", result.StorageKey)
	assert.Equal(t, "abcd1234", result.HashSum)
	assert.Equal(t, "Test file", result.Description)
	assert.Equal(t, updatedAt, result.UpdatedAt)
	assert.Equal(t, data, result.Data)
}

func TestNewFileDataListFromApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []*filedata.FileData
		expected int
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: 0,
		},
		{
			name:     "empty input",
			input:    []*filedata.FileData{},
			expected: 0,
		},
		{
			name: "single file",
			input: []*filedata.FileData{
				{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					StorageKey: "file1.txt",
					Data:       []byte("content1"),
				},
			},
			expected: 1,
		},
		{
			name: "multiple files",
			input: []*filedata.FileData{
				{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					StorageKey: "file1.txt",
					Data:       []byte("content1"),
				},
				{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					StorageKey: "file2.txt",
					Data:       []byte("content2"),
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewFileDataListFromApp(tt.input)
			assert.Len(t, result, tt.expected)

			for i, fd := range tt.input {
				assert.Equal(t, fd.ID, result[i].ID)
				assert.Equal(t, fd.StorageKey, result[i].StorageKey)
				assert.Equal(t, fd.Data, result[i].Data)
			}
		})
	}
}

func TestFileData_ToApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()
	updatedAt := time.Now()

	tests := []struct {
		fileData *FileData
		expected *filedata.FileData
		name     string
		userID   uuid.UUID
	}{
		{
			name:     "nil filedata",
			fileData: nil,
			userID:   userID,
			expected: nil,
		},
		{
			name: "valid filedata",
			fileData: &FileData{
				ID:          fileID,
				UserID:      userID,
				StorageKey:  "test.pdf",
				HashSum:     "hash123",
				Description: "Test PDF",
				UpdatedAt:   updatedAt,
				Data:        []byte("pdf content"),
			},
			userID: userID,
			expected: &filedata.FileData{
				ID:          fileID,
				UserID:      userID,
				StorageKey:  "test.pdf",
				HashSum:     "hash123",
				Description: "Test PDF",
				UpdatedAt:   updatedAt,
				Data:        []byte("pdf content"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.fileData.ToApp(tt.userID)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.UserID, result.UserID)
			assert.Equal(t, tt.expected.StorageKey, result.StorageKey)
			assert.Equal(t, tt.expected.HashSum, result.HashSum)
			assert.Equal(t, tt.expected.Description, result.Description)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
			assert.Equal(t, tt.expected.Data, result.Data)
		})
	}
}

func TestFilesToApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name     string
		files    []*FileData
		userID   uuid.UUID
		expected int
	}{
		{
			name:     "nil files",
			files:    nil,
			userID:   userID,
			expected: 0,
		},
		{
			name:     "empty files",
			files:    []*FileData{},
			userID:   userID,
			expected: 0,
		},
		{
			name: "single file",
			files: []*FileData{
				{
					ID:         uuid.New(),
					StorageKey: "file1.txt",
					Data:       []byte("content"),
				},
			},
			userID:   userID,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FilesToApp(tt.files, tt.userID)

			if tt.files == nil {
				assert.Nil(t, result)
				return
			}

			assert.Len(t, result, tt.expected)

			for i, file := range tt.files {
				assert.Equal(t, file.ID, result[i].ID)
				assert.Equal(t, tt.userID, result[i].UserID)
				assert.Equal(t, file.StorageKey, result[i].StorageKey)
			}
		})
	}
}

func TestFileData_withoutData(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()
	updatedAt := time.Now()
	originalData := []byte("sensitive file content")

	fileData := &FileData{
		ID:          fileID,
		UserID:      userID,
		StorageKey:  "document.pdf",
		HashSum:     "hash456",
		Description: "Important document",
		UpdatedAt:   updatedAt,
		Data:        originalData,
	}

	result := fileData.withoutData()

	require.NotNil(t, result)
	assert.Equal(t, fileID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "document.pdf", result.StorageKey)
	assert.Equal(t, "hash456", result.HashSum)
	assert.Equal(t, "Important document", result.Description)
	assert.Equal(t, updatedAt, result.UpdatedAt)
	assert.Nil(t, result.Data, "Data field should be nil in withoutData result")

	// Verify original is unchanged
	assert.Equal(t, originalData, fileData.Data, "Original fileData should be unchanged")
}
