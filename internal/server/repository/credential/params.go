package credential

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	"github.com/google/uuid"
)

// SaveParams contains parameters for saving credential entities to the repository.
type SaveParams struct {
	// Entity is the credential to be saved.
	Entity *credential.Credential
}

// LoadParams contains parameters for loading credential entities from the repository.
type LoadParams struct {
	// ID specifies the credential ID to load; zero value loads all user credentials.
	ID uuid.UUID
	// UserID identifies the user whose credentials to load.
	UserID uuid.UUID
}
