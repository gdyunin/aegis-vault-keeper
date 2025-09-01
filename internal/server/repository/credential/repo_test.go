package credential

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDBClient implements db.DBClient for testing.
type mockDBClient struct {
	execFunc       func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	queryFunc      func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	queryRowFunc   func(ctx context.Context, query string, args ...interface{}) *sql.Row
	beginTxFunc    func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	commitTxFunc   func(tx *sql.Tx) error
	rollbackTxFunc func(tx *sql.Tx) error
}

func (m *mockDBClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, query, args...)
	}
	return mockResult{}, nil
}

func (m *mockDBClient) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, args...)
	}
	return nil, errors.New("mock not configured")
}

func (m *mockDBClient) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, query, args...)
	}
	return nil
}

func (m *mockDBClient) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.beginTxFunc != nil {
		return m.beginTxFunc(ctx, opts)
	}
	return nil, errors.New("mock not configured")
}

func (m *mockDBClient) CommitTx(tx *sql.Tx) error {
	if m.commitTxFunc != nil {
		return m.commitTxFunc(tx)
	}
	return nil
}

func (m *mockDBClient) RollbackTx(tx *sql.Tx) error {
	if m.rollbackTxFunc != nil {
		return m.rollbackTxFunc(tx)
	}
	return nil
}

// mockResult implements sql.Result for testing.
type mockResult struct{}

func (m mockResult) LastInsertId() (int64, error) { return 1, nil }
func (m mockResult) RowsAffected() (int64, error) { return 1, nil }

// mockKeyProvider implements keyprv.UserKeyProvider for testing.
type mockKeyProvider struct {
	keyFunc func(ctx context.Context, userID uuid.UUID) ([]byte, error)
}

func (m *mockKeyProvider) UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	if m.keyFunc != nil {
		return m.keyFunc(ctx, userID)
	}
	return []byte("12345678901234567890123456789012"), nil
}

func TestNewRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates repository with nil parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository(nil, nil)

			assert.NotNil(t, repo)
			assert.NotNil(t, repo.save)
			assert.NotNil(t, repo.load)
		})
	}
}

func TestRepository_Save(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	credID := uuid.New()
	now := time.Now()

	tests := []struct {
		name          string
		params        SaveParams
		dbClient      *mockDBClient
		keyProvider   *mockKeyProvider
		expectedError string
	}{
		{
			name: "successful save",
			params: SaveParams{
				Entity: &credential.Credential{
					ID:          credID,
					UserID:      userID,
					Login:       []byte("testuser"),
					Password:    []byte("password123"),
					Description: []byte("Test credentials"),
					UpdatedAt:   now,
				},
			},
			dbClient: &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					return mockResult{}, nil
				},
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return []byte("12345678901234567890123456789012"), nil
				},
			},
		},
		{
			name: "database error",
			params: SaveParams{
				Entity: &credential.Credential{
					ID:     credID,
					UserID: userID,
				},
			},
			dbClient: &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					return nil, errors.New("database error")
				},
			},
			keyProvider:   &mockKeyProvider{},
			expectedError: "failed to save credentials",
		},
		{
			name: "key provider error",
			params: SaveParams{
				Entity: &credential.Credential{
					ID:     credID,
					UserID: userID,
				},
			},
			dbClient: &mockDBClient{},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return nil, errors.New("key error")
				},
			},
			expectedError: "failed to save credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository(tt.dbClient, tt.keyProvider)
			err := repo.Save(context.Background(), tt.params)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepository_Load(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	credID := uuid.New()

	tests := []struct {
		dbClient      *mockDBClient
		keyProvider   *mockKeyProvider
		name          string
		expectedError string
		expectedCreds int
		params        LoadParams
	}{
		{
			name: "database error",
			params: LoadParams{
				UserID: userID,
			},
			dbClient: &mockDBClient{
				queryFunc: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
					return nil, errors.New("database error")
				},
			},
			keyProvider:   &mockKeyProvider{},
			expectedError: "failed to load credentials",
		},
		{
			name: "key provider error",
			params: LoadParams{
				UserID: userID,
				ID:     credID,
			},
			dbClient: &mockDBClient{
				queryFunc: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
					return nil, errors.New("query error")
				},
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return nil, errors.New("key error")
				},
			},
			expectedError: "failed to load credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository(tt.dbClient, tt.keyProvider)
			creds, err := repo.Load(context.Background(), tt.params)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, creds)
			} else {
				require.NoError(t, err)
				assert.Len(t, creds, tt.expectedCreds)
			}
		})
	}
}
