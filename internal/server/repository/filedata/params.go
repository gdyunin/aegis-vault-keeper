package filedata

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/google/uuid"
)

// SaveParams contains parameters for saving file data entities to the repository.
type SaveParams struct {
	// Entity is the file data to be saved.
	Entity *filedata.FileData
}

// LoadParams contains parameters for loading file data entities from the repository.
type LoadParams struct {
	// ID specifies the file data ID to load; zero value loads all user file data.
	ID uuid.UUID
	// UserID identifies the user whose file data to load.
	UserID uuid.UUID
}
