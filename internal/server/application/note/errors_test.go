package note

import (
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
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
			input: note.ErrNewNoteParamsValidation,
			want:  ErrNoteAppError,
		},
		{
			name:  "domain_incorrect_text_error/maps_to_specific_error",
			input: note.ErrIncorrectNoteText,
			want:  ErrNoteIncorrectNoteText,
		},
		{
			name:  "unknown_error/maps_to_tech_error",
			input: errors.New("unknown database error"),
			want:  ErrNoteTechError,
		},
		{
			name:  "wrapped_domain_error/maps_correctly",
			input: errors.Join(errors.New("wrapper"), note.ErrIncorrectNoteText),
			want:  ErrNoteIncorrectNoteText,
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
			if tt.want == ErrNoteTechError {
				// For tech errors, check that it contains both tech error and original
				assert.True(t, errors.Is(got, ErrNoteTechError), "should contain tech error")
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
			input:   note.ErrNewNoteParamsValidation,
			wantErr: ErrNoteAppError,
		},
		{
			name:    "domain_incorrect_text_error/maps_to_specific_error",
			input:   note.ErrIncorrectNoteText,
			wantErr: ErrNoteIncorrectNoteText,
		},
		{
			name:       "unknown_error/maps_to_tech_error",
			input:      errors.New("unknown error"),
			wantErr:    ErrNoteTechError,
			wantIsTech: true,
		},
		{
			name:    "wrapped_validation_error/maps_correctly",
			input:   errors.Join(errors.New("wrapper"), note.ErrNewNoteParamsValidation),
			wantErr: ErrNoteAppError,
		},
		{
			name:    "wrapped_text_error/maps_correctly",
			input:   errors.Join(errors.New("wrapper"), note.ErrIncorrectNoteText),
			wantErr: ErrNoteIncorrectNoteText,
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
			name: "ErrNoteAppError",
			err:  ErrNoteAppError,
			want: "note application error",
		},
		{
			name: "ErrNoteTechError",
			err:  ErrNoteTechError,
			want: "note technical error",
		},
		{
			name: "ErrNoteIncorrectNoteText",
			err:  ErrNoteIncorrectNoteText,
			want: "incorrect note text",
		},
		{
			name: "ErrNoteNotFound",
			err:  ErrNoteNotFound,
			want: "note not found",
		},
		{
			name: "ErrNoteAccessDenied",
			err:  ErrNoteAccessDenied,
			want: "access to this note is denied",
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
		ErrNoteAppError,
		ErrNoteTechError,
		ErrNoteIncorrectNoteText,
		ErrNoteNotFound,
		ErrNoteAccessDenied,
	}

	// Check that all errors are distinct
	for i := range errors {
		for j := i + 1; j < len(errors); j++ {
			assert.NotEqual(t, errors[i], errors[j],
				"errors should be distinct: %v vs %v", errors[i], errors[j])
		}
	}
}
