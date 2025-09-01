package filestorage

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
)

// encryptionMw creates middleware that encrypts file data before saving to storage.
func encryptionMw(keyProvider keyprv.UserKeyProvider) saveMw {
	return func(next saveFunc) saveFunc {
		return func(ctx context.Context, p SaveParams) error {
			k, err := keyProvider.UserKeyProvide(ctx, p.UserID)
			if err != nil {
				return fmt.Errorf("failed to get user key: %w", err)
			}

			encryptedData, err := crypto.EncryptAESGCM(k, p.Data)
			if err != nil {
				return fmt.Errorf("failed to encrypt file data: %w", err)
			}

			encryptedParams := SaveParams{
				UserID:     p.UserID,
				StorageKey: p.StorageKey,
				Data:       encryptedData,
			}

			return next(ctx, encryptedParams)
		}
	}
}

// decryptionMw creates middleware that decrypts file data after loading from storage.
func decryptionMw(keyProvider keyprv.UserKeyProvider) loadMw {
	return func(next loadFunc) loadFunc {
		return func(ctx context.Context, p LoadParams) ([]byte, error) {
			encryptedData, err := next(ctx, p)
			if err != nil {
				return nil, err
			}

			k, err := keyProvider.UserKeyProvide(ctx, p.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user key: %w", err)
			}

			decryptedData, err := crypto.DecryptAESGCM(k, encryptedData)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt file data: %w", err)
			}

			return decryptedData, nil
		}
	}
}
