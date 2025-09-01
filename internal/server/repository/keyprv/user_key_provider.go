package keyprv

import (
	"context"

	"github.com/google/uuid"
)

// UserKeyProvider interface for retrieving user-specific encryption keys.
// Implementations must provide secure key derivation and storage.
type UserKeyProvider interface {
	UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error)
}
