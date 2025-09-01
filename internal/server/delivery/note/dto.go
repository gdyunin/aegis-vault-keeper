package note

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/google/uuid"
)

// Note represents a text note entity.
type Note struct {
	// UpdatedAt contains the last modification timestamp.
	UpdatedAt time.Time `json:"updated_at,omitzero"  example:"2023-12-01T10:00:00Z"`
	// Note contains the text content (required, max 1000 chars).
	Note string `json:"note,omitzero"        example:"Important meeting notes"`
	// Description contains optional metadata description (max 255 chars).
	Description string `json:"description,omitzero" example:"Meeting with client ABC"`
	// ID contains the unique note identifier.
	ID uuid.UUID `json:"id,omitzero"          example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ToApp converts delivery DTO to application layer Note entity.
func (n *Note) ToApp(userID uuid.UUID) *note.Note {
	if n == nil {
		return nil
	}
	return &note.Note{
		ID:          n.ID,
		UserID:      userID,
		Note:        n.Note,
		Description: n.Description,
		UpdatedAt:   n.UpdatedAt,
	}
}

// NotesToApp converts a slice of delivery DTOs to application layer Note entities.
func NotesToApp(notes []*Note, userID uuid.UUID) []*note.Note {
	if notes == nil {
		return nil
	}
	result := make([]*note.Note, 0, len(notes))
	for _, n := range notes {
		result = append(result, n.ToApp(userID))
	}
	return result
}

// NewNoteFromApp converts an application layer Note entity to delivery DTO.
func NewNoteFromApp(n *note.Note) *Note {
	if n == nil {
		return nil
	}
	return &Note{
		ID:          n.ID,
		Note:        n.Note,
		Description: n.Description,
		UpdatedAt:   n.UpdatedAt,
	}
}

// NewNotesFromApp converts a slice of application layer Note entities to delivery DTOs.
func NewNotesFromApp(notes []*note.Note) []*Note {
	if notes == nil {
		return nil
	}
	result := make([]*Note, 0, len(notes))
	for _, n := range notes {
		result = append(result, NewNoteFromApp(n))
	}
	return result
}

// PushRequest represents the data required to create or update a note.
type PushRequest struct {
	// Note contains the text content (required, max 1000 chars).
	Note string `json:"note"                 binding:"required" example:"Important meeting notes"`
	// Description contains optional metadata description (max 255 chars).
	Description string `json:"description,omitzero"                    example:"Meeting with client ABC"`
}

// PullRequest represents the request to retrieve a specific note.
type PullRequest struct {
	// ID contains the note identifier (required UUID format).
	ID string `uri:"id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// PushResponse represents the response after creating or updating a note.
type PushResponse struct {
	// ID contains the created or updated note identifier.
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// PullResponse represents the response containing a specific note.
type PullResponse struct {
	// Note contains the requested note data.
	Note *Note `json:"note"`
}

// ListResponse represents the response containing all user's notes.
type ListResponse struct {
	// Notes contains all notes belonging to the authenticated user.
	Notes []*Note `json:"notes"`
}
