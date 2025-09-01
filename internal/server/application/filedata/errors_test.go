package filedata

import (
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   error
		want    error
		name    string
		wantNil bool
	}{
		{
			name:    "nil_error/returns_nil",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name:  "domain_validation_error/maps_to_app_error",
			input: filedata.ErrNewFileParamsValidation,
			want:  ErrFileAppError,
		},
		{
			name:  "domain_storage_key_error/maps_to_specific_error",
			input: filedata.ErrIncorrectStorageKey,
			want:  ErrFileIncorrectStorageKey,
		},
		{
			name:  "domain_hash_sum_error/maps_to_specific_error",
			input: filedata.ErrIncorrectHashSum,
			want:  ErrFileIncorrectHashSum,
		},
		{
			name:  "unknown_error/maps_to_tech_error",
			input: errors.New("unknown database error"),
			want:  ErrFileTechError,
		},
		{
			name:  "wrapped_domain_error/maps_correctly",
			input: errors.Join(errors.New("wrapper"), filedata.ErrIncorrectStorageKey),
			want:  ErrFileIncorrectStorageKey,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mapError(tt.input)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			if tt.want == ErrFileTechError {
				// For tech errors, check that it contains both tech error and original
				assert.True(t, errors.Is(got, ErrFileTechError), "should contain tech error")
				assert.Contains(t, got.Error(), tt.input.Error(), "should contain original error message")
			} else {
				assert.True(t, errors.Is(got, tt.want), "should contain expected error type")
			}
		})
	}
}

func TestMapFn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input      error
		wantErr    error
		name       string
		wantIsTech bool
	}{
		{
			name:    "domain_validation_error/maps_to_app_error",
			input:   filedata.ErrNewFileParamsValidation,
			wantErr: ErrFileAppError,
		},
		{
			name:    "domain_storage_key_error/maps_to_specific_error",
			input:   filedata.ErrIncorrectStorageKey,
			wantErr: ErrFileIncorrectStorageKey,
		},
		{
			name:    "domain_hash_sum_error/maps_to_specific_error",
			input:   filedata.ErrIncorrectHashSum,
			wantErr: ErrFileIncorrectHashSum,
		},
		{
			name:       "unknown_error/maps_to_tech_error",
			input:      errors.New("unknown error"),
			wantErr:    ErrFileTechError,
			wantIsTech: true,
		},
		{
			name:    "wrapped_validation_error/maps_correctly",
			input:   errors.Join(errors.New("wrapper"), filedata.ErrNewFileParamsValidation),
			wantErr: ErrFileAppError,
		},
		{
			name:    "wrapped_storage_key_error/maps_correctly",
			input:   errors.Join(errors.New("wrapper"), filedata.ErrIncorrectStorageKey),
			wantErr: ErrFileIncorrectStorageKey,
		},
		{
			name:    "wrapped_hash_sum_error/maps_correctly",
			input:   errors.Join(errors.New("wrapper"), filedata.ErrIncorrectHashSum),
			wantErr: ErrFileIncorrectHashSum,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mapFn(tt.input)

			require.NotNil(t, got)
			assert.True(t, errors.Is(got, tt.wantErr), "should contain expected error type")

			if tt.wantIsTech {
				// Tech errors should contain the original error
				assert.Contains(t, got.Error(), tt.input.Error(), "should contain original error message")
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrFileAppError",
			err:  ErrFileAppError,
			want: "file application error",
		},
		{
			name: "ErrFileTechError",
			err:  ErrFileTechError,
			want: "file technical error",
		},
		{
			name: "ErrFileIncorrectStorageKey",
			err:  ErrFileIncorrectStorageKey,
			want: "incorrect storage key",
		},
		{
			name: "ErrFileIncorrectHashSum",
			err:  ErrFileIncorrectHashSum,
			want: "incorrect hash sum",
		},
		{
			name: "ErrFileDataRequired",
			err:  ErrFileDataRequired,
			want: "file data is required",
		},
		{
			name: "ErrRollBackFileSaveFailed",
			err:  ErrRollBackFileSaveFailed,
			want: "rollback of file save failed",
		},
		{
			name: "ErrFileNotFound",
			err:  ErrFileNotFound,
			want: "file not found",
		},
		{
			name: "ErrFileAccessDenied",
			err:  ErrFileAccessDenied,
			want: "access to this file is denied",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.err.Error())
			assert.NotNil(t, tt.err)
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	errors := []error{
		ErrFileAppError,
		ErrFileTechError,
		ErrFileIncorrectStorageKey,
		ErrFileIncorrectHashSum,
		ErrFileDataRequired,
		ErrRollBackFileSaveFailed,
		ErrFileNotFound,
		ErrFileAccessDenied,
	}

	// Check that all errors are distinct
	for i := range errors {
		for j := i + 1; j < len(errors); j++ {
			assert.NotEqual(t, errors[i], errors[j],
				"errors should be distinct: %v vs %v", errors[i], errors[j])
		}
	}
}
