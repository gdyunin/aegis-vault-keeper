package note

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
)

// encryptionMw creates a middleware that encrypts note content before saving.
// Both note content and description fields are encrypted using AES-GCM.
func encryptionMw(keyProvider keyprv.UserKeyProvider) saveMw {
	return func(next saveFunc) saveFunc {
		return func(ctx context.Context, p SaveParams) error {
			k, err := keyProvider.UserKeyProvide(ctx, p.Entity.UserID)
			if err != nil {
				return fmt.Errorf("failed to provide user key: %w", err)
			}

			copyEntity := *p.Entity

			if copyEntity.Note, err = crypto.EncryptAESGCM(k, copyEntity.Note); err != nil {
				return fmt.Errorf("failed to encrypt note: %w", err)
			}

			if copyEntity.Description, err = crypto.EncryptAESGCM(k, copyEntity.Description); err != nil {
				return fmt.Errorf("failed to encrypt description: %w", err)
			}

			p.Entity = &copyEntity
			return next(ctx, p)
		}
	}
}

// decryptionMw creates middleware that decrypts note entities after loading from storage.
func decryptionMw(keyProvider keyprv.UserKeyProvider) loadMw {
	return func(next loadFunc) loadFunc {
		return func(ctx context.Context, p LoadParams) ([]*note.Note, error) {
			entities, err := next(ctx, p)
			if err != nil {
				return nil, fmt.Errorf("failed to load entities: %w", err)
			}
			if len(entities) == 0 {
				return []*note.Note{}, nil
			}

			k, err := keyProvider.UserKeyProvide(ctx, p.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to provide user key: %w", err)
			}

			for _, entity := range entities {
				if entity.Note, err = crypto.DecryptAESGCM(k, entity.Note); err != nil {
					return nil, fmt.Errorf("failed to decrypt note: %w", err)
				}
				if entity.Description, err = crypto.DecryptAESGCM(k, entity.Description); err != nil {
					return nil, fmt.Errorf("failed to decrypt description: %w", err)
				}
			}

			return entities, nil
		}
	}
}
