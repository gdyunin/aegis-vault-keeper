package auth

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	"github.com/google/uuid"
)

// SaveParams contains the parameters for saving a user entity to the repository.
type SaveParams struct {
	// Entity contains the user data to be persisted.
	Entity *auth.User
}

// LoadParams contains the parameters for loading a user entity from the repository.
type LoadParams struct {
	// Login contains the user's login identifier for lookup (alternative to ID).
	Login string
	// ID contains the user's unique identifier for lookup (alternative to Login).
	ID uuid.UUID
}
