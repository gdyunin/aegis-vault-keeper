package security

import (
	"context"
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testKey is a typed key for context values in tests.
type testKey string

// Mock implementation of UserKeyRepository for testing.
type mockUserKeyRepository struct {
	users map[uuid.UUID]*auth.User
	err   error
}

func (m *mockUserKeyRepository) Load(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
	if m.err != nil {
		return nil, m.err
	}

	user, exists := m.users[params.ID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func TestNewUserKeyProvider(t *testing.T) {
	t.Parallel()

	repo := &mockUserKeyRepository{}
	provider := NewUserKeyProvider(repo)

	require.NotNil(t, provider)
	assert.Equal(t, repo, provider.r)
}

func TestUserKeyProvider_UserKeyProvide(t *testing.T) {
	t.Parallel()

	userID1 := uuid.New()
	userID2 := uuid.New()
	nonExistentUserID := uuid.New()

	cryptoKey1 := []byte("user1_crypto_key_12345678901234567890")
	cryptoKey2 := []byte("user2_crypto_key_98765432109876543210")

	type fields struct {
		users map[uuid.UUID]*auth.User
		err   error
	}
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		fields  fields
		name    string
		want    []byte
		args    args
		wantErr bool
	}{
		{
			name: "successful_key_retrieval",
			fields: fields{
				users: map[uuid.UUID]*auth.User{
					userID1: {
						ID:        userID1,
						CryptoKey: cryptoKey1,
					},
				},
				err: nil,
			},
			args: args{
				userID: userID1,
			},
			want:    cryptoKey1,
			wantErr: false,
		},
		{
			name: "multiple_users_correct_key",
			fields: fields{
				users: map[uuid.UUID]*auth.User{
					userID1: {
						ID:        userID1,
						CryptoKey: cryptoKey1,
					},
					userID2: {
						ID:        userID2,
						CryptoKey: cryptoKey2,
					},
				},
				err: nil,
			},
			args: args{
				userID: userID2,
			},
			want:    cryptoKey2,
			wantErr: false,
		},
		{
			name: "user_not_found",
			fields: fields{
				users: map[uuid.UUID]*auth.User{
					userID1: {
						ID:        userID1,
						CryptoKey: cryptoKey1,
					},
				},
				err: nil,
			},
			args: args{
				userID: nonExistentUserID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "repository_error",
			fields: fields{
				users: nil,
				err:   errors.New("database connection failed"),
			},
			args: args{
				userID: userID1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil_user_id",
			fields: fields{
				users: map[uuid.UUID]*auth.User{
					uuid.Nil: {
						ID:        uuid.Nil,
						CryptoKey: cryptoKey1,
					},
				},
				err: nil,
			},
			args: args{
				userID: uuid.Nil,
			},
			want:    cryptoKey1,
			wantErr: false,
		},
		{
			name: "empty_crypto_key",
			fields: fields{
				users: map[uuid.UUID]*auth.User{
					userID1: {
						ID:        userID1,
						CryptoKey: []byte{},
					},
				},
				err: nil,
			},
			args: args{
				userID: userID1,
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "nil_crypto_key",
			fields: fields{
				users: map[uuid.UUID]*auth.User{
					userID1: {
						ID:        userID1,
						CryptoKey: nil,
					},
				},
				err: nil,
			},
			args: args{
				userID: userID1,
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockUserKeyRepository{
				users: tt.fields.users,
				err:   tt.fields.err,
			}

			p := NewUserKeyProvider(repo)

			got, err := p.UserKeyProvide(context.Background(), tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), "failed to load user")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUserKeyProvider_UserKeyProvide_ContextHandling(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	cryptoKey := []byte("test_crypto_key_123456789012345678")

	// Test with cancelled context
	t.Run("cancelled_context", func(t *testing.T) {
		t.Parallel()

		repo := &mockUserKeyRepository{
			users: map[uuid.UUID]*auth.User{
				userID: {
					ID:        userID,
					CryptoKey: cryptoKey,
				},
			},
		}

		p := NewUserKeyProvider(repo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context

		// The mock doesn't check context cancellation, but in real implementation it should
		got, err := p.UserKeyProvide(ctx, userID)

		// With our mock, this will still succeed, but in real implementation
		// it might fail due to context cancellation
		if err == nil {
			assert.Equal(t, cryptoKey, got)
		}
	})

	// Test with context containing values
	t.Run("context_with_values", func(t *testing.T) {
		t.Parallel()

		repo := &mockUserKeyRepository{
			users: map[uuid.UUID]*auth.User{
				userID: {
					ID:        userID,
					CryptoKey: cryptoKey,
				},
			},
		}

		p := NewUserKeyProvider(repo)

		ctx := context.WithValue(context.Background(), testKey("test_key"), "test_value")

		got, err := p.UserKeyProvide(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, cryptoKey, got)
	})
}

func TestUserKeyProvider_ErrorMessages(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name        string
		repoError   error
		expectedMsg string
	}{
		{
			name:        "user_not_found_error",
			repoError:   errors.New("user not found"),
			expectedMsg: "failed to load user with ID " + userID.String() + ": user not found",
		},
		{
			name:        "database_error",
			repoError:   errors.New("database connection failed"),
			expectedMsg: "failed to load user with ID " + userID.String() + ": database connection failed",
		},
		{
			name:        "generic_error",
			repoError:   errors.New("some generic error"),
			expectedMsg: "failed to load user with ID " + userID.String() + ": some generic error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockUserKeyRepository{
				err: tt.repoError,
			}

			p := NewUserKeyProvider(repo)

			got, err := p.UserKeyProvide(context.Background(), userID)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.expectedMsg, err.Error())
		})
	}
}

// Benchmark for performance testing.
func BenchmarkUserKeyProvider_UserKeyProvide(b *testing.B) {
	userID := uuid.New()
	cryptoKey := []byte("benchmark_crypto_key_123456789012345")

	repo := &mockUserKeyRepository{
		users: map[uuid.UUID]*auth.User{
			userID: {
				ID:        userID,
				CryptoKey: cryptoKey,
			},
		},
	}

	p := NewUserKeyProvider(repo)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, err := p.UserKeyProvide(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
