package middleware

import (
	"errors"
	"testing"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareErrRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrorIn    error
		name           string
		wantPublicMsg  string
		registryRule   errutil.Rule
		wantStatusCode int
		wantErrorClass errutil.ErrorClass
		wantLogIt      bool
		wantAllowMerge bool
	}{
		{
			name: "success/auth_invalid_access_token_registry",
			registryRule: errutil.Rule{
				ErrorIn: app.ErrAuthInvalidAccessToken,
				HandlePolicy: errutil.Policy{
					StatusCode: 401,
					PublicMsg:  "Your access token is invalid or has expired. Please log in",
					LogIt:      false,
					AllowMerge: false,
					ErrorClass: errutil.ErrorClassAuth,
				},
			},
			wantErrorIn:    app.ErrAuthInvalidAccessToken,
			wantStatusCode: 401,
			wantPublicMsg:  "Your access token is invalid or has expired. Please log in",
			wantLogIt:      false,
			wantAllowMerge: false,
			wantErrorClass: errutil.ErrorClassAuth,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Find the entry in the actual registry
			var foundEntry *errutil.Rule
			for _, entry := range MiddlewareErrRegistry {
				if entry.ErrorIn == tt.wantErrorIn {
					foundEntry = &entry
					break
				}
			}

			require.NotNil(t, foundEntry, "Registry entry should exist for error: %v", tt.wantErrorIn)

			// Verify registry entry fields
			assert.Equal(t, tt.wantErrorIn, foundEntry.ErrorIn)
			assert.Equal(t, tt.wantStatusCode, foundEntry.HandlePolicy.StatusCode)
			assert.Equal(t, tt.wantPublicMsg, foundEntry.HandlePolicy.PublicMsg)
			assert.Equal(t, tt.wantLogIt, foundEntry.HandlePolicy.LogIt)
			assert.Equal(t, tt.wantAllowMerge, foundEntry.HandlePolicy.AllowMerge)
			assert.Equal(t, tt.wantErrorClass, foundEntry.HandlePolicy.ErrorClass)
		})
	}
}

func TestMiddlewareErrRegistry_Coverage(t *testing.T) {
	t.Parallel()

	// Test that registry contains expected entries
	require.NotEmpty(t, MiddlewareErrRegistry, "Middleware error registry should not be empty")

	// Verify specific entries exist
	authTokenErrorFound := false
	for _, entry := range MiddlewareErrRegistry {
		if entry.ErrorIn == app.ErrAuthInvalidAccessToken {
			authTokenErrorFound = true
			break
		}
	}

	assert.True(t, authTokenErrorFound, "Registry should contain auth invalid token error")
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputError       error
		setupGinContext  func() *gin.Context
		name             string
		wantMessages     []string
		wantStatusCode   int
		wantEmptyMessage bool
	}{
		{
			name:       "success/auth_invalid_token_error",
			inputError: app.ErrAuthInvalidAccessToken,
			setupGinContext: func() *gin.Context {
				gin.SetMode(gin.TestMode)
				c, _ := gin.CreateTestContext(nil)
				return c
			},
			wantStatusCode: 401,
			wantMessages:   []string{"Your access token is invalid or has expired. Please log in"},
		},
		{
			name:       "success/unknown_error_handling",
			inputError: errors.New("generic test error"), // Generic error not in registry
			setupGinContext: func() *gin.Context {
				gin.SetMode(gin.TestMode)
				c, _ := gin.CreateTestContext(nil)
				return c
			},
			wantStatusCode:   500,  // Default for unregistered errors
			wantEmptyMessage: true, // Should return default error handling
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			c := tt.setupGinContext()

			// Execute
			statusCode, messages := handleError(tt.inputError, c)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, statusCode)

			if tt.wantEmptyMessage {
				// For unregistered errors, we expect some default message handling
				// The exact behavior depends on errutil.HandleWithRegistry implementation
				require.NotNil(t, messages)
			} else {
				require.Equal(t, tt.wantMessages, messages)
			}
		})
	}
}

func TestHandleError_Integration(t *testing.T) {
	t.Parallel()

	// Integration test to ensure handleError works with real gin context
	gin.SetMode(gin.TestMode)

	tests := []struct {
		inputError     error
		name           string
		wantStatusCode int
	}{
		{
			name:           "integration/auth_error",
			inputError:     app.ErrAuthInvalidAccessToken,
			wantStatusCode: 401,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create real gin context
			c, _ := gin.CreateTestContext(nil)

			// Execute
			statusCode, messages := handleError(tt.inputError, c)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, statusCode)
			assert.NotNil(t, messages, "Messages should not be nil")
		})
	}
}

func TestHandleError_NilError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test with nil error
	statusCode, messages := handleError(nil, c)

	// Should handle nil gracefully - exact behavior depends on errutil implementation
	// but should not panic
	assert.NotZero(t, statusCode, "Status code should be set")
	assert.NotNil(t, messages, "Messages should not be nil")
}
