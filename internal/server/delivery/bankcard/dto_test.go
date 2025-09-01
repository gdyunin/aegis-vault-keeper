package bankcard

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankCard_ToApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dto       *BankCard
		expected  *bankcard.BankCard
		name      string
		userID    uuid.UUID
		expectNil bool
	}{
		{
			name: "valid bank card conversion",
			dto: &BankCard{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "25",
				CVV:         "123",
				Description: "Primary card",
				UpdatedAt:   time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
			},
			userID: uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
			expected: &bankcard.BankCard{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				UserID:      uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "25",
				CVV:         "123",
				Description: "Primary card",
				UpdatedAt:   time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "nil bank card",
			dto:       nil,
			userID:    uuid.New(),
			expectNil: true,
		},
		{
			name: "empty fields",
			dto: &BankCard{
				ID:          uuid.Nil,
				CardNumber:  "",
				CardHolder:  "",
				ExpiryMonth: "",
				ExpiryYear:  "",
				CVV:         "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
			userID: uuid.New(),
			expected: &bankcard.BankCard{
				ID:          uuid.Nil,
				UserID:      uuid.New(),
				CardNumber:  "",
				CardHolder:  "",
				ExpiryMonth: "",
				ExpiryYear:  "",
				CVV:         "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToApp(tt.userID)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.userID, result.UserID) // Always use provided userID
			assert.Equal(t, tt.expected.CardNumber, result.CardNumber)
			assert.Equal(t, tt.expected.CardHolder, result.CardHolder)
			assert.Equal(t, tt.expected.ExpiryMonth, result.ExpiryMonth)
			assert.Equal(t, tt.expected.ExpiryYear, result.ExpiryYear)
			assert.Equal(t, tt.expected.CVV, result.CVV)
			assert.Equal(t, tt.expected.Description, result.Description)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
		})
	}
}

func TestBankCardsToApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dtos      []*BankCard
		expected  []*bankcard.BankCard
		userID    uuid.UUID
		expectNil bool
	}{
		{
			name: "valid bank cards conversion",
			dtos: []*BankCard{
				{
					ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					CardNumber:  "4532015112830366",
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVV:         "123",
					Description: "Primary card",
				},
				{
					ID:          uuid.MustParse("789e0123-e89b-12d3-a456-426614174000"),
					CardNumber:  "5555555555554444",
					CardHolder:  "Jane Smith",
					ExpiryMonth: "06",
					ExpiryYear:  "26",
					CVV:         "456",
					Description: "Secondary card",
				},
			},
			userID: uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
			expected: []*bankcard.BankCard{
				{
					ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					UserID:      uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
					CardNumber:  "4532015112830366",
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVV:         "123",
					Description: "Primary card",
				},
				{
					ID:          uuid.MustParse("789e0123-e89b-12d3-a456-426614174000"),
					UserID:      uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
					CardNumber:  "5555555555554444",
					CardHolder:  "Jane Smith",
					ExpiryMonth: "06",
					ExpiryYear:  "26",
					CVV:         "456",
					Description: "Secondary card",
				},
			},
		},
		{
			name:      "nil slice",
			dtos:      nil,
			userID:    uuid.New(),
			expectNil: true,
		},
		{
			name:     "empty slice",
			dtos:     []*BankCard{},
			userID:   uuid.New(),
			expected: []*bankcard.BankCard{},
		},
		{
			name: "slice with nil elements",
			dtos: []*BankCard{
				{
					ID:         uuid.New(),
					CardNumber: "4532015112830366",
				},
				nil,
				{
					ID:         uuid.New(),
					CardNumber: "5555555555554444",
				},
			},
			userID: uuid.New(),
			expected: []*bankcard.BankCard{
				{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					CardNumber: "4532015112830366",
				},
				nil,
				{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					CardNumber: "5555555555554444",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := BankCardsToApp(tt.dtos, tt.userID)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Len(t, result, len(tt.expected))

			for i, expectedCard := range tt.expected {
				if expectedCard == nil {
					assert.Nil(t, result[i])
					continue
				}

				require.NotNil(t, result[i])
				assert.Equal(t, tt.userID, result[i].UserID)
				assert.Equal(t, expectedCard.CardNumber, result[i].CardNumber)
			}
		})
	}
}

func TestNewBankCardFromApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		appCard   *bankcard.BankCard
		expected  *BankCard
		name      string
		expectNil bool
	}{
		{
			name: "valid app card conversion",
			appCard: &bankcard.BankCard{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				UserID:      uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "25",
				CVV:         "123",
				Description: "Primary card",
				UpdatedAt:   time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
			},
			expected: &BankCard{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				CardNumber:  "4532015112830366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "25",
				CVV:         "123",
				Description: "Primary card",
				UpdatedAt:   time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "nil app card",
			appCard:   nil,
			expectNil: true,
		},
		{
			name: "empty fields",
			appCard: &bankcard.BankCard{
				ID:          uuid.Nil,
				UserID:      uuid.Nil,
				CardNumber:  "",
				CardHolder:  "",
				ExpiryMonth: "",
				ExpiryYear:  "",
				CVV:         "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
			expected: &BankCard{
				ID:          uuid.Nil,
				CardNumber:  "",
				CardHolder:  "",
				ExpiryMonth: "",
				ExpiryYear:  "",
				CVV:         "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewBankCardFromApp(tt.appCard)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.CardNumber, result.CardNumber)
			assert.Equal(t, tt.expected.CardHolder, result.CardHolder)
			assert.Equal(t, tt.expected.ExpiryMonth, result.ExpiryMonth)
			assert.Equal(t, tt.expected.ExpiryYear, result.ExpiryYear)
			assert.Equal(t, tt.expected.CVV, result.CVV)
			assert.Equal(t, tt.expected.Description, result.Description)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
		})
	}
}

func TestNewBankCardsFromApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		appCards  []*bankcard.BankCard
		expected  []*BankCard
		expectNil bool
	}{
		{
			name: "valid app cards conversion",
			appCards: []*bankcard.BankCard{
				{
					ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					UserID:      uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
					CardNumber:  "4532015112830366",
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVV:         "123",
					Description: "Primary card",
				},
				{
					ID:          uuid.MustParse("789e0123-e89b-12d3-a456-426614174000"),
					UserID:      uuid.MustParse("456e7890-e89b-12d3-a456-426614174000"),
					CardNumber:  "5555555555554444",
					CardHolder:  "Jane Smith",
					ExpiryMonth: "06",
					ExpiryYear:  "26",
					CVV:         "456",
					Description: "Secondary card",
				},
			},
			expected: []*BankCard{
				{
					ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					CardNumber:  "4532015112830366",
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVV:         "123",
					Description: "Primary card",
				},
				{
					ID:          uuid.MustParse("789e0123-e89b-12d3-a456-426614174000"),
					CardNumber:  "5555555555554444",
					CardHolder:  "Jane Smith",
					ExpiryMonth: "06",
					ExpiryYear:  "26",
					CVV:         "456",
					Description: "Secondary card",
				},
			},
		},
		{
			name:      "nil slice",
			appCards:  nil,
			expectNil: true,
		},
		{
			name:     "empty slice",
			appCards: []*bankcard.BankCard{},
			expected: []*BankCard{},
		},
		{
			name: "slice with nil elements",
			appCards: []*bankcard.BankCard{
				{
					ID:         uuid.New(),
					CardNumber: "4532015112830366",
				},
				nil,
				{
					ID:         uuid.New(),
					CardNumber: "5555555555554444",
				},
			},
			expected: []*BankCard{
				{
					ID:         uuid.New(),
					CardNumber: "4532015112830366",
				},
				nil,
				{
					ID:         uuid.New(),
					CardNumber: "5555555555554444",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewBankCardsFromApp(tt.appCards)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Len(t, result, len(tt.expected))

			for i, expectedCard := range tt.expected {
				if expectedCard == nil {
					assert.Nil(t, result[i])
					continue
				}

				require.NotNil(t, result[i])
				assert.Equal(t, expectedCard.CardNumber, result[i].CardNumber)
			}
		})
	}
}

func TestBankCard_JSONSerialization(t *testing.T) {
	t.Parallel()

	card := &BankCard{
		ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		CardNumber:  "4532015112830366",
		CardHolder:  "John Doe",
		ExpiryMonth: "12",
		ExpiryYear:  "25",
		CVV:         "123",
		Description: "Primary card",
		UpdatedAt:   time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(card)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled BankCard
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, card.ID, unmarshaled.ID)
	assert.Equal(t, card.CardNumber, unmarshaled.CardNumber)
	assert.Equal(t, card.CardHolder, unmarshaled.CardHolder)
	assert.Equal(t, card.ExpiryMonth, unmarshaled.ExpiryMonth)
	assert.Equal(t, card.ExpiryYear, unmarshaled.ExpiryYear)
	assert.Equal(t, card.CVV, unmarshaled.CVV)
	assert.Equal(t, card.Description, unmarshaled.Description)
	assert.Equal(t, card.UpdatedAt.Unix(), unmarshaled.UpdatedAt.Unix())
}

func TestPushRequest_JSONSerialization(t *testing.T) {
	t.Parallel()

	req := &PushRequest{
		CardNumber:  "4532015112830366",
		CardHolder:  "John Doe",
		ExpiryMonth: "12",
		ExpiryYear:  "25",
		CVV:         "123",
		Description: "Primary card",
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(req)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled PushRequest
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.CardNumber, unmarshaled.CardNumber)
	assert.Equal(t, req.CardHolder, unmarshaled.CardHolder)
	assert.Equal(t, req.ExpiryMonth, unmarshaled.ExpiryMonth)
	assert.Equal(t, req.ExpiryYear, unmarshaled.ExpiryYear)
	assert.Equal(t, req.CVV, unmarshaled.CVV)
	assert.Equal(t, req.Description, unmarshaled.Description)
}

func TestPullRequest_Construction(t *testing.T) {
	t.Parallel()

	req := &PullRequest{
		ID: "123e4567-e89b-12d3-a456-426614174000",
	}

	// Test struct construction and field access
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", req.ID)
	assert.NotEmpty(t, req.ID)
}

func TestResponses_JSONSerialization(t *testing.T) {
	t.Parallel()

	// Test PullResponse
	pullResp := &PullResponse{
		BankCard: &BankCard{
			ID:         uuid.New(),
			CardNumber: "4532015112830366",
		},
	}

	jsonBytes, err := json.Marshal(pullResp)
	require.NoError(t, err)

	var unmarshaledPull PullResponse
	err = json.Unmarshal(jsonBytes, &unmarshaledPull)
	require.NoError(t, err)
	assert.Equal(t, pullResp.BankCard.CardNumber, unmarshaledPull.BankCard.CardNumber)

	// Test ListResponse
	listResp := &ListResponse{
		BankCards: []*BankCard{
			{ID: uuid.New(), CardNumber: "4532015112830366"},
			{ID: uuid.New(), CardNumber: "5555555555554444"},
		},
	}

	jsonBytes, err = json.Marshal(listResp)
	require.NoError(t, err)

	var unmarshaledList ListResponse
	err = json.Unmarshal(jsonBytes, &unmarshaledList)
	require.NoError(t, err)
	assert.Len(t, unmarshaledList.BankCards, 2)

	// Test PushResponse
	pushResp := &PushResponse{
		ID: uuid.New(),
	}

	jsonBytes, err = json.Marshal(pushResp)
	require.NoError(t, err)

	var unmarshaledPush PushResponse
	err = json.Unmarshal(jsonBytes, &unmarshaledPush)
	require.NoError(t, err)
	assert.Equal(t, pushResp.ID, unmarshaledPush.ID)
}
