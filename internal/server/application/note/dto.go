package note

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	"github.com/google/uuid"
)

// Note represents a note data transfer object for application layer communication.
type Note struct {
	// UpdatedAt indicates when the note was last modified.
	UpdatedAt time.Time
	// Note contains the note text content.
	Note string
	// Description contains additional information about the note.
	Description string
	// ID uniquely identifies the note.
	ID uuid.UUID
	// UserID identifies the note owner.
	UserID uuid.UUID
}

// newNoteFromDomain converts a domain note entity to application DTO.
func newNoteFromDomain(c *note.Note) *Note {
	if c == nil {
		return nil
	}
	return &Note{
		ID:          c.ID,
		UserID:      c.UserID,
		Note:        string(c.Note),
		Description: string(c.Description),
		UpdatedAt:   c.UpdatedAt,
	}
}

// newNotesFromDomain converts a slice of domain note entities to application DTOs.
func newNotesFromDomain(cs []*note.Note) []*Note {
	result := make([]*Note, 0, len(cs))
	for _, c := range cs {
		result = append(result, newNoteFromDomain(c))
	}
	return result
}

// PullParams contains parameters for retrieving a specific note.
type PullParams struct {
	// ID specifies the note to retrieve.
	ID uuid.UUID
	// UserID specifies the note owner.
	UserID uuid.UUID
}

// ListParams contains parameters for listing user notes.
type ListParams struct {
	// UserID specifies the note owner.
	UserID uuid.UUID
}

// PushParams contains parameters for creating or updating a note.
type PushParams struct {
	// Note specifies the note text content.
	Note string
	// Description provides additional information about the note.
	Description string
	// ID uniquely identifies the note.
	ID uuid.UUID
	// UserID identifies the note owner.
	UserID uuid.UUID
}
