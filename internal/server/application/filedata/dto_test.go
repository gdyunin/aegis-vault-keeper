package filedata

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileFromDomain(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testUserID := uuid.New()
	testFileID := uuid.New()

	tests := []struct {
		input *filedata.FileData
		want  *FileData
		name  string
	}{
		{
			name: "success/valid_domain_file",
			input: &filedata.FileData{
				ID:          testFileID,
				UserID:      testUserID,
				StorageKey:  []byte("test/file/path"),
				HashSum:     []byte("abc123def456"),
				Description: []byte("test file description"),
				UpdatedAt:   testTime,
			},
			want: &FileData{
				ID:          testFileID,
				UserID:      testUserID,
				StorageKey:  "test/file/path",
				HashSum:     "abc123def456",
				Description: "test file description",
				UpdatedAt:   testTime,
			},
		},
		{
			name: "success/empty_fields",
			input: &filedata.FileData{
				ID:          testFileID,
				UserID:      testUserID,
				StorageKey:  []byte(""),
				HashSum:     []byte(""),
				Description: []byte(""),
				UpdatedAt:   testTime,
			},
			want: &FileData{
				ID:          testFileID,
				UserID:      testUserID,
				StorageKey:  "",
				HashSum:     "",
				Description: "",
				UpdatedAt:   testTime,
			},
		},
		{
			name:  "nil_input/returns_nil",
			input: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newFileFromDomain(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewFilesFromDomain(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testUserID := uuid.New()
	testFileID1 := uuid.New()
	testFileID2 := uuid.New()

	tests := []struct {
		name  string
		input []*filedata.FileData
		want  []*FileData
	}{
		{
			name: "success/multiple_files",
			input: []*filedata.FileData{
				{
					ID:          testFileID1,
					UserID:      testUserID,
					StorageKey:  []byte("file1.txt"),
					HashSum:     []byte("hash1"),
					Description: []byte("desc1"),
					UpdatedAt:   testTime,
				},
				{
					ID:          testFileID2,
					UserID:      testUserID,
					StorageKey:  []byte("file2.txt"),
					HashSum:     []byte("hash2"),
					Description: []byte("desc2"),
					UpdatedAt:   testTime,
				},
			},
			want: []*FileData{
				{
					ID:          testFileID1,
					UserID:      testUserID,
					StorageKey:  "file1.txt",
					HashSum:     "hash1",
					Description: "desc1",
					UpdatedAt:   testTime,
				},
				{
					ID:          testFileID2,
					UserID:      testUserID,
					StorageKey:  "file2.txt",
					HashSum:     "hash2",
					Description: "desc2",
					UpdatedAt:   testTime,
				},
			},
		},
		{
			name:  "success/empty_slice",
			input: []*filedata.FileData{},
			want:  []*FileData{},
		},
		{
			name:  "success/nil_slice",
			input: nil,
			want:  []*FileData{},
		},
		{
			name: "success/slice_with_nil_element",
			input: []*filedata.FileData{
				{
					ID:          testFileID1,
					UserID:      testUserID,
					StorageKey:  []byte("file1.txt"),
					HashSum:     []byte("hash1"),
					Description: []byte("desc1"),
					UpdatedAt:   testTime,
				},
				nil,
			},
			want: []*FileData{
				{
					ID:          testFileID1,
					UserID:      testUserID,
					StorageKey:  "file1.txt",
					HashSum:     "hash1",
					Description: "desc1",
					UpdatedAt:   testTime,
				},
				nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newFilesFromDomain(tt.input)
			require.Len(t, got, len(tt.want))

			for i := range got {
				assert.Equal(t, tt.want[i], got[i])
			}
		})
	}
}

func TestPullParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testFileID := uuid.New()

	tests := []struct {
		name       string
		params     PullParams
		wantID     uuid.UUID
		wantUserID uuid.UUID
	}{
		{
			name: "success/valid_params",
			params: PullParams{
				ID:     testFileID,
				UserID: testUserID,
			},
			wantID:     testFileID,
			wantUserID: testUserID,
		},
		{
			name: "success/nil_uuids",
			params: PullParams{
				ID:     uuid.Nil,
				UserID: uuid.Nil,
			},
			wantID:     uuid.Nil,
			wantUserID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantID, tt.params.ID)
			assert.Equal(t, tt.wantUserID, tt.params.UserID)
		})
	}
}

func TestListParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		name       string
		params     ListParams
		wantUserID uuid.UUID
	}{
		{
			name: "success/valid_params",
			params: ListParams{
				UserID: testUserID,
			},
			wantUserID: testUserID,
		},
		{
			name: "success/nil_uuid",
			params: ListParams{
				UserID: uuid.Nil,
			},
			wantUserID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantUserID, tt.params.UserID)
		})
	}
}

func TestPushParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testFileID := uuid.New()
	testData := []byte("test file content")

	tests := []struct {
		name            string
		wantStorageKey  string
		wantDescription string
		wantData        []byte
		params          PushParams
		wantID          uuid.UUID
		wantUserID      uuid.UUID
	}{
		{
			name: "success/valid_params",
			params: PushParams{
				StorageKey:  "test/path/file.txt",
				Description: "test description",
				Data:        testData,
				ID:          testFileID,
				UserID:      testUserID,
			},
			wantStorageKey:  "test/path/file.txt",
			wantDescription: "test description",
			wantData:        testData,
			wantID:          testFileID,
			wantUserID:      testUserID,
		},
		{
			name: "success/empty_optional_fields",
			params: PushParams{
				StorageKey:  "",
				Description: "",
				Data:        []byte{},
				ID:          uuid.Nil,
				UserID:      uuid.Nil,
			},
			wantStorageKey:  "",
			wantDescription: "",
			wantData:        []byte{},
			wantID:          uuid.Nil,
			wantUserID:      uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantStorageKey, tt.params.StorageKey)
			assert.Equal(t, tt.wantDescription, tt.params.Description)
			assert.Equal(t, tt.wantData, tt.params.Data)
			assert.Equal(t, tt.wantID, tt.params.ID)
			assert.Equal(t, tt.wantUserID, tt.params.UserID)
		})
	}
}

func TestPushParams_calculateDataHashSum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		wantHash string
		data     []byte
	}{
		{
			name: "success/test_data",
			data: []byte("test file content"),
			wantHash: func() string {
				hash := sha256.Sum256([]byte("test file content"))
				return hex.EncodeToString(hash[:])
			}(),
		},
		{
			name: "success/empty_data",
			data: []byte{},
			wantHash: func() string {
				hash := sha256.Sum256([]byte{})
				return hex.EncodeToString(hash[:])
			}(),
		},
		{
			name: "success/binary_data",
			data: []byte{0x00, 0x01, 0xFF, 0xAB, 0xCD},
			wantHash: func() string {
				hash := sha256.Sum256([]byte{0x00, 0x01, 0xFF, 0xAB, 0xCD})
				return hex.EncodeToString(hash[:])
			}(),
		},
		{
			name: "success/large_data",
			data: make([]byte, 1024),
			wantHash: func() string {
				data := make([]byte, 1024)
				hash := sha256.Sum256(data)
				return hex.EncodeToString(hash[:])
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := &PushParams{Data: tt.data}
			got := params.calculateDataHashSum()

			assert.Equal(t, tt.wantHash, got)
			assert.Len(t, got, 64) // SHA256 hash is 64 hex characters
		})
	}
}

func TestPushParams_calculateDataHashSum_Consistency(t *testing.T) {
	t.Parallel()

	// Test that same data produces same hash multiple times
	testData := []byte("consistent test data")
	params := &PushParams{Data: testData}

	hash1 := params.calculateDataHashSum()
	hash2 := params.calculateDataHashSum()

	assert.Equal(t, hash1, hash2, "hash should be consistent for same data")
	assert.NotEmpty(t, hash1)
}

func TestPushParams_calculateDataHashSum_Different(t *testing.T) {
	t.Parallel()

	// Test that different data produces different hashes
	params1 := &PushParams{Data: []byte("data 1")}
	params2 := &PushParams{Data: []byte("data 2")}

	hash1 := params1.calculateDataHashSum()
	hash2 := params2.calculateDataHashSum()

	assert.NotEqual(t, hash1, hash2, "different data should produce different hashes")
}
