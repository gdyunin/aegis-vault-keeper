package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResult implements sql.Result for testing.
type mockResult struct{}

func (m mockResult) LastInsertId() (int64, error) { return 1, nil }
func (m mockResult) RowsAffected() (int64, error) { return 1, nil }

// Mock DB client for testing raw functions.
type mockDBClient struct {
	execFunc     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	queryRowFunc func(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (m *mockDBClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, query, args...)
	}
	return mockResult{}, nil
}

func (m *mockDBClient) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, query, args...)
	}
	return nil
}

func (m *mockDBClient) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mock not configured")
}

func (m *mockDBClient) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, errors.New("mock not configured")
}

func (m *mockDBClient) CommitTx(tx *sql.Tx) error {
	return nil
}

func (m *mockDBClient) RollbackTx(tx *sql.Tx) error {
	return nil
}

// Ensure mockDBClient implements db.DBClient.
var _ db.DBClient = (*mockDBClient)(nil)

func TestRawSave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		execError   error
		expectedErr error
		params      SaveParams
		name        string
		expectErr   bool
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
			execError:   nil,
			expectErr:   false,
			expectedErr: nil,
		},
		{
			name: "save with unique constraint violation",
			params: SaveParams{
				Entity: &auth.User{
					ID:           uuid.New(),
					Login:        "existing_user",
					PasswordHash: "hashed_password",
					CryptoKey:    []byte("crypto_key"),
				},
			},
			execError:   &pgconn.PgError{Code: "23505"},
			expectErr:   true,
			expectedErr: ErrUserAlreadyExists,
		},
		{
			name: "save with generic database error",
			params: SaveParams{
				Entity: &auth.User{
					ID:           uuid.New(),
					Login:        "testuser",
					PasswordHash: "hashed_password",
					CryptoKey:    []byte("crypto_key"),
				},
			},
			execError: errors.New("database connection failed"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					// Verify the query structure
					assert.Contains(t, query, "INSERT INTO aegis_vault_keeper.auth_users")
					assert.Contains(t, query, "ON CONFLICT (id) DO UPDATE SET")

					// Verify parameters
					require.Len(t, args, 4)
					assert.Equal(t, tt.params.Entity.ID, args[0])
					assert.Equal(t, tt.params.Entity.Login, args[1])
					assert.Equal(t, tt.params.Entity.PasswordHash, args[2])
					assert.Equal(t, tt.params.Entity.CryptoKey, args[3])

					return nil, tt.execError
				},
			}

			saveFunc := rawSave(mockDB)
			err := saveFunc(context.Background(), tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRawLoadParameterValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr string
		params      LoadParams
		expectErr   bool
	}{
		{
			name: "load with no parameters should error",
			params: LoadParams{
				ID:    uuid.Nil,
				Login: "",
			},
			expectErr:   true,
			expectedErr: "at least one of ID or Login must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := &mockDBClient{}

			loadFunc := rawLoad(mockDB)
			_, err := loadFunc(context.Background(), tt.params)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestRawFunctionsIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "raw functions return valid function types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := &mockDBClient{}

			// Test that functions are created successfully
			saveFunc := rawSave(mockDB)
			loadFunc := rawLoad(mockDB)

			assert.NotNil(t, saveFunc, "rawSave should return a valid function")
			assert.NotNil(t, loadFunc, "rawLoad should return a valid function")
		})
	}
}

func TestRawSaveQueryConstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "save query has correct structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			queryCalled := false
			mockDB := &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					queryCalled = true

					// Verify query components
					assert.Contains(t, query, "INSERT INTO aegis_vault_keeper.auth_users")
					assert.Contains(t, query, "(id, login, password_hash, crypto_key)")
					assert.Contains(t, query, "VALUES ($1, $2, $3, $4)")
					assert.Contains(t, query, "ON CONFLICT (id) DO UPDATE SET")
					assert.Contains(t, query, "login = EXCLUDED.login")
					assert.Contains(t, query, "password_hash = EXCLUDED.password_hash")
					assert.Contains(t, query, "crypto_key = EXCLUDED.crypto_key")

					return mockResult{}, nil
				},
			}

			user := &auth.User{
				ID:           uuid.New(),
				Login:        "testuser",
				PasswordHash: "hash",
				CryptoKey:    []byte("key"),
			}

			saveFunc := rawSave(mockDB)
			err := saveFunc(context.Background(), SaveParams{Entity: user})

			assert.NoError(t, err)
			assert.True(t, queryCalled, "Exec should have been called")
		})
	}
}
