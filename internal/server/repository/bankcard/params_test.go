package bankcard

import (
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		entity *bankcard.BankCard
		name   string
	}{
		{
			name: "save params with valid bank card entity",
			entity: &bankcard.BankCard{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				CardNumber:  []byte("4111111111111111"),
				CardHolder:  []byte("John Doe"),
				ExpiryMonth: []byte("12"),
				ExpiryYear:  []byte("2025"),
				CVV:         []byte("123"),
				Description: []byte("Main credit card"),
			},
		},
		{
			name:   "save params with nil entity",
			entity: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := SaveParams{
				Entity: tt.entity,
			}

			assert.Equal(t, tt.entity, params.Entity)
		})
	}
}

func TestLoadParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testID := uuid.New()

	tests := []struct {
		name   string
		userID uuid.UUID
		id     uuid.UUID
	}{
		{
			name:   "load params with user ID only",
			userID: testUserID,
			id:     uuid.Nil,
		},
		{
			name:   "load params with both user ID and specific ID",
			userID: testUserID,
			id:     testID,
		},
		{
			name:   "load params with nil values",
			userID: uuid.Nil,
			id:     uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := LoadParams{
				UserID: tt.userID,
				ID:     tt.id,
			}

			assert.Equal(t, tt.userID, params.UserID)
			assert.Equal(t, tt.id, params.ID)
		})
	}
}
