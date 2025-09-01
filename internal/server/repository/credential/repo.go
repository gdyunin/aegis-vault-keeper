package credential

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/middleware"
)

// saveFunc defines the signature for credential save operations.
type saveFunc func(ctx context.Context, params SaveParams) error

// saveMw is middleware for credential save operations.
type saveMw = middleware.Middleware[saveFunc]

// loadFunc defines the signature for credential load operations.
type loadFunc func(ctx context.Context, params LoadParams) ([]*credential.Credential, error)

// loadMw is middleware for credential load operations.
type loadMw = middleware.Middleware[loadFunc]

// Repository provides encrypted credential storage operations using middleware pattern.
type Repository struct {
	// save is the function chain for saving credential data with encryption middleware.
	save saveFunc
	// load is the function chain for loading credential data with decryption middleware.
	load loadFunc
}

// NewRepository creates a new Repository with encryption/decryption middleware.
func NewRepository(dbClient db.DBClient, keyProvider keyprv.UserKeyProvider) *Repository {
	return &Repository{
		save: middleware.Chain(rawSave(dbClient), encryptionMw(keyProvider)),
		load: middleware.Chain(rawLoad(dbClient), decryptionMw(keyProvider)),
	}
}

// Save stores a credential with automatic encryption.
func (r *Repository) Save(ctx context.Context, params SaveParams) error {
	if err := r.save(ctx, params); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	return nil
}

// Load retrieves credentials with automatic decryption.
func (r *Repository) Load(ctx context.Context, params LoadParams) ([]*credential.Credential, error) {
	creds, err := r.load(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	return creds, nil
}
