package filedata

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
)

// Mock key provider for testing encryption middleware.
type mockFileDataKeyProvider struct {
	key       []byte
	shouldErr bool
}

func (m *mockFileDataKeyProvider) UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error) {
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
		keyProvider *mockFileDataKeyProvider
		entity      *filedata.FileData
		name        string
		errorMsg    string
		expectError bool
	}{
		{
			name: "successful encryption",
			keyProvider: &mockFileDataKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entity: &filedata.FileData{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				StorageKey:  []byte("storage/path/file.txt"),
				HashSum:     []byte("sha256hashvalue"),
				Description: []byte("test file description"),
			},
			expectError: false,
		},
		{
			name: "key provider error",
			keyProvider: &mockFileDataKeyProvider{
				shouldErr: true,
			},
			entity: &filedata.FileData{
				ID:     uuid.New(),
				UserID: uuid.New(),
			},
			expectError: true,
			errorMsg:    "failed to provide user key",
		},
		{
			name: "encryption error with invalid key",
			keyProvider: &mockFileDataKeyProvider{
				key:       []byte("short"), // Invalid key length
				shouldErr: false,
			},
			entity: &filedata.FileData{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				StorageKey:  []byte("storage/path/file.txt"),
				HashSum:     []byte("sha256hashvalue"),
				Description: []byte("test file description"),
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
				if len(tt.entity.StorageKey) > 0 {
					assert.NotEqual(t, tt.entity.StorageKey, receivedParams.Entity.StorageKey, "StorageKey should be encrypted")
				}
				if len(tt.entity.HashSum) > 0 {
					assert.NotEqual(t, tt.entity.HashSum, receivedParams.Entity.HashSum, "HashSum should be encrypted")
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
		keyProvider *mockFileDataKeyProvider
		name        string
		errorMsg    string
		entities    []*filedata.FileData
		expectError bool
	}{
		{
			name: "successful decryption with real encrypted data",
			keyProvider: &mockFileDataKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities: func() []*filedata.FileData {
				// Create real encrypted data for testing
				id := uuid.New()
				userID := uuid.New()

				// Use crypto package to create real encrypted values
				storageKeyEncrypted, _ := crypto.EncryptAESGCM(validKey, []byte("storage/path/file.txt"))
				hashSumEncrypted, _ := crypto.EncryptAESGCM(validKey, []byte("sha256hashvalue"))
				descEncrypted, _ := crypto.EncryptAESGCM(validKey, []byte("test file description"))

				return []*filedata.FileData{
					{
						ID:          id,
						UserID:      userID,
						StorageKey:  storageKeyEncrypted,
						HashSum:     hashSumEncrypted,
						Description: descEncrypted,
					},
				}
			}(),
			expectError: false,
		},
		{
			name: "empty entities list",
			keyProvider: &mockFileDataKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities:    []*filedata.FileData{},
			expectError: false,
		},
		{
			name: "key provider error",
			keyProvider: &mockFileDataKeyProvider{
				shouldErr: true,
			},
			entities: []*filedata.FileData{
				{ID: uuid.New(), UserID: uuid.New()},
			},
			expectError: true,
			errorMsg:    "failed to provide user key",
		},
		{
			name: "next function error",
			keyProvider: &mockFileDataKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities:    nil, // Will trigger error in mock next function
			expectError: true,
			errorMsg:    "failed to load entities",
		},
		{
			name: "decryption error with invalid encrypted data",
			keyProvider: &mockFileDataKeyProvider{
				key:       validKey,
				shouldErr: false,
			},
			entities: []*filedata.FileData{
				{
					ID:          uuid.New(),
					UserID:      uuid.New(),
					StorageKey:  []byte("invalid-encrypted-data"),
					HashSum:     []byte("valid-hash"),
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
			nextFunc := func(ctx context.Context, p LoadParams) ([]*filedata.FileData, error) {
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
	keyProvider := &mockFileDataKeyProvider{key: validKey, shouldErr: false}

	t.Run("encryption middleware chains correctly", func(t *testing.T) {
		t.Parallel()

		mw1 := encryptionMw(keyProvider)

		// Second middleware that modifies the entity
		mw2 := func(next saveFunc) saveFunc {
			return func(ctx context.Context, p SaveParams) error {
				// Modify entity to verify chaining
				modifiedKey := append([]byte("modified_"), p.Entity.StorageKey...)
				p.Entity.StorageKey = modifiedKey
				return next(ctx, p)
			}
		}

		var finalEntity *filedata.FileData
		finalFunc := func(ctx context.Context, p SaveParams) error {
			finalEntity = p.Entity
			return nil
		}

		// Chain middlewares: mw1 -> mw2 -> finalFunc
		chained := mw1(mw2(finalFunc))

		entity := &filedata.FileData{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			StorageKey:  []byte("storage/path/file.txt"),
			HashSum:     []byte("sha256hashvalue"),
			Description: []byte("test file description"),
		}

		err := chained(context.Background(), SaveParams{Entity: entity})
		require.NoError(t, err)

		// Verify that both middlewares were applied
		assert.NotNil(t, finalEntity)
		assert.Contains(t, string(finalEntity.StorageKey), "modified_") // mw2 was applied
		assert.NotEqual(
			t,
			"storage/path/file.txt",
			string(finalEntity.StorageKey),
		) // mw1 (encryption) was applied
	})
}
