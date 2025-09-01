package bankcard

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/middleware"
)

// saveFunc defines the signature for bank card save operations.
type saveFunc func(ctx context.Context, params SaveParams) error

// saveMw defines middleware for save operations.
type saveMw = middleware.Middleware[saveFunc]

// loadFunc defines the signature for bank card load operations.
type loadFunc func(ctx context.Context, params LoadParams) ([]*bankcard.BankCard, error)

// loadMw defines middleware for load operations.
type loadMw = middleware.Middleware[loadFunc]

// Repository provides encrypted bank card data persistence with middleware support.
type Repository struct {
	// save is the function chain for saving bank card data with encryption middleware.
	save saveFunc
	// load is the function chain for loading bank card data with decryption middleware.
	load loadFunc
}

// NewRepository creates a new Repository with encryption middleware and database backend.
func NewRepository(dbClient db.DBClient, keyProvider keyprv.UserKeyProvider) *Repository {
	return &Repository{
		save: middleware.Chain(rawSave(dbClient), encryptionMw(keyProvider)),
		load: middleware.Chain(rawLoad(dbClient), decryptionMw(keyProvider)),
	}
}

// Save persists a bank card with automatic encryption of sensitive fields.
func (r *Repository) Save(ctx context.Context, params SaveParams) error {
	if err := r.save(ctx, params); err != nil {
		return fmt.Errorf("failed to save bank card: %w", err)
	}
	return nil
}

// Load retrieves bank cards with automatic decryption of sensitive fields.
func (r *Repository) Load(ctx context.Context, params LoadParams) ([]*bankcard.BankCard, error) {
	cards, err := r.load(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load bank cards: %w", err)
	}
	return cards, nil
}
