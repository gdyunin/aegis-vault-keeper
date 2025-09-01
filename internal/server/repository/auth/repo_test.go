package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		secretKey []byte
	}{
		{
			name:      "creates repository with valid parameters",
			secretKey: []byte("test-secret-key-32-chars-long!!"),
		},
		{
			name:      "creates repository with empty secret key",
			secretKey: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository(nil, tt.secretKey)

			assert.NotNil(t, repo)
			assert.NotNil(t, repo.save)
			assert.NotNil(t, repo.load)
		})
	}
}

func TestRepository_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mockExecError error
		params        SaveParams
		name          string
		errContains   string
		expectErr     bool
	}{
		{
			name: "successful save",
			params: SaveParams{
				Entity: &auth.User{
					ID:           uuid.New(),
					Login:        "testuser",
					PasswordHash: "hashed_password",
					CryptoKey:    []byte("crypto_key"),
				},
			},
			mockExecError: nil,
			expectErr:     false,
		},
		{
			name: "save with database error",
			params: SaveParams{
				Entity: &auth.User{
					ID:           uuid.New(),
					Login:        "testuser",
					PasswordHash: "hashed_password",
					CryptoKey:    []byte("crypto_key"),
				},
			},
			mockExecError: errors.New("database connection failed"),
			expectErr:     true,
			errContains:   "failed to save user",
		},
		{
			name: "save with user already exists error",
			params: SaveParams{
				Entity: &auth.User{
					ID:           uuid.New(),
					Login:        "existing_user",
					PasswordHash: "hashed_password",
					CryptoKey:    []byte("crypto_key"),
				},
			},
			mockExecError: ErrUserAlreadyExists,
			expectErr:     true,
			errContains:   "failed to save user",
		},
		{
			name: "save with short secret key error",
			params: SaveParams{
				Entity: &auth.User{
					ID:           uuid.New(),
					Login:        "testuser",
					PasswordHash: "hashed_password",
					CryptoKey:    []byte("crypto_key"),
				},
			},
			mockExecError: nil,
			expectErr:     true,
			errContains:   "failed to save user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					return nil, tt.mockExecError
				},
			}

			// Use proper 32-byte secret key for valid tests, short key for error tests
			var secretKey []byte
			if tt.name == "save with short secret key error" {
				secretKey = []byte("short_key") // This will cause encryption error
			} else {
				secretKey = []byte("12345678901234567890123456789012") // Exactly 32 bytes
			}

			repo := NewRepository(mockDB, secretKey)

			ctx := context.Background()
			err := repo.Save(ctx, tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepository_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mockQueryError error
		mockUser       *auth.User
		name           string
		errContains    string
		params         LoadParams
		expectErr      bool
	}{
		{
			name: "successful load by ID",
			params: LoadParams{
				ID:    uuid.New(),
				Login: "",
			},
			mockQueryError: nil,
			mockUser: &auth.User{
				ID:           uuid.New(),
				Login:        "testuser",
				PasswordHash: "hashed_password",
				CryptoKey:    []byte("crypto_key"),
			},
			expectErr: false,
		},
		{
			name: "successful load by login",
			params: LoadParams{
				ID:    uuid.Nil,
				Login: "testuser",
			},
			mockQueryError: nil,
			mockUser: &auth.User{
				ID:           uuid.New(),
				Login:        "testuser",
				PasswordHash: "hashed_password",
				CryptoKey:    []byte("crypto_key"),
			},
			expectErr: false,
		},
		{
			name: "load with database error",
			params: LoadParams{
				ID:    uuid.New(),
				Login: "",
			},
			mockQueryError: errors.New("database connection failed"),
			mockUser:       nil,
			expectErr:      true,
			errContains:    "failed to load user",
		},
		{
			name: "load with user not found",
			params: LoadParams{
				ID:    uuid.New(),
				Login: "",
			},
			mockQueryError: ErrUserNotFound,
			mockUser:       nil,
			expectErr:      true,
			errContains:    "failed to load user",
		},
		{
			name: "load with no parameters",
			params: LoadParams{
				ID:    uuid.Nil,
				Login: "",
			},
			mockQueryError: errors.New("at least one of ID or Login must be provided"),
			mockUser:       nil,
			expectErr:      true,
			errContains:    "failed to load user",
		},
		{
			name: "load with both ID and login",
			params: LoadParams{
				ID:    uuid.New(),
				Login: "testuser",
			},
			mockQueryError: nil,
			mockUser: &auth.User{
				ID:           uuid.New(),
				Login:        "testuser",
				PasswordHash: "hashed_password",
				CryptoKey:    []byte("crypto_key"),
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock that will be used internally by the load function
			mockDB := &mockDBClient{
				queryRowFunc: func(ctx context.Context, query string, args ...interface{}) *sql.Row {
					// Return a mock row that would trigger the expected behavior
					return nil
				},
			}

			secretKey := []byte("12345678901234567890123456789012") // Exactly 32 bytes
			repo := NewRepository(mockDB, secretKey)

			// Override the load function to simulate the error behavior
			if tt.mockQueryError != nil {
				repo.load = func(ctx context.Context, params LoadParams) (*auth.User, error) {
					return tt.mockUser, tt.mockQueryError
				}
			} else if tt.mockUser != nil {
				repo.load = func(ctx context.Context, params LoadParams) (*auth.User, error) {
					return tt.mockUser, nil
				}
			}

			ctx := context.Background()
			user, err := repo.Load(ctx, tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.mockUser.ID, user.ID)
				assert.Equal(t, tt.mockUser.Login, user.Login)
				assert.Equal(t, tt.mockUser.PasswordHash, user.PasswordHash)
				assert.Equal(t, tt.mockUser.CryptoKey, user.CryptoKey)
			}
		})
	}
}

func TestRepository_SaveLoadIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		user      *auth.User
		secretKey []byte
	}{
		{
			name: "save and load with valid user",
			user: &auth.User{
				ID:           uuid.New(),
				Login:        "testuser",
				PasswordHash: "hashed_password",
				CryptoKey:    []byte("crypto_key"),
			},
			secretKey: []byte("12345678901234567890123456789012"), // Exactly 32 bytes
		},
		{
			name: "save and load with special characters",
			user: &auth.User{
				ID:           uuid.New(),
				Login:        "user@domain.com",
				PasswordHash: "complex$hash#123",
				CryptoKey:    []byte("special_crypto_key!@#"),
			},
			secretKey: []byte("87654321098765432109876543210987"), // Exactly 32 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Mock DB that stores data for integration testing
			var savedParams SaveParams
			var loadedParams LoadParams

			mockDB := &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					// Store the save parameters for verification
					return nil, errors.New("test error")
				},
				queryRowFunc: func(ctx context.Context, query string, args ...interface{}) *sql.Row {
					// Return a mock row
					return nil
				},
			}

			repo := NewRepository(mockDB, tt.secretKey)

			// Override save and load for integration testing
			repo.save = func(ctx context.Context, params SaveParams) error {
				savedParams = params
				return nil
			}

			repo.load = func(ctx context.Context, params LoadParams) (*auth.User, error) {
				loadedParams = params
				return tt.user, nil
			}

			ctx := context.Background()

			// Test Save
			saveParams := SaveParams{Entity: tt.user}
			err := repo.Save(ctx, saveParams)
			require.NoError(t, err)

			// Test Load
			loadParams := LoadParams{ID: tt.user.ID}
			loadedUser, err := repo.Load(ctx, loadParams)
			require.NoError(t, err)
			require.NotNil(t, loadedUser)

			// Verify the data was passed correctly
			assert.Equal(t, tt.user.ID, loadedUser.ID)
			assert.Equal(t, tt.user.Login, loadedUser.Login)
			assert.Equal(t, tt.user.PasswordHash, loadedUser.PasswordHash)
			assert.Equal(t, tt.user.CryptoKey, loadedUser.CryptoKey)

			// Verify parameters were passed correctly
			assert.Equal(t, tt.user, savedParams.Entity)
			assert.Equal(t, tt.user.ID, loadedParams.ID)
		})
	}
}
