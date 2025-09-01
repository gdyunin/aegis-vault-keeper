package note

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	"github.com/google/uuid"
)

// SaveParams contains the parameters for saving a note entity to the repository.
type SaveParams struct {
	// Entity contains the note data to be persisted.
	Entity *note.Note
}

// LoadParams contains the parameters for loading note entities from the repository.
type LoadParams struct {
	// ID contains the specific note identifier for single record lookup (optional).
	ID uuid.UUID
	// UserID contains the user identifier for filtering notes by owner (required).
	UserID uuid.UUID
}
