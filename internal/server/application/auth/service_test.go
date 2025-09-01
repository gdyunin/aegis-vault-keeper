package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errMockNotImplemented = errors.New("mock function not implemented")
)

// Mock implementations for testing.
type mockRepository struct {
	saveFunc func(ctx context.Context, params repository.SaveParams) error
	loadFunc func(ctx context.Context, params repository.LoadParams) (*auth.User, error)
}

func (m *mockRepository) Save(ctx context.Context, params repository.SaveParams) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, params)
	}
	return nil
}

func (m *mockRepository) Load(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx, params)
	}
	return nil, errMockNotImplemented
}

type mockPasswordHasherVerificator struct {
	hashFunc   func(password string) (string, error)
	verifyFunc func(hash, password string) (bool, error)
}

func (m *mockPasswordHasherVerificator) PasswordHash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "hashed_password", nil
}

func (m *mockPasswordHasherVerificator) PasswordVerify(hash, password string) (bool, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(hash, password)
	}
	return true, nil
}

type mockCryptoKeyGenerator struct {
	generateFunc func(size int) ([]byte, error)
}

func (m *mockCryptoKeyGenerator) CryptoKeyGenerate(size int) ([]byte, error) {
	if m.generateFunc != nil {
		return m.generateFunc(size)
	}
	return []byte("test_key"), nil
}

type mockTokenGenerateValidator struct {
	generateFunc func(userID uuid.UUID) (string, string, time.Time, error)
	validateFunc func(tokenString string) (uuid.UUID, error)
}

func (m *mockTokenGenerateValidator) GenerateAccessToken(
	userID uuid.UUID,
) (string, string, time.Time, error) {
	if m.generateFunc != nil {
		return m.generateFunc(userID)
	}
	return "test_token", "Bearer", time.Now().Add(time.Hour), nil
}

func (m *mockTokenGenerateValidator) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	if m.validateFunc != nil {
		return m.validateFunc(tokenString)
	}
	return uuid.New(), nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{}
	hasher := &mockPasswordHasherVerificator{}
	keyGen := &mockCryptoKeyGenerator{}
	tokenGen := &mockTokenGenerateValidator{}

	service := NewService(repo, hasher, keyGen, tokenGen)

	require.NotNil(t, service)
	assert.Equal(t, repo, service.r)
	assert.Equal(t, hasher, service.passwordHasherVerificator)
	assert.Equal(t, keyGen, service.cryptoKeyGenerator)
	assert.Equal(t, tokenGen, service.tokenGenerateValidator)
}

func TestService_Register(t *testing.T) {
	t.Parallel()

	type args struct {
		params RegisterParams
	}
	tests := []struct {
		setupMocks     func(*mockRepository, *mockPasswordHasherVerificator, *mockCryptoKeyGenerator)
		args           args
		name           string
		expectedErrMsg string
		wantErr        bool
	}{
		{
			name: "successful_registration",
			args: args{
				params: RegisterParams{
					Login:    "testuser",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, keyGen *mockCryptoKeyGenerator) {
				hasher.hashFunc = func(password string) (string, error) {
					return "hashed_" + password, nil
				}
				keyGen.generateFunc = func(size int) ([]byte, error) {
					return []byte("crypto_key"), nil
				}
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "user_creation_failed",
			args: args{
				params: RegisterParams{
					Login:    "",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, keyGen *mockCryptoKeyGenerator) {
				hasher.hashFunc = func(password string) (string, error) {
					return "", auth.ErrIncorrectLogin
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to create new user",
		},
		{
			name: "repository_save_failed",
			args: args{
				params: RegisterParams{
					Login:    "testuser",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, keyGen *mockCryptoKeyGenerator) {
				hasher.hashFunc = func(password string) (string, error) {
					return "hashed_" + password, nil
				}
				keyGen.generateFunc = func(size int) ([]byte, error) {
					return []byte("crypto_key"), nil
				}
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return repository.ErrUserAlreadyExists
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to save user",
		},
		{
			name: "crypto_key_generation_failed",
			args: args{
				params: RegisterParams{
					Login:    "testuser",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, keyGen *mockCryptoKeyGenerator) {
				hasher.hashFunc = func(password string) (string, error) {
					return "hashed_" + password, nil
				}
				keyGen.generateFunc = func(size int) ([]byte, error) {
					return nil, errors.New("key generation failed")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to create new user",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			hasher := &mockPasswordHasherVerificator{}
			keyGen := &mockCryptoKeyGenerator{}
			tokenGen := &mockTokenGenerateValidator{}

			if tt.setupMocks != nil {
				tt.setupMocks(repo, hasher, keyGen)
			}

			service := NewService(repo, hasher, keyGen, tokenGen)
			userID, err := service.Register(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Equal(t, uuid.Nil, userID)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testUser := &auth.User{
		ID:           testUserID,
		Login:        "testuser",
		PasswordHash: "hashed_password",
		CryptoKey:    []byte("crypto_key"),
	}

	type args struct {
		params LoginParams
	}
	tests := []struct {
		setupMocks     func(*mockRepository, *mockPasswordHasherVerificator, *mockTokenGenerateValidator)
		args           args
		name           string
		expectedErrMsg string
		wantErr        bool
		expectToken    bool
	}{
		{
			name: "successful_login",
			args: args{
				params: LoginParams{
					Login:    "testuser",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, tokenGen *mockTokenGenerateValidator) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
					return testUser, nil
				}
				hasher.verifyFunc = func(hash, password string) (bool, error) {
					return true, nil
				}
				tokenGen.generateFunc = func(userID uuid.UUID) (string, string, time.Time, error) {
					return "access_token", "Bearer", time.Now().Add(time.Hour), nil
				}
			},
			expectToken: true,
			wantErr:     false,
		},
		{
			name: "user_not_found",
			args: args{
				params: LoginParams{
					Login:    "nonexistent",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, tokenGen *mockTokenGenerateValidator) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
					return nil, repository.ErrUserNotFound
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to load user",
		},
		{
			name: "wrong_password",
			args: args{
				params: LoginParams{
					Login:    "testuser",
					Password: "wrongpass",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, tokenGen *mockTokenGenerateValidator) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
					return testUser, nil
				}
				hasher.verifyFunc = func(hash, password string) (bool, error) {
					return false, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "authentication failed",
		},
		{
			name: "password_verification_error",
			args: args{
				params: LoginParams{
					Login:    "testuser",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, tokenGen *mockTokenGenerateValidator) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
					return testUser, nil
				}
				hasher.verifyFunc = func(hash, password string) (bool, error) {
					return false, errors.New("verification error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to verify password",
		},
		{
			name: "token_generation_failed",
			args: args{
				params: LoginParams{
					Login:    "testuser",
					Password: "testpass123",
				},
			},
			setupMocks: func(repo *mockRepository, hasher *mockPasswordHasherVerificator, tokenGen *mockTokenGenerateValidator) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) (*auth.User, error) {
					return testUser, nil
				}
				hasher.verifyFunc = func(hash, password string) (bool, error) {
					return true, nil
				}
				tokenGen.generateFunc = func(userID uuid.UUID) (string, string, time.Time, error) {
					return "", "", time.Time{}, errors.New("token generation failed")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to generate access token",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			hasher := &mockPasswordHasherVerificator{}
			keyGen := &mockCryptoKeyGenerator{}
			tokenGen := &mockTokenGenerateValidator{}

			if tt.setupMocks != nil {
				tt.setupMocks(repo, hasher, tokenGen)
			}

			service := NewService(repo, hasher, keyGen, tokenGen)
			token, err := service.Login(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Empty(t, token.AccessToken)
			} else {
				require.NoError(t, err)
				if tt.expectToken {
					assert.NotEmpty(t, token.AccessToken)
					assert.NotEmpty(t, token.TokenType)
					assert.False(t, token.ExpiresAt.IsZero())
				}
			}
		})
	}
}

func TestService_ValidateToken(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		setupMocks     func(*mockTokenGenerateValidator)
		name           string
		tokenString    string
		expectedErrMsg string
		expectedUserID uuid.UUID
		wantErr        bool
	}{
		{
			name:        "valid_token",
			tokenString: "valid_token_string",
			setupMocks: func(tokenGen *mockTokenGenerateValidator) {
				tokenGen.validateFunc = func(tokenString string) (uuid.UUID, error) {
					return testUserID, nil
				}
			},
			wantErr:        false,
			expectedUserID: testUserID,
		},
		{
			name:        "invalid_token",
			tokenString: "invalid_token_string",
			setupMocks: func(tokenGen *mockTokenGenerateValidator) {
				tokenGen.validateFunc = func(tokenString string) (uuid.UUID, error) {
					return uuid.Nil, errors.New("invalid token")
				}
			},
			wantErr:        true,
			expectedUserID: uuid.Nil,
			expectedErrMsg: "failed to validate access token",
		},
		{
			name:        "empty_token",
			tokenString: "",
			setupMocks: func(tokenGen *mockTokenGenerateValidator) {
				tokenGen.validateFunc = func(tokenString string) (uuid.UUID, error) {
					return uuid.Nil, errors.New("empty token")
				}
			},
			wantErr:        true,
			expectedUserID: uuid.Nil,
			expectedErrMsg: "failed to validate access token",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			hasher := &mockPasswordHasherVerificator{}
			keyGen := &mockCryptoKeyGenerator{}
			tokenGen := &mockTokenGenerateValidator{}

			if tt.setupMocks != nil {
				tt.setupMocks(tokenGen)
			}

			service := NewService(repo, hasher, keyGen, tokenGen)
			userID, err := service.ValidateToken(tt.tokenString)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Equal(t, uuid.Nil, userID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, userID)
			}
		})
	}
}
