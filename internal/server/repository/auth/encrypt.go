package auth

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
)

// encryptionMw creates a middleware that encrypts user cryptographic keys before saving.
// Uses master secret key for encryption to protect user-specific encryption keys.
func encryptionMw(secretKey []byte) saveMw {
	return func(next saveFunc) saveFunc {
		return func(ctx context.Context, p SaveParams) error {
			copyEntity := *p.Entity

			encryptedKey, err := crypto.EncryptAESGCM(secretKey, copyEntity.CryptoKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt crypto key: %w", err)
			}
			copyEntity.CryptoKey = encryptedKey

			p.Entity = &copyEntity
			return next(ctx, p)
		}
	}
}

// decryptionMw creates a middleware that decrypts user cryptographic keys after loading.
// Uses master secret key for decryption to recover user-specific encryption keys.
func decryptionMw(secretKey []byte) loadMw {
	return func(next loadFunc) loadFunc {
		return func(ctx context.Context, p LoadParams) (*auth.User, error) {
			entity, err := next(ctx, p)
			if err != nil {
				return nil, fmt.Errorf("failed to load entity: %w", err)
			}

			decryptedKey, err := crypto.DecryptAESGCM(secretKey, entity.CryptoKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt crypto key: %w", err)
			}
			entity.CryptoKey = decryptedKey

			return entity, nil
		}
	}
}
