package security

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenGenerateValidator(t *testing.T) {
	t.Parallel()

	validSecretKey := make([]byte, MinSecretKeyLength)
	for i := range validSecretKey {
		validSecretKey[i] = byte(i)
	}
	validDuration := time.Hour

	type args struct {
		secretKey                 []byte
		accessTokenExpireDuration time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid_secret_key",
			args: args{
				secretKey:                 validSecretKey,
				accessTokenExpireDuration: validDuration,
			},
			wantErr: false,
		},
		{
			name: "exact_minimum_key_length",
			args: args{
				secretKey:                 make([]byte, MinSecretKeyLength),
				accessTokenExpireDuration: validDuration,
			},
			wantErr: false,
		},
		{
			name: "longer_key_valid",
			args: args{
				secretKey:                 make([]byte, MinSecretKeyLength+10),
				accessTokenExpireDuration: validDuration,
			},
			wantErr: false,
		},
		{
			name: "key_too_short",
			args: args{
				secretKey:                 make([]byte, MinSecretKeyLength-1),
				accessTokenExpireDuration: validDuration,
			},
			wantErr: true,
		},
		{
			name: "empty_key",
			args: args{
				secretKey:                 []byte{},
				accessTokenExpireDuration: validDuration,
			},
			wantErr: true,
		},
		{
			name: "nil_key",
			args: args{
				secretKey:                 nil,
				accessTokenExpireDuration: validDuration,
			},
			wantErr: true,
		},
		{
			name: "zero_duration",
			args: args{
				secretKey:                 validSecretKey,
				accessTokenExpireDuration: 0,
			},
			wantErr: false, // Zero duration is valid, though not practical
		},
		{
			name: "negative_duration",
			args: args{
				secretKey:                 validSecretKey,
				accessTokenExpireDuration: -time.Hour,
			},
			wantErr: false, // Negative duration is valid, though not practical
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewTokenGenerateValidator(tt.args.secretKey, tt.args.accessTokenExpireDuration)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), "JWT error")
				assert.Contains(t, err.Error(), "secret key is too short")
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.args.secretKey, got.secretKey)
				assert.Equal(t, tt.args.accessTokenExpireDuration, got.accessTokenExpireDuration)
			}
		})
	}
}

func TestTokenGenerateValidator_GenerateAccessToken(t *testing.T) {
	t.Parallel()

	secretKey := make([]byte, MinSecretKeyLength)
	for i := range secretKey {
		secretKey[i] = byte(i % 256)
	}
	duration := time.Hour

	tgv, err := NewTokenGenerateValidator(secretKey, duration)
	require.NoError(t, err)

	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid_user_id",
			args: args{
				userID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "nil_user_id",
			args: args{
				userID: uuid.Nil,
			},
			wantErr: false, // uuid.Nil is a valid UUID
		},
		{
			name: "specific_user_id",
			args: args{
				userID: uuid.MustParse("12345678-1234-1234-1234-123456789012"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			beforeGeneration := time.Now()
			token, tokenType, expiresAt, err := tgv.GenerateAccessToken(tt.args.userID)
			afterGeneration := time.Now()

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				assert.Empty(t, tokenType)
				assert.True(t, expiresAt.IsZero())
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.Equal(t, TokenTypeBearer, tokenType)

				// Check that expiration time is roughly correct
				expectedExpiry := beforeGeneration.Add(duration)
				assert.True(t, expiresAt.After(expectedExpiry.Add(-time.Second)))
				assert.True(t, expiresAt.Before(afterGeneration.Add(duration).Add(time.Second)))

				// Validate that the token can be parsed back
				userID, err := tgv.ValidateAccessToken(token)
				require.NoError(t, err)
				assert.Equal(t, tt.args.userID, userID)
			}
		})
	}
}

func TestTokenGenerateValidator_ValidateAccessToken(t *testing.T) {
	t.Parallel()

	secretKey := make([]byte, MinSecretKeyLength)
	for i := range secretKey {
		secretKey[i] = byte(i % 256)
	}
	duration := time.Hour

	tgv, err := NewTokenGenerateValidator(secretKey, duration)
	require.NoError(t, err)

	// Generate a valid token for testing
	userID := uuid.New()
	validToken, _, _, err := tgv.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Create expired token generator for testing
	expiredTGV, err := NewTokenGenerateValidator(secretKey, -time.Hour) // Already expired
	require.NoError(t, err)
	expiredToken, _, _, err := expiredTGV.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Create token with different secret for testing
	differentSecretKey := make([]byte, MinSecretKeyLength)
	for i := range differentSecretKey {
		differentSecretKey[i] = byte((i + 1) % 256)
	}
	differentTGV, err := NewTokenGenerateValidator(differentSecretKey, duration)
	require.NoError(t, err)
	differentSecretToken, _, _, err := differentTGV.GenerateAccessToken(userID)
	require.NoError(t, err)

	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "valid_token",
			args: args{
				tokenString: validToken,
			},
			want:    userID,
			wantErr: false,
		},
		{
			name: "invalid_token_empty",
			args: args{
				tokenString: "",
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "invalid_token_malformed",
			args: args{
				tokenString: "invalid.token.string",
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "invalid_token_random_string",
			args: args{
				tokenString: "completely-random-string",
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "expired_token",
			args: args{
				tokenString: expiredToken,
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "token_with_different_secret",
			args: args{
				tokenString: differentSecretToken,
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "token_with_extra_parts",
			args: args{
				tokenString: validToken + ".extra",
			},
			want:    uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tgv.ValidateAccessToken(tt.args.tokenString)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, uuid.Nil, got)
				assert.Contains(t, err.Error(), "JWT error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTokenGenerateValidator_RoundTrip(t *testing.T) {
	t.Parallel()

	secretKey := make([]byte, MinSecretKeyLength)
	for i := range secretKey {
		secretKey[i] = byte(i % 256)
	}
	duration := time.Hour

	tgv, err := NewTokenGenerateValidator(secretKey, duration)
	require.NoError(t, err)

	// Test multiple user IDs
	userIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.Nil,
		uuid.MustParse("12345678-1234-1234-1234-123456789012"),
	}

	for i, userID := range userIDs {
		t.Run(func() string { return "user_" + string(rune(i)) }(), func(t *testing.T) {
			t.Parallel()

			// Generate token
			token, tokenType, expiresAt, err := tgv.GenerateAccessToken(userID)
			require.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.Equal(t, TokenTypeBearer, tokenType)
			assert.True(t, expiresAt.After(time.Now()))

			// Validate token
			gotUserID, err := tgv.ValidateAccessToken(token)
			require.NoError(t, err)
			assert.Equal(t, userID, gotUserID)
		})
	}
}

func TestClaims(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	claims := &Claims{
		UserID: userID,
	}

	assert.Equal(t, userID, claims.UserID)
}

func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Bearer", TokenTypeBearer)
	assert.Equal(t, 32, MinSecretKeyLength)
}

// Benchmark token generation and validation.
func BenchmarkTokenGenerateValidator_GenerateAccessToken(b *testing.B) {
	secretKey := make([]byte, MinSecretKeyLength)
	tgv, _ := NewTokenGenerateValidator(secretKey, time.Hour)
	userID := uuid.New()

	b.ResetTimer()
	for range b.N {
		_, _, _, err := tgv.GenerateAccessToken(userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenGenerateValidator_ValidateAccessToken(b *testing.B) {
	secretKey := make([]byte, MinSecretKeyLength)
	tgv, _ := NewTokenGenerateValidator(secretKey, time.Hour)
	userID := uuid.New()
	token, _, _, _ := tgv.GenerateAccessToken(userID)

	b.ResetTimer()
	for range b.N {
		_, err := tgv.ValidateAccessToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}
