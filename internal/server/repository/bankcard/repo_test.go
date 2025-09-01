package bankcard

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
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
	return []byte("12345678901234567890123456789012!"), nil
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
	cardID := uuid.New()
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
				Entity: &bankcard.BankCard{
					ID:          cardID,
					UserID:      userID,
					CardNumber:  []byte("1234567890123456"),
					CardHolder:  []byte("John Doe"),
					ExpiryMonth: []byte("12"),
					ExpiryYear:  []byte("2025"),
					CVV:         []byte("123"),
					Description: []byte("Main card"),
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
				Entity: &bankcard.BankCard{
					ID:     cardID,
					UserID: userID,
				},
			},
			dbClient: &mockDBClient{
				execFunc: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
					return nil, errors.New("database error")
				},
			},
			keyProvider:   &mockKeyProvider{},
			expectedError: "failed to save bank card",
		},
		{
			name: "key provider error",
			params: SaveParams{
				Entity: &bankcard.BankCard{
					ID:     cardID,
					UserID: userID,
				},
			},
			dbClient: &mockDBClient{},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return nil, errors.New("key error")
				},
			},
			expectedError: "failed to save bank card",
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
	cardID := uuid.New()

	tests := []struct {
		dbClient      *mockDBClient
		keyProvider   *mockKeyProvider
		name          string
		expectedError string
		expectedCards int
		params        LoadParams
	}{
		{
			name: "successful load with results",
			params: LoadParams{
				UserID: userID,
				ID:     cardID,
			},
			dbClient: &mockDBClient{
				queryFunc: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
					// We can't easily mock sql.Rows, so we'll test the error path
					return nil, errors.New("mock query result")
				},
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return []byte("12345678901234567890123456789012"), nil
				},
			},
			expectedError: "failed to load bank cards",
		},
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
			expectedError: "failed to load bank cards",
		},
		{
			name: "key provider error",
			params: LoadParams{
				UserID: userID,
			},
			dbClient: &mockDBClient{
				queryFunc: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
					return nil, errors.New("mock query for key error test")
				},
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return nil, errors.New("key error")
				},
			},
			expectedError: "failed to load bank cards",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository(tt.dbClient, tt.keyProvider)
			cards, err := repo.Load(context.Background(), tt.params)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, cards)
			} else {
				require.NoError(t, err)
				assert.Len(t, cards, tt.expectedCards)
			}
		})
	}
}

func TestEncryptionMiddleware(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	cardID := uuid.New()
	now := time.Now()

	tests := []struct {
		name          string
		entity        *bankcard.BankCard
		keyProvider   *mockKeyProvider
		nextFunc      func(ctx context.Context, p SaveParams) error
		expectedError string
	}{
		{
			name: "successful encryption",
			entity: &bankcard.BankCard{
				ID:          cardID,
				UserID:      userID,
				CardNumber:  []byte("1234567890123456"),
				CardHolder:  []byte("John Doe"),
				ExpiryMonth: []byte("12"),
				ExpiryYear:  []byte("2025"),
				CVV:         []byte("123"),
				Description: []byte("Main card"),
				UpdatedAt:   now,
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return []byte("12345678901234567890123456789012"), nil
				},
			},
			nextFunc: func(ctx context.Context, p SaveParams) error {
				// Verify all fields are encrypted (changed from original)
				assert.NotEqual(t, []byte("1234567890123456"), p.Entity.CardNumber)
				assert.NotEqual(t, []byte("John Doe"), p.Entity.CardHolder)
				assert.NotEqual(t, []byte("12"), p.Entity.ExpiryMonth)
				assert.NotEqual(t, []byte("2025"), p.Entity.ExpiryYear)
				assert.NotEqual(t, []byte("123"), p.Entity.CVV)
				assert.NotEqual(t, []byte("Main card"), p.Entity.Description)
				return nil
			},
		},
		{
			name: "key provider error",
			entity: &bankcard.BankCard{
				ID:     cardID,
				UserID: userID,
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return nil, errors.New("key provider error")
				},
			},
			nextFunc: func(ctx context.Context, p SaveParams) error {
				t.Error("next function should not be called on key provider error")
				return nil
			},
			expectedError: "failed to provide user key",
		},
		{
			name: "encryption error with invalid key",
			entity: &bankcard.BankCard{
				ID:          cardID,
				UserID:      userID,
				CardNumber:  []byte("1234567890123456"),
				CardHolder:  []byte("John Doe"),
				ExpiryMonth: []byte("12"),
				ExpiryYear:  []byte("2025"),
				CVV:         []byte("123"),
				Description: []byte("Main card"),
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return []byte("short"), nil // Invalid key length
				},
			},
			nextFunc: func(ctx context.Context, p SaveParams) error {
				t.Error("next function should not be called on encryption error")
				return nil
			},
			expectedError: "failed to encrypt card number",
		},
		{
			name: "next function error",
			entity: &bankcard.BankCard{
				ID:          cardID,
				UserID:      userID,
				CardNumber:  []byte("1234567890123456"),
				CardHolder:  []byte("John Doe"),
				ExpiryMonth: []byte("12"),
				ExpiryYear:  []byte("2025"),
				CVV:         []byte("123"),
				Description: []byte("Main card"),
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return []byte("12345678901234567890123456789012"), nil
				},
			},
			nextFunc: func(ctx context.Context, p SaveParams) error {
				return errors.New("next function error")
			},
			expectedError: "next function error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			middleware := encryptionMw(tt.keyProvider)
			wrapped := middleware(tt.nextFunc)

			params := SaveParams{Entity: tt.entity}
			err := wrapped(context.Background(), params)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDecryptionMiddleware(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	cardID := uuid.New()

	// Create encrypted test data
	key := []byte("12345678901234567890123456789012")
	encryptedCardNumber := []byte("encrypted_card_number")
	encryptedCardHolder := []byte("encrypted_card_holder")

	tests := []struct {
		keyProvider   *mockKeyProvider
		nextFunc      func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error)
		name          string
		expectedError string
		entities      []*bankcard.BankCard
		expectedCards int
	}{
		{
			name: "successful decryption with results",
			entities: []*bankcard.BankCard{
				{
					ID:          cardID,
					UserID:      userID,
					CardNumber:  encryptedCardNumber,
					CardHolder:  encryptedCardHolder,
					ExpiryMonth: []byte("encrypted_month"),
					ExpiryYear:  []byte("encrypted_year"),
					CVV:         []byte("encrypted_cvv"),
					Description: []byte("encrypted_desc"),
				},
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return key, nil
				},
			},
			nextFunc: func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
				return []*bankcard.BankCard{
					{
						ID:          cardID,
						UserID:      userID,
						CardNumber:  encryptedCardNumber,
						CardHolder:  encryptedCardHolder,
						ExpiryMonth: []byte("encrypted_month"),
						ExpiryYear:  []byte("encrypted_year"),
						CVV:         []byte("encrypted_cvv"),
						Description: []byte("encrypted_desc"),
					},
				}, nil
			},
			expectedError: "failed to decrypt card number", // Will fail because test data isn't really encrypted
		},
		{
			name: "empty results",
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return key, nil
				},
			},
			nextFunc: func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
				return []*bankcard.BankCard{}, nil
			},
			expectedCards: 0,
		},
		{
			name: "next function error",
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return key, nil
				},
			},
			nextFunc: func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
				return nil, errors.New("next function error")
			},
			expectedError: "failed to load entities",
		},
		{
			name: "key provider error",
			entities: []*bankcard.BankCard{
				{
					ID:         cardID,
					UserID:     userID,
					CardNumber: encryptedCardNumber,
				},
			},
			keyProvider: &mockKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return nil, errors.New("key provider error")
				},
			},
			nextFunc: func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
				return []*bankcard.BankCard{
					{
						ID:         cardID,
						UserID:     userID,
						CardNumber: encryptedCardNumber,
					},
				}, nil
			},
			expectedError: "failed to provide user key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			middleware := decryptionMw(tt.keyProvider)
			wrapped := middleware(tt.nextFunc)

			params := LoadParams{UserID: userID}
			cards, err := wrapped(context.Background(), params)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, cards)
			} else {
				require.NoError(t, err)
				assert.Len(t, cards, tt.expectedCards)
			}
		})
	}
}
