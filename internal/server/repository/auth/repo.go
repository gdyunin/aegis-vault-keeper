package auth

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/middleware"
)

// saveFunc defines the signature for user persistence operations with middleware support.
type saveFunc func(ctx context.Context, params SaveParams) error

// saveMw defines middleware type for save operations.
type saveMw = middleware.Middleware[saveFunc]

// loadFunc defines the signature for user retrieval operations with middleware support.
type loadFunc func(ctx context.Context, params LoadParams) (*auth.User, error)

// loadMw defines middleware type for load operations.
type loadMw = middleware.Middleware[loadFunc]

// Repository provides encrypted user data persistence with middleware-based encryption.
type Repository struct {
	// save is the middleware chain for user persistence operations.
	save saveFunc
	// load is the middleware chain for user retrieval operations.
	load loadFunc
}

// NewRepository creates a new user repository with encryption middleware and database client.
func NewRepository(dbClient db.DBClient, secretKey []byte) *Repository {
	return &Repository{
		save: middleware.Chain(rawSave(dbClient), encryptionMw(secretKey)),
		load: middleware.Chain(rawLoad(dbClient), decryptionMw(secretKey)),
	}
}

// Save stores a user with automatic encryption.
func (r *Repository) Save(ctx context.Context, params SaveParams) error {
	if err := r.save(ctx, params); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

// Load retrieves a user with automatic decryption.
func (r *Repository) Load(ctx context.Context, params LoadParams) (*auth.User, error) {
	u, err := r.load(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	return u, nil
}
