package auth

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(t *testing.T, req RegisterRequest)
		request      RegisterRequest
		name         string
		jsonData     string
		expectValid  bool
	}{
		{
			name: "valid register request",
			request: RegisterRequest{
				Login:    "user@example.com",
				Password: "securePassword123",
			},
			expectValid: true,
			validateFunc: func(t *testing.T, req RegisterRequest) {
				t.Helper()
				assert.Equal(t, "user@example.com", req.Login)
				assert.Equal(t, "securePassword123", req.Password)
			},
		},
		{
			name: "empty login",
			request: RegisterRequest{
				Login:    "",
				Password: "securePassword123",
			},
			expectValid: false,
		},
		{
			name: "empty password",
			request: RegisterRequest{
				Login:    "user@example.com",
				Password: "",
			},
			expectValid: false,
		},
		{
			name: "both fields empty",
			request: RegisterRequest{
				Login:    "",
				Password: "",
			},
			expectValid: false,
		},
		{
			name:     "valid JSON unmarshaling",
			jsonData: `{"login":"test@example.com","password":"myPassword"}`,
			validateFunc: func(t *testing.T, req RegisterRequest) {
				t.Helper()
				assert.Equal(t, "test@example.com", req.Login)
				assert.Equal(t, "myPassword", req.Password)
			},
		},
		{
			name:     "JSON with extra fields ignored",
			jsonData: `{"login":"test@example.com","password":"myPassword","extra":"ignored"}`,
			validateFunc: func(t *testing.T, req RegisterRequest) {
				t.Helper()
				assert.Equal(t, "test@example.com", req.Login)
				assert.Equal(t, "myPassword", req.Password)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.jsonData != "" {
				var req RegisterRequest
				err := json.Unmarshal([]byte(tt.jsonData), &req)
				require.NoError(t, err)
				tt.validateFunc(t, req)
			} else if tt.validateFunc != nil {
				tt.validateFunc(t, tt.request)
			}

			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.request)
			require.NoError(t, err)
			assert.Contains(t, string(jsonBytes), `"login"`)
			assert.Contains(t, string(jsonBytes), `"password"`)
		})
	}
}

func TestLoginRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(t *testing.T, req LoginRequest)
		request      LoginRequest
		name         string
		jsonData     string
		expectValid  bool
	}{
		{
			name: "valid login request",
			request: LoginRequest{
				Login:    "user@example.com",
				Password: "securePassword123",
			},
			expectValid: true,
			validateFunc: func(t *testing.T, req LoginRequest) {
				t.Helper()
				assert.Equal(t, "user@example.com", req.Login)
				assert.Equal(t, "securePassword123", req.Password)
			},
		},
		{
			name: "empty login",
			request: LoginRequest{
				Login:    "",
				Password: "securePassword123",
			},
			expectValid: false,
		},
		{
			name: "empty password",
			request: LoginRequest{
				Login:    "user@example.com",
				Password: "",
			},
			expectValid: false,
		},
		{
			name: "both fields empty",
			request: LoginRequest{
				Login:    "",
				Password: "",
			},
			expectValid: false,
		},
		{
			name:     "valid JSON unmarshaling",
			jsonData: `{"login":"admin@test.com","password":"adminPass"}`,
			validateFunc: func(t *testing.T, req LoginRequest) {
				t.Helper()
				assert.Equal(t, "admin@test.com", req.Login)
				assert.Equal(t, "adminPass", req.Password)
			},
		},
		{
			name:     "JSON with extra fields ignored",
			jsonData: `{"login":"admin@test.com","password":"adminPass","remember_me":true}`,
			validateFunc: func(t *testing.T, req LoginRequest) {
				t.Helper()
				assert.Equal(t, "admin@test.com", req.Login)
				assert.Equal(t, "adminPass", req.Password)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.jsonData != "" {
				var req LoginRequest
				err := json.Unmarshal([]byte(tt.jsonData), &req)
				require.NoError(t, err)
				tt.validateFunc(t, req)
			} else if tt.validateFunc != nil {
				tt.validateFunc(t, tt.request)
			}

			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.request)
			require.NoError(t, err)
			assert.Contains(t, string(jsonBytes), `"login"`)
			assert.Contains(t, string(jsonBytes), `"password"`)
		})
	}
}

func TestRegisterResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(t *testing.T, resp RegisterResponse)
		name         string
		jsonData     string
		response     RegisterResponse
	}{
		{
			name: "valid register response",
			response: RegisterResponse{
				ID: uuid.New(),
			},
			validateFunc: func(t *testing.T, resp RegisterResponse) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, resp.ID)
			},
		},
		{
			name: "nil UUID",
			response: RegisterResponse{
				ID: uuid.Nil,
			},
			validateFunc: func(t *testing.T, resp RegisterResponse) {
				t.Helper()
				assert.Equal(t, uuid.Nil, resp.ID)
			},
		},
		{
			name:     "valid JSON unmarshaling",
			jsonData: `{"id":"123e4567-e89b-12d3-a456-426614174000"}`,
			validateFunc: func(t *testing.T, resp RegisterResponse) {
				t.Helper()
				expectedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				assert.Equal(t, expectedUUID, resp.ID)
			},
		},
		{
			name:     "JSON with extra fields ignored",
			jsonData: `{"id":"123e4567-e89b-12d3-a456-426614174000","extra":"ignored"}`,
			validateFunc: func(t *testing.T, resp RegisterResponse) {
				t.Helper()
				expectedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				assert.Equal(t, expectedUUID, resp.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.jsonData != "" {
				var resp RegisterResponse
				err := json.Unmarshal([]byte(tt.jsonData), &resp)
				require.NoError(t, err)
				tt.validateFunc(t, resp)
			} else {
				tt.validateFunc(t, tt.response)
			}

			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.response)
			require.NoError(t, err)
			assert.Contains(t, string(jsonBytes), `"id"`)
		})
	}
}

func TestAccessToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(t *testing.T, token AccessToken)
		token        AccessToken
		name         string
		jsonData     string
	}{
		{
			name: "valid access token",
			token: AccessToken{
				AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				ExpiresAt:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
				TokenType:   "Bearer",
			},
			validateFunc: func(t *testing.T, token AccessToken) {
				t.Helper()
				assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", token.AccessToken)
				assert.Equal(t, "Bearer", token.TokenType)
				assert.Equal(t, 2023, token.ExpiresAt.Year())
			},
		},
		{
			name: "empty access token",
			token: AccessToken{
				AccessToken: "",
				ExpiresAt:   time.Time{},
				TokenType:   "",
			},
			validateFunc: func(t *testing.T, token AccessToken) {
				t.Helper()
				assert.Empty(t, token.AccessToken)
				assert.Empty(t, token.TokenType)
				assert.True(t, token.ExpiresAt.IsZero())
			},
		},
		{
			name: "bearer token type",
			token: AccessToken{
				AccessToken: "abc123",
				ExpiresAt:   time.Now().Add(time.Hour),
				TokenType:   "Bearer",
			},
			validateFunc: func(t *testing.T, token AccessToken) {
				t.Helper()
				assert.Equal(t, "abc123", token.AccessToken)
				assert.Equal(t, "Bearer", token.TokenType)
				assert.False(t, token.ExpiresAt.IsZero())
			},
		},
		{
			name:     "valid JSON unmarshaling",
			jsonData: `{"access_token":"test-token","expires_at":"2023-12-31T23:59:59Z","token_type":"Bearer"}`,
			validateFunc: func(t *testing.T, token AccessToken) {
				t.Helper()
				assert.Equal(t, "test-token", token.AccessToken)
				assert.Equal(t, "Bearer", token.TokenType)
				expectedTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
				assert.Equal(t, expectedTime, token.ExpiresAt)
			},
		},
		{
			name: "JSON with extra fields ignored",
			jsonData: `{"access_token":"test-token","expires_at":"2023-12-31T23:59:59Z",` +
				`"token_type":"Bearer","extra":"ignored"}`,
			validateFunc: func(t *testing.T, token AccessToken) {
				t.Helper()
				assert.Equal(t, "test-token", token.AccessToken)
				assert.Equal(t, "Bearer", token.TokenType)
				expectedTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
				assert.Equal(t, expectedTime, token.ExpiresAt)
			},
		},
		{
			name: "current time handling",
			jsonData: func() string {
				now := time.Now().UTC().Truncate(time.Second)
				return `{"access_token":"current-token","expires_at":"` + now.Format(
					time.RFC3339,
				) + `","token_type":"Bearer"}`
			}(),
			validateFunc: func(t *testing.T, token AccessToken) {
				t.Helper()
				assert.Equal(t, "current-token", token.AccessToken)
				assert.Equal(t, "Bearer", token.TokenType)
				assert.False(t, token.ExpiresAt.IsZero())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.jsonData != "" {
				var token AccessToken
				err := json.Unmarshal([]byte(tt.jsonData), &token)
				require.NoError(t, err)
				tt.validateFunc(t, token)
			} else {
				tt.validateFunc(t, tt.token)
			}

			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.token)
			require.NoError(t, err)
			assert.Contains(t, string(jsonBytes), `"access_token"`)
			assert.Contains(t, string(jsonBytes), `"expires_at"`)
			assert.Contains(t, string(jsonBytes), `"token_type"`)
		})
	}
}

func TestAccessToken_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := AccessToken{
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwi" +
			"YXQiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		ExpiresAt: time.Date(2024, 12, 31, 23, 59, 59, 123456789, time.UTC),
		TokenType: "Bearer",
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back to struct
	var unmarshaled AccessToken
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	// Verify fields match (time may have different precision)
	assert.Equal(t, original.AccessToken, unmarshaled.AccessToken)
	assert.Equal(t, original.TokenType, unmarshaled.TokenType)
	assert.Equal(t, original.ExpiresAt.Unix(), unmarshaled.ExpiresAt.Unix())
}
