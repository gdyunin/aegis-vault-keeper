package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthService is a mock implementation of the Service interface for testing.
type mockAuthService struct {
	registerFunc func(context.Context, auth.RegisterParams) (uuid.UUID, error)
	loginFunc    func(context.Context, auth.LoginParams) (auth.AccessToken, error)
}

func (m *mockAuthService) Register(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, params)
	}
	return uuid.Nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, params)
	}
	return auth.AccessToken{}, nil
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	service := &mockAuthService{}
	handler := NewHandler(service)

	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.s)
}

func TestHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		requestBody    interface{}
		mockSetup      func(*mockAuthService)
		validateResp   func(t *testing.T, body []byte)
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				testID := uuid.New()
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					assert.Equal(t, "test@example.com", params.Login)
					assert.Equal(t, "securePassword123", params.Password)
					return testID, nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp RegisterResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, resp.ID)
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: `{"login": "test@example.com", "password":`,
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					t.Error("service should not be called with invalid JSON")
					return uuid.Nil, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "missing login field",
			requestBody: RegisterRequest{
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					t.Error("service should not be called with missing login")
					return uuid.Nil, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "missing password field",
			requestBody: RegisterRequest{
				Login: "test@example.com",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					t.Error("service should not be called with missing password")
					return uuid.Nil, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "user already exists error",
			requestBody: RegisterRequest{
				Login:    "existing@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					return uuid.Nil, auth.ErrAuthUserAlreadyExists
				}
			},
			expectedStatus: http.StatusConflict,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "User with this login already exists")
			},
		},
		{
			name: "incorrect login error",
			requestBody: RegisterRequest{
				Login:    "invalid-login",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					return uuid.Nil, auth.ErrAuthIncorrectLogin
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The login provided is not valid")
			},
		},
		{
			name: "incorrect password error",
			requestBody: RegisterRequest{
				Login:    "test@example.com",
				Password: "weak",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					return uuid.Nil, auth.ErrAuthIncorrectPassword
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The password provided is not valid")
			},
		},
		{
			name: "internal server error",
			requestBody: RegisterRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					return uuid.Nil, auth.ErrAuthTechError
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Internal Server Error")
			},
		},
		{
			name: "generic application error",
			requestBody: RegisterRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					return uuid.Nil, auth.ErrAuthAppError
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The parameters provided are invalid")
			},
		},
		{
			name:        "invalid content type",
			requestBody: "not-json",
			contentType: "text/plain",
			mockSetup: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, params auth.RegisterParams) (uuid.UUID, error) {
					t.Error("service should not be called with invalid content type")
					return uuid.Nil, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			mockService := &mockAuthService{}
			tt.mockSetup(mockService)
			handler := NewHandler(mockService)

			// Create request
			var bodyReader *bytes.Reader
			if str, ok := tt.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(str))
			} else {
				bodyBytes, err := json.Marshal(tt.requestBody)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bodyReader)
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(rec)
			c.Request = req

			// Execute
			handler.Register(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.validateResp(t, rec.Body.Bytes())
		})
	}
}

func TestHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		requestBody    interface{}
		mockSetup      func(*mockAuthService)
		validateResp   func(t *testing.T, body []byte)
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				expiresAt := time.Now().Add(time.Hour)
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					assert.Equal(t, "test@example.com", params.Login)
					assert.Equal(t, "securePassword123", params.Password)
					return auth.AccessToken{
						AccessToken: "test-jwt-token",
						ExpiresAt:   expiresAt,
						TokenType:   "Bearer",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp AccessToken
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, "test-jwt-token", resp.AccessToken)
				assert.Equal(t, "Bearer", resp.TokenType)
				assert.False(t, resp.ExpiresAt.IsZero())
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: `{"login": "test@example.com", "password":`,
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					t.Error("service should not be called with invalid JSON")
					return auth.AccessToken{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "missing login field",
			requestBody: LoginRequest{
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					t.Error("service should not be called with missing login")
					return auth.AccessToken{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "missing password field",
			requestBody: LoginRequest{
				Login: "test@example.com",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					t.Error("service should not be called with missing password")
					return auth.AccessToken{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "wrong login or password error",
			requestBody: LoginRequest{
				Login:    "wrong@example.com",
				Password: "wrongPassword",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, auth.ErrAuthWrongLoginOrPassword
				}
			},
			expectedStatus: http.StatusUnauthorized,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The provided login or password is incorrect")
			},
		},
		{
			name: "invalid access token error",
			requestBody: LoginRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, auth.ErrAuthInvalidAccessToken
				}
			},
			expectedStatus: http.StatusUnauthorized,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Your access token is invalid or has expired. Please log in")
			},
		},
		{
			name: "incorrect login validation error",
			requestBody: LoginRequest{
				Login:    "invalid-login",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, auth.ErrAuthIncorrectLogin
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The login provided is not valid")
			},
		},
		{
			name: "incorrect password validation error",
			requestBody: LoginRequest{
				Login:    "test@example.com",
				Password: "weak",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, auth.ErrAuthIncorrectPassword
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The password provided is not valid")
			},
		},
		{
			name: "internal server error",
			requestBody: LoginRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, auth.ErrAuthTechError
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Internal Server Error")
			},
		},
		{
			name: "generic application error",
			requestBody: LoginRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, auth.ErrAuthAppError
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "The parameters provided are invalid")
			},
		},
		{
			name:        "invalid content type",
			requestBody: "not-json",
			contentType: "text/plain",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					t.Error("service should not be called with invalid content type")
					return auth.AccessToken{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Bad Request")
			},
		},
		{
			name: "unknown error handling",
			requestBody: LoginRequest{
				Login:    "test@example.com",
				Password: "securePassword123",
			},
			contentType: "application/json",
			mockSetup: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, params auth.LoginParams) (auth.AccessToken, error) {
					return auth.AccessToken{}, errors.New("unknown error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "Internal Server Error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			mockService := &mockAuthService{}
			tt.mockSetup(mockService)
			handler := NewHandler(mockService)

			// Create request
			var bodyReader *bytes.Reader
			if str, ok := tt.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(str))
			} else {
				bodyBytes, err := json.Marshal(tt.requestBody)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bodyReader)
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(rec)
			c.Request = req

			// Execute
			handler.Login(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.validateResp(t, rec.Body.Bytes())
		})
	}
}
