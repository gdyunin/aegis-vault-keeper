package credential

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
)

// encryptionMw creates middleware that encrypts credential fields before saving to the database.
func encryptionMw(keyProvider keyprv.UserKeyProvider) saveMw {
	return func(next saveFunc) saveFunc {
		return func(ctx context.Context, p SaveParams) error {
			k, err := keyProvider.UserKeyProvide(ctx, p.Entity.UserID)
			if err != nil {
				return fmt.Errorf("failed to provide user key: %w", err)
			}

			copyEntity := *p.Entity

			if copyEntity.Login, err = crypto.EncryptAESGCM(k, copyEntity.Login); err != nil {
				return fmt.Errorf("failed to encrypt login: %w", err)
			}
			if copyEntity.Password, err = crypto.EncryptAESGCM(k, copyEntity.Password); err != nil {
				return fmt.Errorf("failed to encrypt password: %w", err)
			}
			if copyEntity.Description, err = crypto.EncryptAESGCM(k, copyEntity.Description); err != nil {
				return fmt.Errorf("failed to encrypt description: %w", err)
			}

			p.Entity = &copyEntity
			return next(ctx, p)
		}
	}
}

// DecryptionMw creates middleware that decrypts credential fields after loading from the database.
// All sensitive fields (login, password, description) are decrypted using AES-GCM with the user's encryption key.
func decryptionMw(keyProvider keyprv.UserKeyProvider) loadMw {
	return func(next loadFunc) loadFunc {
		return func(ctx context.Context, p LoadParams) ([]*credential.Credential, error) {
			entities, err := next(ctx, p)
			if err != nil {
				return nil, fmt.Errorf("failed to load entities: %w", err)
			}
			if len(entities) == 0 {
				return []*credential.Credential{}, nil
			}

			k, err := keyProvider.UserKeyProvide(ctx, p.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to provide user key: %w", err)
			}

			for _, entity := range entities {
				if entity.Login, err = crypto.DecryptAESGCM(k, entity.Login); err != nil {
					return nil, fmt.Errorf("failed to decrypt login: %w", err)
				}
				if entity.Password, err = crypto.DecryptAESGCM(k, entity.Password); err != nil {
					return nil, fmt.Errorf("failed to decrypt password: %w", err)
				}
				if entity.Description, err = crypto.DecryptAESGCM(k, entity.Description); err != nil {
					return nil, fmt.Errorf("failed to decrypt description: %w", err)
				}
			}

			return entities, nil
		}
	}
}
