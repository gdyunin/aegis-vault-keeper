package bankcard

import (
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBankCardFromDomain(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testID := uuid.New()
	testUserID := uuid.New()

	tests := []struct {
		input  *bankcard.BankCard
		expect *BankCard
		name   string
	}{
		{
			name:   "nil_input",
			input:  nil,
			expect: nil,
		},
		{
			name: "valid_card",
			input: &bankcard.BankCard{
				ID:          testID,
				UserID:      testUserID,
				CardNumber:  []byte("4532015112830366"),
				CardHolder:  []byte("John Doe"),
				ExpiryMonth: []byte("12"),
				ExpiryYear:  []byte("2025"),
				CVV:         []byte("123"),
				Description: []byte("Test card"),
				UpdatedAt:   testTime,
			},
			expect: &BankCard{
				ID:          testID,
				UserID:      testUserID,
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				Description: "Test card",
				UpdatedAt:   testTime,
			},
		},
		{
			name: "empty_fields",
			input: &bankcard.BankCard{
				ID:          testID,
				UserID:      testUserID,
				CardNumber:  []byte(""),
				CardHolder:  []byte(""),
				ExpiryMonth: []byte(""),
				ExpiryYear:  []byte(""),
				CVV:         []byte(""),
				Description: []byte(""),
				UpdatedAt:   testTime,
			},
			expect: &BankCard{
				ID:          testID,
				UserID:      testUserID,
				CardNumber:  "",
				CardHolder:  "",
				ExpiryMonth: "",
				ExpiryYear:  "",
				CVV:         "",
				Description: "",
				UpdatedAt:   testTime,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := newBankCardFromDomain(tt.input)

			if tt.expect == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expect.ID, result.ID)
				assert.Equal(t, tt.expect.UserID, result.UserID)
				assert.Equal(t, tt.expect.CardNumber, result.CardNumber)
				assert.Equal(t, tt.expect.CardHolder, result.CardHolder)
				assert.Equal(t, tt.expect.ExpiryMonth, result.ExpiryMonth)
				assert.Equal(t, tt.expect.ExpiryYear, result.ExpiryYear)
				assert.Equal(t, tt.expect.CVV, result.CVV)
				assert.Equal(t, tt.expect.Description, result.Description)
				assert.Equal(t, tt.expect.UpdatedAt, result.UpdatedAt)
			}
		})
	}
}

func TestNewBankCardsFromDomain(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testID1 := uuid.New()
	testID2 := uuid.New()
	testUserID := uuid.New()

	tests := []struct {
		name          string
		input         []*bankcard.BankCard
		expectedCount int
	}{
		{
			name:          "nil_input",
			input:         nil,
			expectedCount: 0,
		},
		{
			name:          "empty_slice",
			input:         []*bankcard.BankCard{},
			expectedCount: 0,
		},
		{
			name: "single_card",
			input: []*bankcard.BankCard{
				{
					ID:          testID1,
					UserID:      testUserID,
					CardNumber:  []byte("4532015112830366"),
					CardHolder:  []byte("John Doe"),
					ExpiryMonth: []byte("12"),
					ExpiryYear:  []byte("2025"),
					CVV:         []byte("123"),
					Description: []byte("Test card 1"),
					UpdatedAt:   testTime,
				},
			},
			expectedCount: 1,
		},
		{
			name: "multiple_cards",
			input: []*bankcard.BankCard{
				{
					ID:          testID1,
					UserID:      testUserID,
					CardNumber:  []byte("4532015112830366"),
					CardHolder:  []byte("John Doe"),
					ExpiryMonth: []byte("12"),
					ExpiryYear:  []byte("2025"),
					CVV:         []byte("123"),
					Description: []byte("Test card 1"),
					UpdatedAt:   testTime,
				},
				{
					ID:          testID2,
					UserID:      testUserID,
					CardNumber:  []byte("5555555555554444"),
					CardHolder:  []byte("Jane Smith"),
					ExpiryMonth: []byte("06"),
					ExpiryYear:  []byte("2026"),
					CVV:         []byte("456"),
					Description: []byte("Test card 2"),
					UpdatedAt:   testTime,
				},
			},
			expectedCount: 2,
		},
		{
			name: "cards_with_nil",
			input: []*bankcard.BankCard{
				{
					ID:          testID1,
					UserID:      testUserID,
					CardNumber:  []byte("4532015112830366"),
					CardHolder:  []byte("John Doe"),
					ExpiryMonth: []byte("12"),
					ExpiryYear:  []byte("2025"),
					CVV:         []byte("123"),
					Description: []byte("Test card 1"),
					UpdatedAt:   testTime,
				},
				nil, // This will result in nil in the output
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := newBankCardsFromDomain(tt.input)

			assert.Len(t, result, tt.expectedCount)

			// Verify individual cards
			for i, card := range result {
				if i < len(tt.input) && tt.input[i] != nil {
					require.NotNil(t, card)
					assert.Equal(t, tt.input[i].ID, card.ID)
					assert.Equal(t, tt.input[i].UserID, card.UserID)
					assert.Equal(t, string(tt.input[i].CardNumber), card.CardNumber)
					assert.Equal(t, string(tt.input[i].CardHolder), card.CardHolder)
				} else if i < len(tt.input) && tt.input[i] == nil {
					assert.Nil(t, card)
				}
			}
		})
	}
}

func TestBankCardDTO(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testID := uuid.New()
	testUserID := uuid.New()

	tests := []struct {
		name   string
		card   BankCard
		expect BankCard
	}{
		{
			name: "complete_card",
			card: BankCard{
				ID:          testID,
				UserID:      testUserID,
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				Description: "Test card",
				UpdatedAt:   testTime,
			},
			expect: BankCard{
				ID:          testID,
				UserID:      testUserID,
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				Description: "Test card",
				UpdatedAt:   testTime,
			},
		},
		{
			name: "minimal_card",
			card: BankCard{
				ID:     testID,
				UserID: testUserID,
			},
			expect: BankCard{
				ID:     testID,
				UserID: testUserID,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expect.ID, tt.card.ID)
			assert.Equal(t, tt.expect.UserID, tt.card.UserID)
			assert.Equal(t, tt.expect.CardNumber, tt.card.CardNumber)
			assert.Equal(t, tt.expect.CardHolder, tt.card.CardHolder)
			assert.Equal(t, tt.expect.ExpiryMonth, tt.card.ExpiryMonth)
			assert.Equal(t, tt.expect.ExpiryYear, tt.card.ExpiryYear)
			assert.Equal(t, tt.expect.CVV, tt.card.CVV)
			assert.Equal(t, tt.expect.Description, tt.card.Description)
			assert.Equal(t, tt.expect.UpdatedAt, tt.card.UpdatedAt)
		})
	}
}

func TestPullParams(t *testing.T) {
	t.Parallel()

	testID := uuid.New()
	testUserID := uuid.New()

	params := PullParams{
		ID:     testID,
		UserID: testUserID,
	}

	assert.Equal(t, testID, params.ID)
	assert.Equal(t, testUserID, params.UserID)
}

func TestListParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	params := ListParams{
		UserID: testUserID,
	}

	assert.Equal(t, testUserID, params.UserID)
}

func TestPushParams(t *testing.T) {
	t.Parallel()

	testID := uuid.New()
	testUserID := uuid.New()

	params := PushParams{
		ID:          testID,
		UserID:      testUserID,
		CardNumber:  "4532015112830366",
		CardHolder:  "John Doe",
		ExpiryMonth: "12",
		ExpiryYear:  "2025",
		CVV:         "123",
		Description: "Test card",
	}

	assert.Equal(t, testID, params.ID)
	assert.Equal(t, testUserID, params.UserID)
	assert.Equal(t, "4532015112830366", params.CardNumber)
	assert.Equal(t, "John Doe", params.CardHolder)
	assert.Equal(t, "12", params.ExpiryMonth)
	assert.Equal(t, "2025", params.ExpiryYear)
	assert.Equal(t, "123", params.CVV)
	assert.Equal(t, "Test card", params.Description)
}
