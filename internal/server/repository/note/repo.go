package note

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/middleware"
)

// saveFunc defines the signature for note save operations.
type saveFunc func(ctx context.Context, params SaveParams) error

// saveMw defines middleware for save operations.
type saveMw = middleware.Middleware[saveFunc]

// loadFunc defines the signature for note load operations.
type loadFunc func(ctx context.Context, params LoadParams) ([]*note.Note, error)

// loadMw defines middleware for load operations.
type loadMw = middleware.Middleware[loadFunc]

// Repository provides encrypted note data persistence with middleware support.
type Repository struct {
	// save is the function chain for saving note data with encryption middleware.
	save saveFunc
	// load is the function chain for loading note data with decryption middleware.
	load loadFunc
}

// NewRepository creates a new Repository with encryption middleware and database backend.
func NewRepository(dbClient db.DBClient, keyProvider keyprv.UserKeyProvider) *Repository {
	return &Repository{
		save: middleware.Chain(rawSave(dbClient), encryptionMw(keyProvider)),
		load: middleware.Chain(rawLoad(dbClient), decryptionMw(keyProvider)),
	}
}

// Save persists a note with automatic encryption of sensitive fields.
func (r *Repository) Save(ctx context.Context, params SaveParams) error {
	if err := r.save(ctx, params); err != nil {
		return fmt.Errorf("failed to save note: %w", err)
	}
	return nil
}

// Load retrieves notes with automatic decryption of sensitive fields.
func (r *Repository) Load(ctx context.Context, params LoadParams) ([]*note.Note, error) {
	notes, err := r.load(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", err)
	}
	return notes, nil
}
