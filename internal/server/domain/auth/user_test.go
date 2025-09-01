package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing.
type mockPasswordHasher struct {
	hashFunc func(password string) (string, error)
}

func (m *mockPasswordHasher) PasswordHash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "hashed_" + password, nil
}

type mockCryptoKeyGenerator struct {
	generateFunc func(size int) ([]byte, error)
}

func (m *mockCryptoKeyGenerator) CryptoKeyGenerate(size int) ([]byte, error) {
	if m.generateFunc != nil {
		return m.generateFunc(size)
	}
	return make([]byte, size), nil
}

type mockPasswordVerificator struct {
	verifyFunc func(hashedData, verifyingData string) (bool, error)
}

func (m *mockPasswordVerificator) PasswordVerify(hashedData, verifyingData string) (bool, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(hashedData, verifyingData)
	}
	return hashedData == "hashed_"+verifyingData, nil
}

func TestNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hasher        PasswordHasher
		cryptoKeyGen  CryptoKeyGenerator
		expectedError error
		validateUser  func(*testing.T, *User)
		params        NewUserParams
		name          string
		expectError   bool
	}{
		{
			name: "valid user creation",
			params: NewUserParams{
				Login:    "validuser",
				Password: "validpassword",
			},
			hasher:       &mockPasswordHasher{},
			cryptoKeyGen: &mockCryptoKeyGenerator{},
			expectError:  false,
			validateUser: func(t *testing.T, user *User) {
				t.Helper()
				assert.Equal(t, "validuser", user.Login)
				assert.Equal(t, "hashed_validpassword", user.PasswordHash)
				assert.Len(t, user.CryptoKey, 32)
				assert.NotEqual(t, uuid.UUID{}, user.ID)
			},
		},
		{
			name: "minimum valid parameters",
			params: NewUserParams{
				Login:    "abcde",    // 5 chars minimum
				Password: "12345678", // 8 chars minimum
			},
			hasher:       &mockPasswordHasher{},
			cryptoKeyGen: &mockCryptoKeyGenerator{},
			expectError:  false,
			validateUser: func(t *testing.T, user *User) {
				t.Helper()
				assert.Equal(t, "abcde", user.Login)
				assert.Equal(t, "hashed_12345678", user.PasswordHash)
			},
		},
		{
			name: "maximum valid parameters",
			params: NewUserParams{
				Login:    strings.Repeat("a", 50), // 50 chars maximum
				Password: strings.Repeat("b", 64), // 64 chars maximum
			},
			hasher:       &mockPasswordHasher{},
			cryptoKeyGen: &mockCryptoKeyGenerator{},
			expectError:  false,
			validateUser: func(t *testing.T, user *User) {
				t.Helper()
				assert.Equal(t, strings.Repeat("a", 50), user.Login)
				assert.Equal(t, "hashed_"+strings.Repeat("b", 64), user.PasswordHash)
			},
		},
		{
			name: "invalid login too short",
			params: NewUserParams{
				Login:    "abc", // 3 chars, too short
				Password: "validpassword",
			},
			hasher:        &mockPasswordHasher{},
			cryptoKeyGen:  &mockCryptoKeyGenerator{},
			expectError:   true,
			expectedError: ErrNewUserParamsValidation,
		},
		{
			name: "invalid login too long",
			params: NewUserParams{
				Login:    strings.Repeat("a", 51), // 51 chars, too long
				Password: "validpassword",
			},
			hasher:        &mockPasswordHasher{},
			cryptoKeyGen:  &mockCryptoKeyGenerator{},
			expectError:   true,
			expectedError: ErrNewUserParamsValidation,
		},
		{
			name: "invalid password too short",
			params: NewUserParams{
				Login:    "validuser",
				Password: "1234567", // 7 chars, too short
			},
			hasher:        &mockPasswordHasher{},
			cryptoKeyGen:  &mockCryptoKeyGenerator{},
			expectError:   true,
			expectedError: ErrNewUserParamsValidation,
		},
		{
			name: "invalid password too long",
			params: NewUserParams{
				Login:    "validuser",
				Password: strings.Repeat("a", 65), // 65 chars, too long
			},
			hasher:        &mockPasswordHasher{},
			cryptoKeyGen:  &mockCryptoKeyGenerator{},
			expectError:   true,
			expectedError: ErrNewUserParamsValidation,
		},
		{
			name: "crypto key generation failure",
			params: NewUserParams{
				Login:    "validuser",
				Password: "validpassword",
			},
			hasher: &mockPasswordHasher{},
			cryptoKeyGen: &mockCryptoKeyGenerator{
				generateFunc: func(size int) ([]byte, error) {
					return nil, errors.New("crypto key generation failed")
				},
			},
			expectError:   true,
			expectedError: ErrCryptoKeyGenerate,
		},
		{
			name: "password hashing failure",
			params: NewUserParams{
				Login:    "validuser",
				Password: "validpassword",
			},
			hasher: &mockPasswordHasher{
				hashFunc: func(password string) (string, error) {
					return "", errors.New("password hashing failed")
				},
			},
			cryptoKeyGen:  &mockCryptoKeyGenerator{},
			expectError:   true,
			expectedError: ErrPasswordHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := NewUser(tt.params, tt.hasher, tt.cryptoKeyGen)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, user)
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			if tt.validateUser != nil {
				tt.validateUser(t, user)
			}
		})
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	t.Parallel()

	user := &User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: "hashed_correctpassword",
		CryptoKey:    []byte("test_crypto_key"),
	}

	tests := []struct {
		verificator   PasswordVerificator
		expectedError error
		name          string
		password      string
		expectSuccess bool
		expectError   bool
	}{
		{
			name:          "correct password verification",
			verificator:   &mockPasswordVerificator{},
			password:      "correctpassword",
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "incorrect password verification",
			verificator:   &mockPasswordVerificator{},
			password:      "wrongpassword",
			expectSuccess: false,
			expectError:   false,
		},
		{
			name: "password verification error",
			verificator: &mockPasswordVerificator{
				verifyFunc: func(hashedData, verifyingData string) (bool, error) {
					return false, errors.New("verification service error")
				},
			},
			password:      "anypassword",
			expectSuccess: false,
			expectError:   true,
			expectedError: ErrPasswordVerificationFailed,
		},
		{
			name:          "empty password",
			verificator:   &mockPasswordVerificator{},
			password:      "",
			expectSuccess: false,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			verified, err := user.VerifyPassword(tt.verificator, tt.password)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.False(t, verified)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, verified)
		})
	}
}

func TestNewUserParams_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedError error
		params        NewUserParams
		name          string
		expectError   bool
	}{
		{
			name: "valid parameters",
			params: NewUserParams{
				Login:    "validuser",
				Password: "validpassword",
			},
			expectError: false,
		},
		{
			name: "minimum valid lengths",
			params: NewUserParams{
				Login:    "abcde",    // 5 chars
				Password: "12345678", // 8 chars
			},
			expectError: false,
		},
		{
			name: "maximum valid lengths",
			params: NewUserParams{
				Login:    strings.Repeat("a", 50), // 50 chars
				Password: strings.Repeat("b", 64), // 64 chars
			},
			expectError: false,
		},
		{
			name: "login too short",
			params: NewUserParams{
				Login:    "abc", // 3 chars
				Password: "validpassword",
			},
			expectError:   true,
			expectedError: ErrIncorrectLogin,
		},
		{
			name: "login too long",
			params: NewUserParams{
				Login:    strings.Repeat("a", 51), // 51 chars
				Password: "validpassword",
			},
			expectError:   true,
			expectedError: ErrIncorrectLogin,
		},
		{
			name: "password too short",
			params: NewUserParams{
				Login:    "validuser",
				Password: "1234567", // 7 chars
			},
			expectError:   true,
			expectedError: ErrIncorrectPassword,
		},
		{
			name: "password too long",
			params: NewUserParams{
				Login:    "validuser",
				Password: strings.Repeat("a", 65), // 65 chars
			},
			expectError:   true,
			expectedError: ErrIncorrectPassword,
		},
		{
			name: "both login and password invalid",
			params: NewUserParams{
				Login:    "ab",  // too short
				Password: "123", // too short
			},
			expectError:   true,
			expectedError: ErrIncorrectLogin, // will contain both errors
		},
		{
			name: "empty parameters",
			params: NewUserParams{
				Login:    "",
				Password: "",
			},
			expectError:   true,
			expectedError: ErrIncorrectLogin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewUserParams_validateLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		login       string
		expectError bool
	}{
		{
			name:        "valid login minimum length",
			login:       "abcde",
			expectError: false,
		},
		{
			name:        "valid login maximum length",
			login:       strings.Repeat("a", 50),
			expectError: false,
		},
		{
			name:        "valid login middle length",
			login:       "validuser123",
			expectError: false,
		},
		{
			name:        "login too short",
			login:       "abcd",
			expectError: true,
		},
		{
			name:        "login too long",
			login:       strings.Repeat("a", 51),
			expectError: true,
		},
		{
			name:        "empty login",
			login:       "",
			expectError: true,
		},
		{
			name:        "unicode login",
			login:       "user测试",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := &NewUserParams{
				Login: tt.login,
			}

			err := params.validateLogin()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrIncorrectLogin)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewUserParams_validatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password minimum length",
			password:    "12345678",
			expectError: false,
		},
		{
			name:        "valid password maximum length",
			password:    strings.Repeat("a", 64),
			expectError: false,
		},
		{
			name:        "valid password middle length",
			password:    "validpassword123",
			expectError: false,
		},
		{
			name:        "password too short",
			password:    "1234567",
			expectError: true,
		},
		{
			name:        "password too long",
			password:    strings.Repeat("a", 65),
			expectError: true,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
		},
		{
			name:        "unicode password",
			password:    "pass测试123",
			expectError: false,
		},
		{
			name:        "special characters password",
			password:    "p@ssw0rd!",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := &NewUserParams{
				Password: tt.password,
			}

			err := params.validatePassword()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrIncorrectPassword)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUser_FieldValidation(t *testing.T) {
	t.Parallel()

	hasher := &mockPasswordHasher{}
	cryptoKeyGen := &mockCryptoKeyGenerator{}

	params := NewUserParams{
		Login:    "testuser",
		Password: "testpassword",
	}

	user, err := NewUser(params, hasher, cryptoKeyGen)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Test that all fields are properly set
	assert.NotEqual(t, uuid.UUID{}, user.ID, "User ID should not be zero")
	assert.Equal(t, "testuser", user.Login, "Login should match input")
	assert.Equal(t, "hashed_testpassword", user.PasswordHash, "Password should be hashed")
	assert.Len(t, user.CryptoKey, 32, "CryptoKey should be 32 bytes")
}

func TestConstants(t *testing.T) {
	t.Parallel()

	// Test that constants have expected values
	assert.Equal(t, 32, cryptoKeySize, "CryptoKeySize should be 32")
	assert.Equal(t, 5, loginMinLen, "LoginMinLen should be 5")
	assert.Equal(t, 50, loginMaxLen, "LoginMaxLen should be 50")
	assert.Equal(t, 8, passwordMinLen, "PasswordMinLen should be 8")
	assert.Equal(t, 64, passwordMaxLen, "PasswordMaxLen should be 64")
}
