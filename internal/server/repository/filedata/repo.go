package filedata

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/middleware"
)

// saveFunc defines the signature for file data save operations.
type saveFunc func(ctx context.Context, params SaveParams) error

// saveMw is middleware for file data save operations.
type saveMw = middleware.Middleware[saveFunc]

// loadFunc defines the signature for file data load operations.
type loadFunc func(ctx context.Context, params LoadParams) ([]*filedata.FileData, error)

// loadMw is middleware for file data load operations.
type loadMw = middleware.Middleware[loadFunc]

// Repository provides encrypted file data storage operations using middleware pattern.
type Repository struct {
	// save is the function chain for saving file metadata with encryption middleware.
	save saveFunc
	// load is the function chain for loading file metadata with decryption middleware.
	load loadFunc
}

// NewRepository creates a new Repository with encryption/decryption middleware.
func NewRepository(dbClient db.DBClient, keyProvider keyprv.UserKeyProvider) *Repository {
	return &Repository{
		save: middleware.Chain(rawSave(dbClient), encryptionMw(keyProvider)),
		load: middleware.Chain(rawLoad(dbClient), decryptionMw(keyProvider)),
	}
}

// Save stores file data with automatic encryption.
func (r *Repository) Save(ctx context.Context, params SaveParams) error {
	if err := r.save(ctx, params); err != nil {
		return fmt.Errorf("failed to save files: %w", err)
	}
	return nil
}

// Load retrieves file data with automatic decryption.
func (r *Repository) Load(ctx context.Context, params LoadParams) ([]*filedata.FileData, error) {
	fds, err := r.load(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load files: %w", err)
	}
	return fds, nil
}
