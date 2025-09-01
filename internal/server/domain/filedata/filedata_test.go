package filedata

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	validHashSum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // SHA256 of empty string

	type args struct {
		params NewFileDataParams
	}
	tests := []struct {
		want    func(t *testing.T, fd *FileData)
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid/complete_file",
			args: args{
				params: NewFileDataParams{
					Description: "Important document",
					StorageKey:  "documents/file.pdf",
					HashSum:     validHashSum,
					UserID:      userID,
				},
			},
			want: func(t *testing.T, fd *FileData) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, fd.ID)
				assert.Equal(t, userID, fd.UserID)
				assert.Equal(t, []byte("Important document"), fd.Description)
				assert.Equal(t, []byte("documents/file.pdf"), fd.StorageKey)
				assert.Equal(t, []byte(validHashSum), fd.HashSum)
				assert.WithinDuration(t, time.Now(), fd.UpdatedAt, time.Second)
			},
		},
		{
			name: "valid/minimal_file",
			args: args{
				params: NewFileDataParams{
					Description: "",
					StorageKey:  "file.txt",
					HashSum:     validHashSum,
					UserID:      userID,
				},
			},
			want: func(t *testing.T, fd *FileData) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, fd.ID)
				assert.Equal(t, userID, fd.UserID)
				assert.Equal(t, []byte(""), fd.Description)
				assert.Equal(t, []byte("file.txt"), fd.StorageKey)
				assert.Equal(t, []byte(validHashSum), fd.HashSum)
			},
		},
		{
			name: "valid/normalized_storage_key",
			args: args{
				params: NewFileDataParams{
					Description: "Test file",
					StorageKey:  `documents\folder\file.txt`,
					HashSum:     validHashSum,
					UserID:      userID,
				},
			},
			want: func(t *testing.T, fd *FileData) {
				t.Helper()
				assert.Equal(t, []byte("documents/folder/file.txt"), fd.StorageKey)
			},
		},
		{
			name: "valid/trimmed_and_lowercase_hash",
			args: args{
				params: NewFileDataParams{
					Description: "Test file",
					StorageKey:  "file.txt",
					HashSum:     "  " + strings.ToUpper(validHashSum) + "  ",
					UserID:      userID,
				},
			},
			want: func(t *testing.T, fd *FileData) {
				t.Helper()
				assert.Equal(t, []byte(validHashSum), fd.HashSum)
			},
		},
		{
			name: "invalid/empty_storage_key",
			args: args{
				params: NewFileDataParams{
					Description: "Test file",
					StorageKey:  "",
					HashSum:     validHashSum,
					UserID:      userID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid/path_traversal_storage_key",
			args: args{
				params: NewFileDataParams{
					Description: "Test file",
					StorageKey:  "../../../etc/passwd",
					HashSum:     validHashSum,
					UserID:      userID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid/invalid_hash_sum",
			args: args{
				params: NewFileDataParams{
					Description: "Test file",
					StorageKey:  "file.txt",
					HashSum:     "invalid_hash",
					UserID:      userID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid/multiple_errors",
			args: args{
				params: NewFileDataParams{
					Description: "Test file",
					StorageKey:  "",
					HashSum:     "invalid",
					UserID:      userID,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewFile(tt.args.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestNewFileDataParams_Validate(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	validHashSum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	tests := []struct {
		errType error
		name    string
		params  NewFileDataParams
		wantErr bool
	}{
		{
			name: "valid/standard_file",
			params: NewFileDataParams{
				Description: "Test document",
				StorageKey:  "documents/test.pdf",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "valid/simple_filename",
			params: NewFileDataParams{
				Description: "",
				StorageKey:  "file.txt",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "valid/nested_path",
			params: NewFileDataParams{
				Description: "Nested file",
				StorageKey:  "folder/subfolder/file.doc",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "invalid/empty_storage_key",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectStorageKey,
		},
		{
			name: "invalid/absolute_path",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "/absolute/path",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectStorageKey,
		},
		{
			name: "invalid/path_traversal",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "../parent/file.txt",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectStorageKey,
		},
		{
			name: "invalid/current_directory",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "./file.txt",
				HashSum:     validHashSum,
				UserID:      userID,
			},
			wantErr: false, // ./ prefix is normalized away
		},
		{
			name: "invalid/empty_hash_sum",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "file.txt",
				HashSum:     "",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectHashSum,
		},
		{
			name: "invalid/short_hash_sum",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "file.txt",
				HashSum:     "abc123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectHashSum,
		},
		{
			name: "invalid/long_hash_sum",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "file.txt",
				HashSum:     validHashSum + "extra",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectHashSum,
		},
		{
			name: "invalid/non_hex_hash_sum",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "file.txt",
				HashSum:     "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // 'g' is not hex
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectHashSum,
		},
		{
			name: "invalid/multiple_errors",
			params: NewFileDataParams{
				Description: "Test",
				StorageKey:  "",
				HashSum:     "invalid",
				UserID:      userID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidStorageKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "valid/simple_filename",
			key:  "file.txt",
			want: true,
		},
		{
			name: "valid/nested_path",
			key:  "folder/subfolder/file.doc",
			want: true,
		},
		{
			name: "valid/path_with_extension",
			key:  "documents/important.pdf",
			want: true,
		},
		{
			name: "valid/deep_nesting",
			key:  "a/b/c/d/e/file.txt",
			want: true,
		},
		{
			name: "invalid/empty_string",
			key:  "",
			want: false,
		},
		{
			name: "invalid/absolute_unix_path",
			key:  "/etc/passwd",
			want: false,
		},
		{
			name: "invalid/absolute_windows_path",
			key:  `\Windows\System32`,
			want: false,
		},
		{
			name: "invalid/parent_directory",
			key:  "../file.txt",
			want: false,
		},
		{
			name: "valid/nested_parent_directory_resolved",
			key:  "folder/../file.txt",
			want: true, // filepath.Clean resolves this to "file.txt"
		},
		{
			name: "valid/current_directory_cleaned",
			key:  "./file.txt",
			want: true, // filepath.Clean resolves this to "file.txt"
		},
		{
			name: "valid/double_dot_resolved",
			key:  "folder/../other/file.txt",
			want: true, // filepath.Clean resolves this to "other/file.txt"
		},
		{
			name: "valid/empty_segment_cleaned",
			key:  "folder//file.txt",
			want: true, // filepath.Clean resolves this to "folder/file.txt"
		},
		{
			name: "valid/trailing_slash_cleaned",
			key:  "folder/file.txt/",
			want: true, // filepath.Clean resolves this to "folder/file.txt"
		},
		{
			name: "invalid/only_dots",
			key:  "..",
			want: false,
		},
		{
			name: "invalid/only_single_dot",
			key:  ".",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := validStorageKey(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsHexSHA256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid/empty_string_hash",
			s:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			want: true,
		},
		{
			name: "valid/uppercase_hex",
			s:    "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			want: true,
		},
		{
			name: "valid/mixed_case_hex",
			s:    "e3B0c44298Fc1c149afbF4c8996fb92427ae41E4649b934ca495991b7852B855",
			want: true,
		},
		{
			name: "valid/all_zeros",
			s:    "0000000000000000000000000000000000000000000000000000000000000000",
			want: true,
		},
		{
			name: "valid/all_fs",
			s:    "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			want: true,
		},
		{
			name: "invalid/too_short",
			s:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85",
			want: false,
		},
		{
			name: "invalid/too_long",
			s:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8551",
			want: false,
		},
		{
			name: "invalid/non_hex_characters",
			s:    "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			want: false,
		},
		{
			name: "invalid/empty_string",
			s:    "",
			want: false,
		},
		{
			name: "invalid/spaces",
			s:    "e3b0c442 98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			want: false,
		},
		{
			name: "invalid/special_characters",
			s:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85!",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isHexSHA256(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeSlash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "unix_path_unchanged",
			path: "folder/file.txt",
			want: "folder/file.txt",
		},
		{
			name: "windows_path_normalized",
			path: `folder\file.txt`,
			want: "folder/file.txt",
		},
		{
			name: "mixed_separators",
			path: `folder\subfolder/file.txt`,
			want: "folder/subfolder/file.txt",
		},
		{
			name: "current_directory_prefix_removed",
			path: "./folder/file.txt",
			want: "folder/file.txt",
		},
		{
			name: "current_directory_prefix_windows",
			path: `.\\folder\file.txt`,
			want: "/folder/file.txt", // .\\ becomes //, then ./ is stripped leaving /folder/file.txt
		},
		{
			name: "empty_string",
			path: "",
			want: "",
		},
		{
			name: "only_current_directory",
			path: "./",
			want: "",
		},
		{
			name: "only_backslash",
			path: `\`,
			want: "/",
		},
		{
			name: "mixed_complex_path",
			path: `.\folder\subfolder/file.txt`,
			want: "folder/subfolder/file.txt",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeSlash(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}
