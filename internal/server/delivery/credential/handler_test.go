package credential

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockService implements the Service interface for testing.
type mockService struct {
	pullFunc func(ctx context.Context, params credential.PullParams) (*credential.Credential, error)
	listFunc func(ctx context.Context, params credential.ListParams) ([]*credential.Credential, error)
	pushFunc func(ctx context.Context, params *credential.PushParams) (uuid.UUID, error)
}

func (m *mockService) Pull(
	ctx context.Context,
	params credential.PullParams,
) (*credential.Credential, error) {
	if m.pullFunc != nil {
		return m.pullFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) List(
	ctx context.Context,
	params credential.ListParams,
) ([]*credential.Credential, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) Push(ctx context.Context, params *credential.PushParams) (uuid.UUID, error) {
	if m.pushFunc != nil {
		return m.pushFunc(ctx, params)
	}
	return uuid.Nil, errors.New("not implemented")
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	service := &mockService{}
	handler := NewHandler(service)

	require.NotNil(t, handler)
	assert.Equal(t, service, handler.s)
}

func TestHandler_Pull(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	credID := uuid.New()

	tests := []struct {
		expectedBody   interface{}
		setupContext   func(c *gin.Context)
		mockSetup      func(m *mockService)
		name           string
		urlParam       string
		expectedStatus int
	}{
		{
			name: "successful pull",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			urlParam: credID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params credential.PullParams) (*credential.Credential, error) {
					assert.Equal(t, credID, params.ID)
					assert.Equal(t, userID, params.UserID)
					return &credential.Credential{
						ID:          credID,
						UserID:      userID,
						Login:       "user@example.com",
						Password:    "password123",
						Description: "Test credential",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody: PullResponse{
				Credential: &Credential{
					ID:          credID,
					Login:       "user@example.com",
					Password:    "password123",
					Description: "Test credential",
				},
			},
		},
		{
			name: "missing user ID",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			urlParam:       credID.String(),
			mockSetup:      func(m *mockService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.DefaultInternalServerError,
		},
		{
			name: "invalid UUID in path",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			urlParam:       "invalid-uuid",
			mockSetup:      func(m *mockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.DefaultBadRequestError,
		},
		{
			name: "service returns not found error",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			urlParam: credID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params credential.PullParams) (*credential.Credential, error) {
					return nil, credential.ErrCredentialNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: response.Error{
				Messages: []string{"Credential not found"},
			},
		},
		{
			name: "service returns access denied error",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			urlParam: credID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params credential.PullParams) (*credential.Credential, error) {
					return nil, credential.ErrCredentialAccessDenied
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: response.Error{
				Messages: []string{"Access to this credential is denied"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := &mockService{}
			tt.mockSetup(mockSvc)
			handler := NewHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			req := httptest.NewRequest(http.MethodGet, "/credentials/"+tt.urlParam, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.urlParam}}

			// Mock the extractor behavior for binding URI
			if tt.urlParam != "" && tt.urlParam != "invalid-uuid" {
				// Valid UUID case - the BindURI will succeed
				c.Request.URL.Path = "/credentials/" + tt.urlParam
			}

			tt.setupContext(c)

			handler.Pull(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var actualBody interface{}
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err)

				expectedBytes, err := json.Marshal(tt.expectedBody)
				require.NoError(t, err)

				var expectedBody interface{}
				err = json.Unmarshal(expectedBytes, &expectedBody)
				require.NoError(t, err)

				assert.Equal(t, expectedBody, actualBody)
			}
		})
	}
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		expectedBody   interface{}
		setupContext   func(c *gin.Context)
		mockSetup      func(m *mockService)
		name           string
		expectedStatus int
	}{
		{
			name: "successful list with credentials",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			mockSetup: func(m *mockService) {
				m.listFunc = func(ctx context.Context, params credential.ListParams) ([]*credential.Credential, error) {
					assert.Equal(t, userID, params.UserID)
					return []*credential.Credential{
						{
							ID:          uuid.New(),
							UserID:      userID,
							Login:       "user1@example.com",
							Password:    "password1",
							Description: "Test credential 1",
						},
						{
							ID:          uuid.New(),
							UserID:      userID,
							Login:       "user2@example.com",
							Password:    "password2",
							Description: "Test credential 2",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful list with empty result",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			mockSetup: func(m *mockService) {
				m.listFunc = func(ctx context.Context, params credential.ListParams) ([]*credential.Credential, error) {
					return []*credential.Credential{}, nil
				}
			},
			expectedStatus: http.StatusOK, // Gin returns 200 even when c.Status(204) is called
			expectedBody:   nil,
		},
		{
			name: "missing user ID",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			mockSetup:      func(m *mockService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.DefaultInternalServerError,
		},
		{
			name: "service returns error",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			mockSetup: func(m *mockService) {
				m.listFunc = func(ctx context.Context, params credential.ListParams) ([]*credential.Credential, error) {
					return nil, credential.ErrCredentialTechError
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: response.Error{
				Messages: []string{"Internal Server Error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := &mockService{}
			tt.mockSetup(mockSvc)
			handler := NewHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
			c.Request = req

			tt.setupContext(c)

			handler.List(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var actualBody interface{}
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err)

				expectedBytes, err := json.Marshal(tt.expectedBody)
				require.NoError(t, err)

				var expectedBody interface{}
				err = json.Unmarshal(expectedBytes, &expectedBody)
				require.NoError(t, err)

				assert.Equal(t, expectedBody, actualBody)
			}
		})
	}
}

func TestHandler_Push(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	credID := uuid.New()
	newCredID := uuid.New()

	tests := []struct {
		requestBody    interface{}
		expectedBody   interface{}
		setupContext   func(c *gin.Context)
		mockSetup      func(m *mockService)
		name           string
		urlParam       string
		expectedStatus int
	}{
		{
			name: "successful create (POST)",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody: PushRequest{
				Login:       "user@example.com",
				Password:    "password123",
				Description: "Test credential",
			},
			urlParam: "",
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *credential.PushParams) (uuid.UUID, error) {
					assert.Equal(t, uuid.Nil, params.ID)
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, "user@example.com", params.Login)
					assert.Equal(t, "password123", params.Password)
					assert.Equal(t, "Test credential", params.Description)
					return newCredID, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: PushResponse{
				ID: newCredID,
			},
		},
		{
			name: "successful update (PUT)",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody: PushRequest{
				Login:       "updated@example.com",
				Password:    "newpassword123",
				Description: "Updated credential",
			},
			urlParam: credID.String(),
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *credential.PushParams) (uuid.UUID, error) {
					assert.Equal(t, credID, params.ID)
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, "updated@example.com", params.Login)
					assert.Equal(t, "newpassword123", params.Password)
					assert.Equal(t, "Updated credential", params.Description)
					return credID, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: PushResponse{
				ID: credID,
			},
		},
		{
			name: "missing user ID",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			requestBody: PushRequest{
				Login:    "user@example.com",
				Password: "password123",
			},
			urlParam:       "",
			mockSetup:      func(m *mockService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.DefaultInternalServerError,
		},
		{
			name: "invalid JSON body",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody:    "invalid json",
			urlParam:       "",
			mockSetup:      func(m *mockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.DefaultBadRequestError,
		},
		{
			name: "invalid UUID in path",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody: PushRequest{
				Login:    "user@example.com",
				Password: "password123",
			},
			urlParam:       "invalid-uuid",
			mockSetup:      func(m *mockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.DefaultBadRequestError,
		},
		{
			name: "service returns validation error",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody: PushRequest{
				Login:    "valid@email.com", // Make valid to pass JSON binding
				Password: "password123",
			},
			urlParam: "",
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *credential.PushParams) (uuid.UUID, error) {
					return uuid.Nil, credential.ErrCredentialIncorrectLogin
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: response.Error{
				Messages: []string{"Invalid login"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := &mockService{}
			tt.mockSetup(mockSvc)
			handler := NewHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Prepare request body
			var bodyReader *bytes.Reader
			if tt.requestBody != nil {
				bodyBytes, err := json.Marshal(tt.requestBody)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}

			url := "/credentials"
			if tt.urlParam != "" {
				url = fmt.Sprintf("/credentials/%s", tt.urlParam)
				c.Params = gin.Params{{Key: "id", Value: tt.urlParam}}
			}

			req := httptest.NewRequest(http.MethodPost, url, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			tt.setupContext(c)

			handler.Push(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var actualBody interface{}
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err)

				expectedBytes, err := json.Marshal(tt.expectedBody)
				require.NoError(t, err)

				var expectedBody interface{}
				err = json.Unmarshal(expectedBytes, &expectedBody)
				require.NoError(t, err)

				assert.Equal(t, expectedBody, actualBody)
			}
		})
	}
}
