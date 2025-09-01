package security

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
	"github.com/google/uuid"
)

// UserKeyRepository defines the interface for retrieving user cryptographic keys.
type UserKeyRepository interface {
	// Load retrieves user data including the cryptographic key using the provided parameters.
	Load(ctx context.Context, params repository.LoadParams) (*auth.User, error)
}

// UserKeyProvider provides access to user-specific cryptographic keys for encryption/decryption.
type UserKeyProvider struct {
	// r is the repository used to fetch user data and cryptographic keys.
	r UserKeyRepository
}

// NewUserKeyProvider creates a new UserKeyProvider with the specified repository.
func NewUserKeyProvider(r UserKeyRepository) *UserKeyProvider {
	return &UserKeyProvider{
		r: r,
	}
}

// UserKeyProvide retrieves the cryptographic key for the specified user ID.
func (p *UserKeyProvider) UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	u, err := p.r.Load(ctx, repository.LoadParams{ID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to load user with ID %s: %w", userID, err)
	}
	return u.CryptoKey, nil
}
