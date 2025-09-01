package filedata

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
)

// encryptionMw creates middleware that encrypts file data fields before saving to the database.
func encryptionMw(keyProvider keyprv.UserKeyProvider) saveMw {
	return func(next saveFunc) saveFunc {
		return func(ctx context.Context, p SaveParams) error {
			k, err := keyProvider.UserKeyProvide(ctx, p.Entity.UserID)
			if err != nil {
				return fmt.Errorf("failed to provide user key: %w", err)
			}

			copyEntity := *p.Entity

			if copyEntity.StorageKey, err = crypto.EncryptAESGCM(k, copyEntity.StorageKey); err != nil {
				return fmt.Errorf("failed to encrypt storage key: %w", err)
			}
			if copyEntity.HashSum, err = crypto.EncryptAESGCM(k, copyEntity.HashSum); err != nil {
				return fmt.Errorf("failed to encrypt hash sum: %w", err)
			}
			if copyEntity.Description, err = crypto.EncryptAESGCM(k, copyEntity.Description); err != nil {
				return fmt.Errorf("failed to encrypt description: %w", err)
			}

			p.Entity = &copyEntity
			return next(ctx, p)
		}
	}
}

// decryptionMw creates middleware that decrypts file data fields after loading from the database.
func decryptionMw(keyProvider keyprv.UserKeyProvider) loadMw {
	return func(next loadFunc) loadFunc {
		return func(ctx context.Context, p LoadParams) ([]*filedata.FileData, error) {
			entities, err := next(ctx, p)
			if err != nil {
				return nil, fmt.Errorf("failed to load entities: %w", err)
			}
			if len(entities) == 0 {
				return []*filedata.FileData{}, nil
			}

			k, err := keyProvider.UserKeyProvide(ctx, p.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to provide user key: %w", err)
			}

			for _, entity := range entities {
				if entity.StorageKey, err = crypto.DecryptAESGCM(k, entity.StorageKey); err != nil {
					return nil, fmt.Errorf("failed to decrypt storage key: %w", err)
				}
				if entity.HashSum, err = crypto.DecryptAESGCM(k, entity.HashSum); err != nil {
					return nil, fmt.Errorf("failed to decrypt hash sum: %w", err)
				}
				if entity.Description, err = crypto.DecryptAESGCM(k, entity.Description); err != nil {
					return nil, fmt.Errorf("failed to decrypt description: %w", err)
				}
			}

			return entities, nil
		}
	}
}
