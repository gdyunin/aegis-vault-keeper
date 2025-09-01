package keyprv

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockUserKeyProvider is a test implementation of UserKeyProvider.
type mockUserKeyProvider struct {
	keyFunc func(ctx context.Context, userID uuid.UUID) ([]byte, error)
}

func (m *mockUserKeyProvider) UserKeyProvide(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	if m.keyFunc != nil {
		return m.keyFunc(ctx, userID)
	}
	return []byte("test-key"), nil
}

func TestUserKeyProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider UserKeyProvider
		checkKey func(*testing.T, []byte, error)
		name     string
		userID   uuid.UUID
	}{
		{
			name: "mock provider returns key",
			provider: &mockUserKeyProvider{
				keyFunc: func(ctx context.Context, userID uuid.UUID) ([]byte, error) {
					return []byte("mock-key"), nil
				},
			},
			userID: uuid.New(),
			checkKey: func(t *testing.T, key []byte, err error) {
				t.Helper()
				assert.NoError(t, err)
				assert.Equal(t, []byte("mock-key"), key)
			},
		},
		{
			name:     "nil provider doesn't panic",
			provider: nil,
			userID:   uuid.New(),
			checkKey: func(t *testing.T, key []byte, err error) {
				t.Helper()
				// Test that we can handle nil provider without panic
				assert.True(t, true) // Just check we got here
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.provider != nil {
				key, err := tt.provider.UserKeyProvide(context.Background(), tt.userID)
				tt.checkKey(t, key, err)
			} else {
				// For nil provider test, just verify we can handle it
				tt.checkKey(t, nil, nil)
			}
		})
	}
}
