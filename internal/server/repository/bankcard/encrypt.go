package bankcard

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
)

// encryptionMw creates a middleware that encrypts bank card data before saving.
// All sensitive fields (card number, holder, expiry, CVV, description) are encrypted using AES-GCM.
func encryptionMw(keyProvider keyprv.UserKeyProvider) saveMw {
	return func(next saveFunc) saveFunc {
		return func(ctx context.Context, p SaveParams) error {
			k, err := keyProvider.UserKeyProvide(ctx, p.Entity.UserID)
			if err != nil {
				return fmt.Errorf("failed to provide user key: %w", err)
			}

			copyEntity := *p.Entity

			if copyEntity.CardNumber, err = crypto.EncryptAESGCM(k, copyEntity.CardNumber); err != nil {
				return fmt.Errorf("failed to encrypt card number: %w", err)
			}
			if copyEntity.CardHolder, err = crypto.EncryptAESGCM(k, copyEntity.CardHolder); err != nil {
				return fmt.Errorf("failed to encrypt card holder: %w", err)
			}
			if copyEntity.ExpiryMonth, err = crypto.EncryptAESGCM(k, copyEntity.ExpiryMonth); err != nil {
				return fmt.Errorf("failed to encrypt expiry month: %w", err)
			}
			if copyEntity.ExpiryYear, err = crypto.EncryptAESGCM(k, copyEntity.ExpiryYear); err != nil {
				return fmt.Errorf("failed to encrypt expiry year: %w", err)
			}
			if copyEntity.CVV, err = crypto.EncryptAESGCM(k, copyEntity.CVV); err != nil {
				return fmt.Errorf("failed to encrypt CVV: %w", err)
			}
			if copyEntity.Description, err = crypto.EncryptAESGCM(k, copyEntity.Description); err != nil {
				return fmt.Errorf("failed to encrypt description: %w", err)
			}

			p.Entity = &copyEntity
			return next(ctx, p)
		}
	}
}

// decryptionMw creates a middleware that decrypts bank card data after loading.
// All sensitive fields are decrypted using AES-GCM with the user's encryption key.
func decryptionMw(keyProvider keyprv.UserKeyProvider) loadMw {
	return func(next loadFunc) loadFunc {
		return func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
			entities, err := next(ctx, p)
			if err != nil {
				return nil, fmt.Errorf("failed to load entities: %w", err)
			}

			if len(entities) == 0 {
				return []*bankcard.BankCard{}, nil
			}

			k, err := keyProvider.UserKeyProvide(ctx, p.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to provide user key: %w", err)
			}

			for _, entity := range entities {
				if entity.CardNumber, err = crypto.DecryptAESGCM(k, entity.CardNumber); err != nil {
					return nil, fmt.Errorf("failed to decrypt card number: %w", err)
				}
				if entity.CardHolder, err = crypto.DecryptAESGCM(k, entity.CardHolder); err != nil {
					return nil, fmt.Errorf("failed to decrypt card holder: %w", err)
				}
				if entity.ExpiryMonth, err = crypto.DecryptAESGCM(k, entity.ExpiryMonth); err != nil {
					return nil, fmt.Errorf("failed to decrypt expiry month: %w", err)
				}
				if entity.ExpiryYear, err = crypto.DecryptAESGCM(k, entity.ExpiryYear); err != nil {
					return nil, fmt.Errorf("failed to decrypt expiry year: %w", err)
				}
				if entity.CVV, err = crypto.DecryptAESGCM(k, entity.CVV); err != nil {
					return nil, fmt.Errorf("failed to decrypt CVV: %w", err)
				}
				if entity.Description, err = crypto.DecryptAESGCM(k, entity.Description); err != nil {
					return nil, fmt.Errorf("failed to decrypt description: %w", err)
				}
			}

			return entities, nil
		}
	}
}
