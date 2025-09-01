package note

import (
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNote_ToApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	noteID := uuid.New()
	fixedTime := time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		note     *Note
		expected *note.Note
		name     string
		userID   uuid.UUID
	}{
		{
			name: "valid note conversion",
			note: &Note{
				ID:          noteID,
				Note:        "Important meeting notes",
				Description: "Meeting with client ABC",
				UpdatedAt:   fixedTime,
			},
			userID: userID,
			expected: &note.Note{
				ID:          noteID,
				UserID:      userID,
				Note:        "Important meeting notes",
				Description: "Meeting with client ABC",
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:     "nil note",
			note:     nil,
			userID:   userID,
			expected: nil,
		},
		{
			name: "empty fields",
			note: &Note{
				ID:          uuid.Nil,
				Note:        "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
			userID: userID,
			expected: &note.Note{
				ID:          uuid.Nil,
				UserID:      userID,
				Note:        "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
		},
		{
			name: "long note content",
			note: &Note{
				ID: noteID,
				Note: "This is a very long note that contains important information about the project " +
					"status and next steps to be taken",
				Description: "Project status update",
				UpdatedAt:   fixedTime,
			},
			userID: userID,
			expected: &note.Note{
				ID:     noteID,
				UserID: userID,
				Note: "This is a very long note that contains important information about the project " +
					"status and next steps to be taken",
				Description: "Project status update",
				UpdatedAt:   fixedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.note.ToApp(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNotesToApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	noteID1 := uuid.New()
	noteID2 := uuid.New()

	tests := []struct {
		name     string
		notes    []*Note
		expected []*note.Note
		userID   uuid.UUID
	}{
		{
			name: "multiple notes",
			notes: []*Note{
				{
					ID:          noteID1,
					Note:        "First note",
					Description: "First description",
				},
				{
					ID:          noteID2,
					Note:        "Second note",
					Description: "Second description",
				},
			},
			userID: userID,
			expected: []*note.Note{
				{
					ID:          noteID1,
					UserID:      userID,
					Note:        "First note",
					Description: "First description",
				},
				{
					ID:          noteID2,
					UserID:      userID,
					Note:        "Second note",
					Description: "Second description",
				},
			},
		},
		{
			name:     "nil slice",
			notes:    nil,
			userID:   userID,
			expected: nil,
		},
		{
			name:     "empty slice",
			notes:    []*Note{},
			userID:   userID,
			expected: []*note.Note{},
		},
		{
			name: "single note",
			notes: []*Note{
				{
					ID:          noteID1,
					Note:        "Single note content",
					Description: "Single note description",
				},
			},
			userID: userID,
			expected: []*note.Note{
				{
					ID:          noteID1,
					UserID:      userID,
					Note:        "Single note content",
					Description: "Single note description",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NotesToApp(tt.notes, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewNoteFromApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	noteID := uuid.New()
	fixedTime := time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		appNote  *note.Note
		expected *Note
		name     string
	}{
		{
			name: "valid app note conversion",
			appNote: &note.Note{
				ID:          noteID,
				UserID:      userID,
				Note:        "Important meeting notes",
				Description: "Meeting with client ABC",
				UpdatedAt:   fixedTime,
			},
			expected: &Note{
				ID:          noteID,
				Note:        "Important meeting notes",
				Description: "Meeting with client ABC",
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:     "nil app note",
			appNote:  nil,
			expected: nil,
		},
		{
			name: "empty fields",
			appNote: &note.Note{
				ID:          uuid.Nil,
				UserID:      userID,
				Note:        "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
			expected: &Note{
				ID:          uuid.Nil,
				Note:        "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
		},
		{
			name: "note without description",
			appNote: &note.Note{
				ID:          noteID,
				UserID:      userID,
				Note:        "Note without description",
				Description: "",
				UpdatedAt:   fixedTime,
			},
			expected: &Note{
				ID:          noteID,
				Note:        "Note without description",
				Description: "",
				UpdatedAt:   fixedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewNoteFromApp(tt.appNote)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewNotesFromApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	noteID1 := uuid.New()
	noteID2 := uuid.New()

	tests := []struct {
		name     string
		appNotes []*note.Note
		expected []*Note
	}{
		{
			name: "multiple app notes",
			appNotes: []*note.Note{
				{
					ID:          noteID1,
					UserID:      userID,
					Note:        "First note",
					Description: "First description",
				},
				{
					ID:          noteID2,
					UserID:      userID,
					Note:        "Second note",
					Description: "Second description",
				},
			},
			expected: []*Note{
				{
					ID:          noteID1,
					Note:        "First note",
					Description: "First description",
				},
				{
					ID:          noteID2,
					Note:        "Second note",
					Description: "Second description",
				},
			},
		},
		{
			name:     "nil slice",
			appNotes: nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			appNotes: []*note.Note{},
			expected: []*Note{},
		},
		{
			name: "mixed content",
			appNotes: []*note.Note{
				{
					ID:          noteID1,
					UserID:      userID,
					Note:        "Note with description",
					Description: "Has description",
				},
				{
					ID:          noteID2,
					UserID:      userID,
					Note:        "Note without description",
					Description: "",
				},
			},
			expected: []*Note{
				{
					ID:          noteID1,
					Note:        "Note with description",
					Description: "Has description",
				},
				{
					ID:          noteID2,
					Note:        "Note without description",
					Description: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewNotesFromApp(tt.appNotes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
