package note

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNote(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		errorType   error
		name        string
		params      NewNoteParams
		expectError bool
	}{
		{
			name: "valid note with description",
			params: NewNoteParams{
				Note:        "Test note content",
				Description: "Test description",
				UserID:      userID,
			},
			expectError: false,
		},
		{
			name: "valid note without description",
			params: NewNoteParams{
				Note:        "Test note content",
				Description: "",
				UserID:      userID,
			},
			expectError: false,
		},
		{
			name: "empty note text",
			params: NewNoteParams{
				Note:        "",
				Description: "Test description",
				UserID:      userID,
			},
			expectError: true,
			errorType:   ErrNewNoteParamsValidation,
		},
		{
			name: "whitespace only note",
			params: NewNoteParams{
				Note:        "   ",
				Description: "Test description",
				UserID:      userID,
			},
			expectError: false, // Whitespace is considered valid content
		},
		{
			name: "zero uuid",
			params: NewNoteParams{
				Note:        "Test note",
				Description: "Test description",
				UserID:      uuid.UUID{},
			},
			expectError: false, // Zero UUID is valid from domain perspective
		},
		{
			name: "long note content",
			params: NewNoteParams{
				Note:        string(make([]byte, 10000)),
				Description: "Test description",
				UserID:      userID,
			},
			expectError: false,
		},
		{
			name: "unicode content",
			params: NewNoteParams{
				Note:        "测试笔记内容 🚀 emoji test",
				Description: "描述 with special chars: @#$%^&*()",
				UserID:      userID,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			note, err := NewNote(tt.params)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, note)
				assert.ErrorIs(t, err, tt.errorType)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, note)

			// Verify note fields
			assert.NotEqual(t, uuid.UUID{}, note.ID)
			assert.Equal(t, tt.params.UserID, note.UserID)
			assert.Equal(t, []byte(tt.params.Note), note.Note)
			assert.Equal(t, []byte(tt.params.Description), note.Description)
			assert.WithinDuration(t, time.Now(), note.UpdatedAt, time.Second)
		})
	}
}

func TestNewNoteParams_Validate(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		errorType   error
		name        string
		params      NewNoteParams
		expectError bool
	}{
		{
			name: "valid parameters",
			params: NewNoteParams{
				Note:        "Valid note",
				Description: "Valid description",
				UserID:      userID,
			},
			expectError: false,
		},
		{
			name: "empty note text",
			params: NewNoteParams{
				Note:        "",
				Description: "Valid description",
				UserID:      userID,
			},
			expectError: true,
			errorType:   ErrIncorrectNoteText,
		},
		{
			name: "note with only spaces",
			params: NewNoteParams{
				Note:        "   ",
				Description: "Valid description",
				UserID:      userID,
			},
			expectError: false, // Spaces are valid content
		},
		{
			name: "empty description is valid",
			params: NewNoteParams{
				Note:        "Valid note",
				Description: "",
				UserID:      userID,
			},
			expectError: false,
		},
		{
			name: "zero UUID is valid",
			params: NewNoteParams{
				Note:        "Valid note",
				Description: "Valid description",
				UserID:      uuid.UUID{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewNoteParams_validateNote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		noteText    string
		expectError bool
	}{
		{
			name:        "valid note text",
			noteText:    "This is a valid note",
			expectError: false,
		},
		{
			name:        "empty note text",
			noteText:    "",
			expectError: true,
		},
		{
			name:        "whitespace note text",
			noteText:    "   ",
			expectError: false,
		},
		{
			name:        "single character",
			noteText:    "a",
			expectError: false,
		},
		{
			name:        "unicode characters",
			noteText:    "测试 🚀",
			expectError: false,
		},
		{
			name:        "newline characters",
			noteText:    "line1\nline2\n",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := &NewNoteParams{
				Note: tt.noteText,
			}

			err := params.validateNote()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrIncorrectNoteText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNote_FieldValidation(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	params := NewNoteParams{
		Note:        "Test note",
		Description: "Test description",
		UserID:      userID,
	}

	note, err := NewNote(params)
	require.NoError(t, err)
	require.NotNil(t, note)

	// Test that all fields are properly set
	assert.NotEqual(t, uuid.UUID{}, note.ID, "Note ID should not be zero")
	assert.Equal(t, userID, note.UserID, "User ID should match input")
	assert.Equal(t, "Test note", string(note.Note), "Note content should match input")
	assert.Equal(t, "Test description", string(note.Description), "Description should match input")
	assert.True(t, time.Since(note.UpdatedAt) < time.Second, "UpdatedAt should be recent")
}

func TestNote_ByteArrayConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		noteText    string
		description string
	}{
		{
			name:        "ascii text",
			noteText:    "Simple ASCII text",
			description: "ASCII description",
		},
		{
			name:        "unicode text",
			noteText:    "Unicode: 测试 🚀 ñoño",
			description: "Unicode desc: café résumé",
		},
		{
			name:        "empty strings",
			noteText:    "",
			description: "",
		},
		{
			name:        "special characters",
			noteText:    "Special: !@#$%^&*()",
			description: "More special: <>/\\|[]{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Skip empty note text as it's invalid
			if tt.noteText == "" {
				return
			}

			params := NewNoteParams{
				Note:        tt.noteText,
				Description: tt.description,
				UserID:      uuid.New(),
			}

			note, err := NewNote(params)
			require.NoError(t, err)

			// Verify round-trip conversion
			assert.Equal(t, tt.noteText, string(note.Note))
			assert.Equal(t, tt.description, string(note.Description))
			assert.Equal(t, []byte(tt.noteText), note.Note)
			assert.Equal(t, []byte(tt.description), note.Description)
		})
	}
}
