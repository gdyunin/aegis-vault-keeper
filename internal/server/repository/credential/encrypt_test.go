package credential

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
)

// Mock key provider for testing encryption middleware.
type mockEncryptKeyProvider struct {
	key       []byte
	shouldErr bool
}

func (m *mockEncryptKeyProvider) UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	if m.shouldErr {
		return nil, assert.AnError
	}
	return m.key, nil
}

func TestEncryptionMw(t *testing.T) {
	t.Parallel()

	// Valid 32-byte key for AES-256
	validKey := []byte("12345678901234567890123456789012")

	tests := []struct {
		keyProvider *mockEncryptKeyProvider
		entity      *credential.Credential
		name        string
		errorMsg    string
		expectError bool
	}{
		{
			name: "successful encryption",
			keyProvider: &mockEncryptKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entity: &credential.Credential{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Login:       []byte("testuser"),
				Password:    []byte("testpass"),
				Description: []byte("test desc"),
			},
			expectError: false,
		},
		{
			name: "key provider error",
			keyProvider: &mockEncryptKeyProvider{
				shouldErr: true,
			},
			entity: &credential.Credential{
				ID:     uuid.New(),
				UserID: uuid.New(),
			},
			expectError: true,
			errorMsg:    "failed to provide user key",
		},
		{
			name: "encryption error with invalid key",
			keyProvider: &mockEncryptKeyProvider{
				key:       []byte("short"), // Invalid key length
				shouldErr: false,
			},
			entity: &credential.Credential{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Login:       []byte("testuser"),
				Password:    []byte("testpass"),
				Description: []byte("test desc"),
			},
			expectError: true,
			errorMsg:    "failed to encrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create middleware
			mw := encryptionMw(tt.keyProvider)

			// Mock next function
			var nextCalled bool
			var receivedParams SaveParams
			nextFunc := func(ctx context.Context, p SaveParams) error {
				nextCalled = true
				receivedParams = p
				return nil
			}

			// Apply middleware
			wrappedFunc := mw(nextFunc)

			// Execute
			params := SaveParams{Entity: tt.entity}
			err := wrappedFunc(context.Background(), params)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.False(t, nextCalled, "Next function should not be called on error")
			} else {
				require.NoError(t, err)
				assert.True(t, nextCalled, "Next function should be called on success")

				// Verify that fields were encrypted (should be different from original)
				if len(tt.entity.Login) > 0 {
					assert.NotEqual(t, tt.entity.Login, receivedParams.Entity.Login, "Login should be encrypted")
				}
				if len(tt.entity.Password) > 0 {
					assert.NotEqual(t, tt.entity.Password, receivedParams.Entity.Password, "Password should be encrypted")
				}
				if len(tt.entity.Description) > 0 {
					assert.NotEqual(t, tt.entity.Description, receivedParams.Entity.Description, "Description should be encrypted")
				}

				// Verify entity structure is preserved
				assert.Equal(t, tt.entity.ID, receivedParams.Entity.ID, "ID should be preserved")
				assert.Equal(t, tt.entity.UserID, receivedParams.Entity.UserID, "UserID should be preserved")
			}
		})
	}
}

func TestDecryptionMw(t *testing.T) {
	t.Parallel()

	// Valid 32-byte key for AES-256
	validKey := []byte("12345678901234567890123456789012")

	tests := []struct {
		keyProvider *mockEncryptKeyProvider
		name        string
		errorMsg    string
		entities    []*credential.Credential
		expectError bool
	}{
		{
			name: "successful decryption with real encrypted data",
			keyProvider: &mockEncryptKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities: func() []*credential.Credential {
				// Create real encrypted data for testing
				id := uuid.New()
				userID := uuid.New()

				// Use crypto package to create real encrypted values
				loginEncrypted, _ := crypto.EncryptAESGCM(validKey, []byte("test_login"))
				passwordEncrypted, _ := crypto.EncryptAESGCM(validKey, []byte("test_password"))
				descEncrypted, _ := crypto.EncryptAESGCM(validKey, []byte("test_description"))

				return []*credential.Credential{
					{
						ID:          id,
						UserID:      userID,
						Login:       loginEncrypted,
						Password:    passwordEncrypted,
						Description: descEncrypted,
					},
				}
			}(),
			expectError: false,
		},
		{
			name: "empty entities list",
			keyProvider: &mockEncryptKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities:    []*credential.Credential{},
			expectError: false,
		},
		{
			name: "key provider error",
			keyProvider: &mockEncryptKeyProvider{
				shouldErr: true,
			},
			entities: []*credential.Credential{
				{ID: uuid.New(), UserID: uuid.New()},
			},
			expectError: true,
			errorMsg:    "failed to provide user key",
		},
		{
			name: "next function error",
			keyProvider: &mockEncryptKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities:    nil, // Will trigger error in mock next function
			expectError: true,
			errorMsg:    "failed to load entities",
		},
		{
			name: "decryption error with invalid encrypted data",
			keyProvider: &mockEncryptKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities: []*credential.Credential{
				{
					ID:          uuid.New(),
					UserID:      uuid.New(),
					Login:       []byte("invalid-encrypted-data"),
					Password:    []byte("valid-password"),
					Description: []byte("valid-description"),
				},
			},
			expectError: true,
			errorMsg:    "failed to decrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create middleware
			mw := decryptionMw(tt.keyProvider)

			// Mock next function that simulates database load
			nextFunc := func(ctx context.Context, p LoadParams) ([]*credential.Credential, error) {
				if tt.entities == nil {
					return nil, assert.AnError
				}
				return tt.entities, nil
			}

			// Apply middleware
			wrappedFunc := mw(nextFunc)

			// Execute
			params := LoadParams{UserID: uuid.New()}
			result, err := wrappedFunc(context.Background(), params)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)

				if len(tt.entities) == 0 {
					assert.Empty(t, result)
				} else {
					assert.Len(t, result, len(tt.entities))
					// For successful decryption, verify the structure is maintained
					for i, entity := range result {
						assert.Equal(t, tt.entities[i].ID, entity.ID)
						assert.Equal(t, tt.entities[i].UserID, entity.UserID)
					}
				}
			}
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	t.Parallel()

	validKey := []byte("12345678901234567890123456789012")
	keyProvider := &mockEncryptKeyProvider{key: validKey, shouldErr: false}

	t.Run("encryption middleware chains correctly", func(t *testing.T) {
		t.Parallel()

		mw1 := encryptionMw(keyProvider)

		// Second middleware that modifies the entity
		mw2 := func(next saveFunc) saveFunc {
			return func(ctx context.Context, p SaveParams) error {
				// Modify entity to verify chaining
				modifiedLogin := append([]byte("modified_"), p.Entity.Login...)
				p.Entity.Login = modifiedLogin
				return next(ctx, p)
			}
		}

		var finalEntity *credential.Credential
		finalFunc := func(ctx context.Context, p SaveParams) error {
			finalEntity = p.Entity
			return nil
		}

		// Chain middlewares: mw1 -> mw2 -> finalFunc
		chained := mw1(mw2(finalFunc))

		entity := &credential.Credential{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Login:       []byte("testuser"),
			Password:    []byte("testpass"),
			Description: []byte("test desc"),
		}

		err := chained(context.Background(), SaveParams{Entity: entity})
		require.NoError(t, err)

		// Verify that both middlewares were applied
		assert.NotNil(t, finalEntity)
		assert.Contains(t, string(finalEntity.Login), "modified_") // mw2 was applied
		assert.NotEqual(t, "testuser", string(finalEntity.Login))  // mw1 (encryption) was applied
	})
}
