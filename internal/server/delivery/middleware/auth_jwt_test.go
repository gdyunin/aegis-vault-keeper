package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuthWithJWTService implements AuthWithJWTService interface for testing.
type MockAuthWithJWTService struct {
	ValidateTokenFunc func(token string) (uuid.UUID, error)
}

func (m *MockAuthWithJWTService) ValidateToken(token string) (uuid.UUID, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(token)
	}
	return uuid.New(), nil
}

func TestAuthWithJWT(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		setupRequest     func() *http.Request
		setupMockService func(*MockAuthWithJWTService)
		name             string
		wantResponseBody string
		wantStatusCode   int
		wantUserID       uuid.UUID
		wantAborted      bool
		wantUserIDInCtx  bool
	}{
		{
			name: "success/valid_token",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer valid_token_123")
				return req
			},
			setupMockService: func(m *MockAuthWithJWTService) {
				m.ValidateTokenFunc = func(token string) (uuid.UUID, error) {
					assert.Equal(t, "valid_token_123", token)
					return testUserID, nil
				}
			},
			wantStatusCode:  http.StatusOK,
			wantAborted:     false,
			wantUserIDInCtx: true,
			wantUserID:      testUserID,
		},
		{
			name: "error/missing_authorization_header",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				// No Authorization header
				return req
			},
			setupMockService: func(m *MockAuthWithJWTService) {},
			wantStatusCode:   http.StatusUnauthorized,
			wantAborted:      true,
			wantUserIDInCtx:  false,
		},
		{
			name: "error/empty_authorization_header",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "")
				return req
			},
			setupMockService: func(m *MockAuthWithJWTService) {},
			wantStatusCode:   http.StatusUnauthorized,
			wantAborted:      true,
			wantUserIDInCtx:  false,
		},
		{
			name: "success/bearer_token_without_prefix",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "raw_token_without_bearer")
				return req
			},
			setupMockService: func(m *MockAuthWithJWTService) {
				m.ValidateTokenFunc = func(token string) (uuid.UUID, error) {
					assert.Equal(t, "raw_token_without_bearer", token)
					return testUserID, nil
				}
			},
			wantStatusCode:  http.StatusOK,
			wantAborted:     false,
			wantUserIDInCtx: true,
			wantUserID:      testUserID,
		},
		{
			name: "error/invalid_token",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer invalid_token")
				return req
			},
			setupMockService: func(m *MockAuthWithJWTService) {
				m.ValidateTokenFunc = func(token string) (uuid.UUID, error) {
					return uuid.Nil, errors.New("invalid token")
				}
			},
			wantStatusCode:  http.StatusInternalServerError, // Based on error registry
			wantAborted:     true,
			wantUserIDInCtx: false,
		},
		{
			name: "success/bearer_prefix_trimmed",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer token_after_bearer")
				return req
			},
			setupMockService: func(m *MockAuthWithJWTService) {
				m.ValidateTokenFunc = func(token string) (uuid.UUID, error) {
					assert.Equal(t, "token_after_bearer", token)
					return testUserID, nil
				}
			},
			wantStatusCode:  http.StatusOK,
			wantAborted:     false,
			wantUserIDInCtx: true,
			wantUserID:      testUserID,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			mockService := &MockAuthWithJWTService{}
			if tt.setupMockService != nil {
				tt.setupMockService(mockService)
			}

			// Create router with middleware
			router := gin.New()
			router.Use(AuthWithJWT(mockService))

			// Add test endpoint that checks context
			router.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get(consts.CtxKeyUserID)
				if exists {
					c.JSON(http.StatusOK, gin.H{"user_id": userID})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no user in context"})
				}
			})

			// Create test request
			req := tt.setupRequest()
			recorder := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(recorder, req)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, recorder.Code)

			if tt.wantUserIDInCtx {
				// Check that the endpoint response contains the user ID
				assert.Contains(t, recorder.Body.String(), tt.wantUserID.String())
			}

			// Note: We can't directly check if c.Abort() was called because it's internal to gin,
			// but we can infer it from the response status and whether our test endpoint was reached
		})
	}
}

func TestAuthWithJWT_WithServer(t *testing.T) {
	t.Parallel()

	// Test using httptest.Server as recommended in Issue #16
	gin.SetMode(gin.TestMode)
	testUserID := uuid.New()

	mockService := &MockAuthWithJWTService{
		ValidateTokenFunc: func(token string) (uuid.UUID, error) {
			if token == "valid_token" {
				return testUserID, nil
			}
			return uuid.Nil, errors.New("invalid token")
		},
	}

	router := gin.New()
	router.Use(AuthWithJWT(mockService))
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get(consts.CtxKeyUserID)
		if exists {
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "no user"})
		}
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
	}{
		{
			name:           "success/valid_bearer_token",
			authHeader:     "Bearer valid_token",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "error/invalid_token",
			authHeader:     "Bearer invalid_token",
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "error/no_auth_header",
			authHeader:     "",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Make HTTP request to test server
			req, err := http.NewRequest(http.MethodGet, server.URL+"/protected", nil)
			require.NoError(t, err)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
