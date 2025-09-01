package crypto

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptAESGCM(t *testing.T) {
	t.Parallel()

	// Valid 32-byte key for AES-256
	validKey := make([]byte, 32)
	copy(validKey, "testtesttesttesttesttesttesttest")

	// Valid 16-byte key for AES-128
	validKey16 := make([]byte, 16)
	copy(validKey16, "testtesttesttest")

	tests := []struct {
		validateFunc func(*testing.T, []byte, []byte, []byte)
		name         string
		key          []byte
		plaintext    []byte
		expectError  bool
	}{
		{
			name:        "valid encryption with 32-byte key",
			key:         validKey,
			plaintext:   []byte("Hello, World!"),
			expectError: false,
			validateFunc: func(t *testing.T, key, plaintext, ciphertext []byte) {
				t.Helper()
				// Verify the result can be decrypted
				decrypted, err := DecryptAESGCM(key, ciphertext)
				require.NoError(t, err)
				assert.Equal(t, plaintext, decrypted)

				// Verify nonce is prepended (GCM nonce is 12 bytes)
				assert.GreaterOrEqual(t, len(ciphertext), 12)
			},
		},
		{
			name:        "valid encryption with 16-byte key",
			key:         validKey16,
			plaintext:   []byte("Test message"),
			expectError: false,
			validateFunc: func(t *testing.T, key, plaintext, ciphertext []byte) {
				t.Helper()
				decrypted, err := DecryptAESGCM(key, ciphertext)
				require.NoError(t, err)
				assert.Equal(t, plaintext, decrypted)
			},
		},
		{
			name:        "empty plaintext",
			key:         validKey,
			plaintext:   []byte(""),
			expectError: false,
			validateFunc: func(t *testing.T, key, plaintext, ciphertext []byte) {
				t.Helper()
				decrypted, err := DecryptAESGCM(key, ciphertext)
				require.NoError(t, err)
				// The actual behavior: empty slice input becomes nil on decryption
				if len(plaintext) == 0 {
					assert.True(t, len(decrypted) == 0, "Decrypted should be empty")
				} else {
					assert.Equal(t, plaintext, decrypted)
				}
			},
		},
		{
			name:        "large plaintext",
			key:         validKey,
			plaintext:   bytes.Repeat([]byte("A"), 10000),
			expectError: false,
			validateFunc: func(t *testing.T, key, plaintext, ciphertext []byte) {
				t.Helper()
				decrypted, err := DecryptAESGCM(key, ciphertext)
				require.NoError(t, err)
				assert.Equal(t, plaintext, decrypted)
			},
		},
		{
			name:        "unicode plaintext",
			key:         validKey,
			plaintext:   []byte("Hello 世界 🌍 测试"),
			expectError: false,
			validateFunc: func(t *testing.T, key, plaintext, ciphertext []byte) {
				t.Helper()
				decrypted, err := DecryptAESGCM(key, ciphertext)
				require.NoError(t, err)
				assert.Equal(t, plaintext, decrypted)
			},
		},
		{
			name:        "invalid key length",
			key:         []byte("short"),
			plaintext:   []byte("test"),
			expectError: true,
		},
		{
			name:        "nil key",
			key:         nil,
			plaintext:   []byte("test"),
			expectError: true,
		},
		{
			name:        "nil plaintext",
			key:         validKey,
			plaintext:   nil,
			expectError: false,
			validateFunc: func(t *testing.T, key, plaintext, ciphertext []byte) {
				t.Helper()
				decrypted, err := DecryptAESGCM(key, ciphertext)
				require.NoError(t, err)
				// nil plaintext becomes empty slice after round trip
				if plaintext == nil {
					assert.Equal(t, []byte(nil), decrypted)
				} else {
					assert.Equal(t, plaintext, decrypted)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := EncryptAESGCM(tt.key, tt.plaintext)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.key, tt.plaintext, result)
			}
		})
	}
}

func TestDecryptAESGCM(t *testing.T) {
	validKey := make([]byte, 32)
	copy(validKey, "testtesttesttesttesttesttesttest")

	tests := []struct {
		setupFunc   func() ([]byte, []byte, []byte)
		name        string
		expectError bool
	}{
		{
			name: "valid decryption",
			setupFunc: func() ([]byte, []byte, []byte) {
				plaintext := []byte("test message")
				ciphertext, err := EncryptAESGCM(validKey, plaintext)
				require.NoError(t, err)
				return validKey, ciphertext, plaintext
			},
			expectError: false,
		},
		{
			name: "invalid key",
			setupFunc: func() ([]byte, []byte, []byte) {
				plaintext := []byte("test message")
				ciphertext, err := EncryptAESGCM(validKey, plaintext)
				require.NoError(t, err)
				wrongKey := make([]byte, 32)
				copy(wrongKey, "wrongwrongwrongwrongwrongwrongwr")
				return wrongKey, ciphertext, nil
			},
			expectError: true,
		},
		{
			name: "invalid key length",
			setupFunc: func() ([]byte, []byte, []byte) {
				plaintext := []byte("test message")
				ciphertext, err := EncryptAESGCM(validKey, plaintext)
				require.NoError(t, err)
				return []byte("short"), ciphertext, nil
			},
			expectError: true,
		},
		{
			name: "nil key",
			setupFunc: func() ([]byte, []byte, []byte) {
				plaintext := []byte("test message")
				ciphertext, err := EncryptAESGCM(validKey, plaintext)
				require.NoError(t, err)
				return nil, ciphertext, nil
			},
			expectError: true,
		},
		{
			name: "data too short",
			setupFunc: func() ([]byte, []byte, []byte) {
				return validKey, []byte("short"), nil
			},
			expectError: true,
		},
		{
			name: "empty data",
			setupFunc: func() ([]byte, []byte, []byte) {
				return validKey, []byte{}, nil
			},
			expectError: true,
		},
		{
			name: "nil data",
			setupFunc: func() ([]byte, []byte, []byte) {
				return validKey, nil, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, data, expected := tt.setupFunc()

			result, err := DecryptAESGCM(key, data)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expected, result)
		})
	}
}

func TestAESGCM_RoundTrip(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "simple text",
			plaintext: []byte("Hello, World!"),
		},
		{
			name:      "empty data",
			plaintext: []byte{},
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name:      "unicode text",
			plaintext: []byte("测试 🚀 émoji"),
		},
		{
			name:      "large data",
			plaintext: bytes.Repeat([]byte("test"), 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Encrypt
			ciphertext, err := EncryptAESGCM(key, tt.plaintext)
			require.NoError(t, err)

			// Decrypt
			decrypted, err := DecryptAESGCM(key, ciphertext)
			require.NoError(t, err)

			// Verify
			if len(tt.plaintext) == 0 && len(decrypted) == 0 {
				// Both are empty, test passes
				return
			}
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestAESGCM_NonceUniqueness(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	plaintext := []byte("test message")

	// Encrypt same message multiple times
	ciphertexts := make([][]byte, 100)
	for i := range 100 {
		ciphertext, err := EncryptAESGCM(key, plaintext)
		require.NoError(t, err)
		ciphertexts[i] = ciphertext
	}

	// Verify all nonces are different
	nonces := make(map[string]bool)
	for _, ciphertext := range ciphertexts {
		nonce := string(ciphertext[:12]) // GCM nonce is 12 bytes
		assert.False(t, nonces[nonce], "Nonce should be unique")
		nonces[nonce] = true
	}
}

func TestHashBcrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(*testing.T, string, string)
		name         string
		data         string
		expectError  bool
	}{
		{
			name:        "valid password",
			data:        "password123",
			expectError: false,
			validateFunc: func(t *testing.T, data, hash string) {
				t.Helper()
				// Verify the hash can be verified
				verified, err := VerifyBcrypt(hash, data)
				require.NoError(t, err)
				assert.True(t, verified)

				// Verify hash format
				assert.True(t, strings.HasPrefix(hash, "$2"))
			},
		},
		{
			name:        "empty password",
			data:        "",
			expectError: false,
			validateFunc: func(t *testing.T, data, hash string) {
				t.Helper()
				verified, err := VerifyBcrypt(hash, data)
				require.NoError(t, err)
				assert.True(t, verified)
			},
		},
		{
			name:        "unicode password",
			data:        "测试密码🔐",
			expectError: false,
			validateFunc: func(t *testing.T, data, hash string) {
				t.Helper()
				verified, err := VerifyBcrypt(hash, data)
				require.NoError(t, err)
				assert.True(t, verified)
			},
		},
		{
			name:        "special characters",
			data:        "p@$$w0rd!@#",
			expectError: false,
			validateFunc: func(t *testing.T, data, hash string) {
				t.Helper()
				verified, err := VerifyBcrypt(hash, data)
				require.NoError(t, err)
				assert.True(t, verified)
			},
		},
		{
			name:        "maximum valid length",
			data:        strings.Repeat("a", MaxBcryptInputLength),
			expectError: false,
			validateFunc: func(t *testing.T, data, hash string) {
				t.Helper()
				verified, err := VerifyBcrypt(hash, data)
				require.NoError(t, err)
				assert.True(t, verified)
			},
		},
		{
			name:        "exceeds maximum length",
			data:        strings.Repeat("a", MaxBcryptInputLength+1),
			expectError: true,
		},
		{
			name:        "much longer than maximum",
			data:        strings.Repeat("a", 1000),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := HashBcrypt(tt.data)

			if tt.expectError {
				require.Error(t, err)
				assert.Empty(t, result)
				assert.Contains(t, err.Error(), "bcrypt error")
				assert.Contains(t, err.Error(), "exceeds maximum length")
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.NotEqual(t, tt.data, result, "Hash should not equal original data")

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.data, result)
			}
		})
	}
}

func TestVerifyBcrypt(t *testing.T) {
	t.Parallel()

	// Create valid hash for testing
	validPassword := "testpassword"
	validHash, err := HashBcrypt(validPassword)
	require.NoError(t, err)

	tests := []struct {
		name          string
		hashedData    string
		verifyingData string
		expectedMatch bool
		expectError   bool
	}{
		{
			name:          "correct password",
			hashedData:    validHash,
			verifyingData: validPassword,
			expectedMatch: true,
			expectError:   false,
		},
		{
			name:          "incorrect password",
			hashedData:    validHash,
			verifyingData: "wrongpassword",
			expectedMatch: false,
			expectError:   false,
		},
		{
			name:          "empty password against valid hash",
			hashedData:    validHash,
			verifyingData: "",
			expectedMatch: false,
			expectError:   false,
		},
		{
			name:          "case sensitive",
			hashedData:    validHash,
			verifyingData: "TestPassword",
			expectedMatch: false,
			expectError:   false,
		},
		{
			name:          "invalid hash format",
			hashedData:    "invalid_hash",
			verifyingData: "anypassword",
			expectedMatch: false,
			expectError:   true,
		},
		{
			name:          "empty hash",
			hashedData:    "",
			verifyingData: "anypassword",
			expectedMatch: false,
			expectError:   true,
		},
		{
			name:          "malformed bcrypt hash",
			hashedData:    "$2a$10$invalid",
			verifyingData: "anypassword",
			expectedMatch: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, err := VerifyBcrypt(tt.hashedData, tt.verifyingData)

			if tt.expectError {
				require.Error(t, err)
				assert.False(t, match)
				assert.Contains(t, err.Error(), "bcrypt error")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMatch, match)
		})
	}
}

func TestBcrypt_HashUniqueness(t *testing.T) {
	t.Parallel()

	password := "testpassword"
	hashes := make(map[string]bool)

	// Generate multiple hashes of the same password
	for range 10 {
		hash, err := HashBcrypt(password)
		require.NoError(t, err)

		// Each hash should be unique due to salt
		assert.False(t, hashes[hash], "Hash should be unique due to salt")
		hashes[hash] = true

		// But should still verify correctly
		verified, err := VerifyBcrypt(hash, password)
		require.NoError(t, err)
		assert.True(t, verified)
	}
}

func TestBcrypt_EmptyStringHandling(t *testing.T) {
	t.Parallel()

	// Test empty string hashing and verification
	hash, err := HashBcrypt("")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Should verify correctly
	verified, err := VerifyBcrypt(hash, "")
	require.NoError(t, err)
	assert.True(t, verified)

	// Should not verify with non-empty string
	verified, err = VerifyBcrypt(hash, "nonempty")
	require.NoError(t, err)
	assert.False(t, verified)
}

func TestConstants(t *testing.T) {
	t.Parallel()

	// Test that the constant matches bcrypt's actual limit
	assert.Equal(t, 72, MaxBcryptInputLength, "MaxBcryptInputLength should match bcrypt's limit")
}
