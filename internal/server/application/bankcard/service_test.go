package bankcard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/bankcard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository for testing.
type mockRepository struct {
	saveFunc func(ctx context.Context, params repository.SaveParams) error
	loadFunc func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error)
}

func (m *mockRepository) Save(ctx context.Context, params repository.SaveParams) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, params)
	}
	return nil
}

func (m *mockRepository) Load(
	ctx context.Context,
	params repository.LoadParams,
) ([]*bankcard.BankCard, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx, params)
	}
	return nil, nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{}
	service := NewService(repo)

	require.NotNil(t, service)
	assert.Equal(t, repo, service.r)
}

func TestService_Pull(t *testing.T) {
	t.Parallel()

	testCardID := uuid.New()
	testUserID := uuid.New()
	testTime := time.Now()

	testCard := &bankcard.BankCard{
		ID:          testCardID,
		UserID:      testUserID,
		CardNumber:  []byte("4532015112830366"), // Valid Luhn test card
		CardHolder:  []byte("John Doe"),
		ExpiryMonth: []byte("12"),
		ExpiryYear:  []byte("2025"),
		CVV:         []byte("123"),
		Description: []byte("Test card"),
		UpdatedAt:   testTime,
	}

	type args struct {
		params PullParams
	}
	tests := []struct {
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		args           args
		wantErr        bool
		expectCard     bool
	}{
		{
			name: "successful_pull",
			args: args{
				params: PullParams{
					ID:     testCardID,
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{testCard}, nil
				}
			},
			expectCard: true,
			wantErr:    false,
		},
		{
			name: "card_not_found",
			args: args{
				params: PullParams{
					ID:     testCardID,
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{}, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "bank card not found",
		},
		{
			name: "repository_load_failed",
			args: args{
				params: PullParams{
					ID:     testCardID,
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to load bank cards",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			card, err := service.Pull(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Nil(t, card)
			} else {
				require.NoError(t, err)
				if tt.expectCard {
					require.NotNil(t, card)
					assert.Equal(t, testCardID, card.ID)
					assert.Equal(t, testUserID, card.UserID)
					assert.Equal(t, "4532015112830366", card.CardNumber)
					assert.Equal(t, "John Doe", card.CardHolder)
				}
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testTime := time.Now()

	testCards := []*bankcard.BankCard{
		{
			ID:          uuid.New(),
			UserID:      testUserID,
			CardNumber:  []byte("4532015112830366"), // Valid Luhn test card
			CardHolder:  []byte("John Doe"),
			ExpiryMonth: []byte("12"),
			ExpiryYear:  []byte("2025"),
			CVV:         []byte("123"),
			Description: []byte("Test card 1"),
			UpdatedAt:   testTime,
		},
		{
			ID:          uuid.New(),
			UserID:      testUserID,
			CardNumber:  []byte("5555555555554444"), // Valid MasterCard test card
			CardHolder:  []byte("Jane Smith"),
			ExpiryMonth: []byte("06"),
			ExpiryYear:  []byte("2026"),
			CVV:         []byte("456"),
			Description: []byte("Test card 2"),
			UpdatedAt:   testTime,
		},
	}

	type args struct {
		params ListParams
	}
	tests := []struct {
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		expectedCount  int
		args           args
		wantErr        bool
	}{
		{
			name: "successful_list",
			args: args{
				params: ListParams{
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return testCards, nil
				}
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "empty_list",
			args: args{
				params: ListParams{
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{}, nil
				}
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "repository_load_failed",
			args: args{
				params: ListParams{
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to load bank cards",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			cards, err := service.List(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Nil(t, cards)
			} else {
				require.NoError(t, err)
				assert.Len(t, cards, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, testUserID, cards[0].UserID)
				}
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testCardID := uuid.New()

	type args struct {
		params *PushParams
	}
	tests := []struct {
		args           args
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		wantErr        bool
		expectID       bool
	}{
		{
			name: "successful_create",
			args: args{
				params: &PushParams{
					UserID:      testUserID,
					CardNumber:  "4532015112830366", // Valid Luhn test card
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CVV:         "123",
					Description: "Test card",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return nil
				}
			},
			expectID: true,
			wantErr:  false,
		},
		{
			name: "successful_update",
			args: args{
				params: &PushParams{
					ID:          testCardID,
					UserID:      testUserID,
					CardNumber:  "4532015112830366", // Valid Luhn test card
					CardHolder:  "John Doe Updated",
					ExpiryMonth: "12",
					ExpiryYear:  "2026",
					CVV:         "123",
					Description: "Updated test card",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{
						{
							ID:     testCardID,
							UserID: testUserID,
						},
					}, nil
				}
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return nil
				}
			},
			expectID: true,
			wantErr:  false,
		},
		{
			name: "invalid_card_data",
			args: args{
				params: &PushParams{
					UserID:      testUserID,
					CardNumber:  "invalid",
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CVV:         "123",
					Description: "Test card",
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to create bank card",
		},
		{
			name: "repository_save_failed",
			args: args{
				params: &PushParams{
					UserID:      testUserID,
					CardNumber:  "4532015112830366", // Valid Luhn test card
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CVV:         "123",
					Description: "Test card",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to save bank card",
		},
		{
			name: "update_access_denied",
			args: args{
				params: &PushParams{
					ID:          testCardID,
					UserID:      testUserID,
					CardNumber:  "4532015112830366", // Valid Luhn test card
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CVV:         "123",
					Description: "Test card",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{}, nil // No card found
				}
			},
			wantErr:        true,
			expectedErrMsg: "access check for updating bank card failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			cardID, err := service.Push(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Equal(t, uuid.Nil, cardID)
			} else {
				require.NoError(t, err)
				if tt.expectID {
					assert.NotEqual(t, uuid.Nil, cardID)
				}
			}
		})
	}
}

func TestService_checkAccessToUpdate(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testCardID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		cardID         uuid.UUID
		userID         uuid.UUID
		wantErr        bool
	}{
		{
			name:   "access_granted",
			cardID: testCardID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{
						{
							ID:     testCardID,
							UserID: testUserID,
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:   "card_not_found",
			cardID: testCardID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{}, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "bank card for update not found",
		},
		{
			name:   "access_denied_different_user",
			cardID: testCardID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return []*bankcard.BankCard{
						{
							ID:     testCardID,
							UserID: otherUserID, // Different user
						},
					}, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "access denied to bank card",
		},
		{
			name:   "repository_error",
			cardID: testCardID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to pull existing bank card",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			err := service.checkAccessToUpdate(context.Background(), tt.cardID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
