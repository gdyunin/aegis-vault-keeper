package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params RegisterParams
		expect RegisterParams
	}{
		{
			name: "valid_params",
			params: RegisterParams{
				Login:    "testuser",
				Password: "testpass123",
			},
			expect: RegisterParams{
				Login:    "testuser",
				Password: "testpass123",
			},
		},
		{
			name: "empty_params",
			params: RegisterParams{
				Login:    "",
				Password: "",
			},
			expect: RegisterParams{
				Login:    "",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expect.Login, tt.params.Login)
			assert.Equal(t, tt.expect.Password, tt.params.Password)
		})
	}
}

func TestLoginParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params LoginParams
		expect LoginParams
	}{
		{
			name: "valid_params",
			params: LoginParams{
				Login:    "testuser",
				Password: "testpass123",
			},
			expect: LoginParams{
				Login:    "testuser",
				Password: "testpass123",
			},
		},
		{
			name: "empty_params",
			params: LoginParams{
				Login:    "",
				Password: "",
			},
			expect: LoginParams{
				Login:    "",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expect.Login, tt.params.Login)
			assert.Equal(t, tt.expect.Password, tt.params.Password)
		})
	}
}

func TestAccessToken(t *testing.T) {
	t.Parallel()

	testTime := time.Now().Add(time.Hour)

	tests := []struct {
		name   string
		token  AccessToken
		expect AccessToken
	}{
		{
			name: "valid_token",
			token: AccessToken{
				AccessToken: "jwt_token_string",
				ExpiresAt:   testTime,
				TokenType:   "Bearer",
			},
			expect: AccessToken{
				AccessToken: "jwt_token_string",
				ExpiresAt:   testTime,
				TokenType:   "Bearer",
			},
		},
		{
			name: "empty_token",
			token: AccessToken{
				AccessToken: "",
				ExpiresAt:   time.Time{},
				TokenType:   "",
			},
			expect: AccessToken{
				AccessToken: "",
				ExpiresAt:   time.Time{},
				TokenType:   "",
			},
		},
		{
			name: "custom_token_type",
			token: AccessToken{
				AccessToken: "custom_token",
				ExpiresAt:   testTime,
				TokenType:   "Custom",
			},
			expect: AccessToken{
				AccessToken: "custom_token",
				ExpiresAt:   testTime,
				TokenType:   "Custom",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expect.AccessToken, tt.token.AccessToken)
			assert.Equal(t, tt.expect.ExpiresAt, tt.token.ExpiresAt)
			assert.Equal(t, tt.expect.TokenType, tt.token.TokenType)

			// Verify struct fields are accessible
			if tt.expect.AccessToken != "" {
				require.NotEmpty(t, tt.token.AccessToken)
			}
			if !tt.expect.ExpiresAt.IsZero() {
				assert.False(t, tt.token.ExpiresAt.IsZero())
			}
		})
	}
}
