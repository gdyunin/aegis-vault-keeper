package filestorage

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/middleware"
)

// saveFunc defines the signature for file storage save operations.
type saveFunc func(ctx context.Context, params SaveParams) error

// saveMw is middleware for file storage save operations.
type saveMw = middleware.Middleware[saveFunc]

// loadFunc defines the signature for file storage load operations.
type loadFunc func(ctx context.Context, params LoadParams) ([]byte, error)

// loadMw is middleware for file storage load operations.
type loadMw = middleware.Middleware[loadFunc]

// deleteFunc defines the signature for file storage delete operations.
type deleteFunc func(ctx context.Context, params DeleteParams) error

// Repository provides encrypted filesystem storage operations using middleware pattern.
type Repository struct {
	// save is the function chain for saving file data with encryption middleware.
	save saveFunc
	// load is the function chain for loading file data with decryption middleware.
	load loadFunc
	// delete is the function for removing files from the filesystem.
	delete deleteFunc
}

// NewRepository creates a new Repository with encryption/decryption middleware for filesystem storage.
func NewRepository(basePath string, keyProvider keyprv.UserKeyProvider) *Repository {
	return &Repository{
		save:   middleware.Chain(rawSave(basePath), encryptionMw(keyProvider)),
		load:   middleware.Chain(rawLoad(basePath), decryptionMw(keyProvider)),
		delete: rawDelete(basePath),
	}
}

// Save stores file data to filesystem with automatic encryption.
func (r *Repository) Save(ctx context.Context, params SaveParams) error {
	if err := r.save(ctx, params); err != nil {
		return fmt.Errorf("failed to save file to storage: %w", err)
	}
	return nil
}

// Load retrieves file data from filesystem with automatic decryption.
func (r *Repository) Load(ctx context.Context, params LoadParams) ([]byte, error) {
	data, err := r.load(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load file from storage: %w", err)
	}
	return data, nil
}

// Delete removes file data from filesystem.
func (r *Repository) Delete(ctx context.Context, params DeleteParams) error {
	if err := r.delete(ctx, params); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return nil
}
