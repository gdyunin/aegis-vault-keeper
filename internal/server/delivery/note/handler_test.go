package note

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockService implements the Service interface for testing.
type mockService struct {
	pullFunc func(ctx context.Context, params note.PullParams) (*note.Note, error)
	listFunc func(ctx context.Context, params note.ListParams) ([]*note.Note, error)
	pushFunc func(ctx context.Context, params *note.PushParams) (uuid.UUID, error)
}

func (m *mockService) Pull(ctx context.Context, params note.PullParams) (*note.Note, error) {
	if m.pullFunc != nil {
		return m.pullFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) List(ctx context.Context, params note.ListParams) ([]*note.Note, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) Push(ctx context.Context, params *note.PushParams) (uuid.UUID, error) {
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
	noteID := uuid.New()

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
			urlParam: noteID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params note.PullParams) (*note.Note, error) {
					assert.Equal(t, noteID, params.ID)
					assert.Equal(t, userID, params.UserID)
					return &note.Note{
						ID:          noteID,
						UserID:      userID,
						Note:        "Important meeting notes",
						Description: "Meeting with client ABC",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody: PullResponse{
				Note: &Note{
					ID:          noteID,
					Note:        "Important meeting notes",
					Description: "Meeting with client ABC",
				},
			},
		},
		{
			name: "missing user ID",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			urlParam:       noteID.String(),
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
			urlParam: noteID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params note.PullParams) (*note.Note, error) {
					return nil, note.ErrNoteNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: response.Error{
				Messages: []string{"Note not found"},
			},
		},
		{
			name: "service returns access denied error",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			urlParam: noteID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params note.PullParams) (*note.Note, error) {
					return nil, note.ErrNoteAccessDenied
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: response.Error{
				Messages: []string{"Access to this note is denied"},
			},
		},
		{
			name: "service returns tech error",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			urlParam: noteID.String(),
			mockSetup: func(m *mockService) {
				m.pullFunc = func(ctx context.Context, params note.PullParams) (*note.Note, error) {
					return nil, note.ErrNoteTechError
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

			// Setup request
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.urlParam, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.urlParam}}

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
			name: "successful list with notes",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			mockSetup: func(m *mockService) {
				m.listFunc = func(ctx context.Context, params note.ListParams) ([]*note.Note, error) {
					assert.Equal(t, userID, params.UserID)
					return []*note.Note{
						{
							ID:          uuid.New(),
							UserID:      userID,
							Note:        "First note",
							Description: "First description",
						},
						{
							ID:          uuid.New(),
							UserID:      userID,
							Note:        "Second note",
							Description: "Second description",
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
				m.listFunc = func(ctx context.Context, params note.ListParams) ([]*note.Note, error) {
					return []*note.Note{}, nil // Return empty slice explicitly
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
				m.listFunc = func(ctx context.Context, params note.ListParams) ([]*note.Note, error) {
					return nil, note.ErrNoteTechError
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

			req := httptest.NewRequest(http.MethodGet, "/notes", nil)
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
	noteID := uuid.New()
	newNoteID := uuid.New()

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
				Note:        "Important meeting notes",
				Description: "Meeting with client ABC",
			},
			urlParam: "",
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *note.PushParams) (uuid.UUID, error) {
					assert.Equal(t, uuid.Nil, params.ID)
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, "Important meeting notes", params.Note)
					assert.Equal(t, "Meeting with client ABC", params.Description)
					return newNoteID, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: PushResponse{
				ID: newNoteID,
			},
		},
		{
			name: "successful update (PUT)",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody: PushRequest{
				Note:        "Updated note content",
				Description: "Updated description",
			},
			urlParam: noteID.String(),
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *note.PushParams) (uuid.UUID, error) {
					assert.Equal(t, noteID, params.ID)
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, "Updated note content", params.Note)
					assert.Equal(t, "Updated description", params.Description)
					return noteID, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: PushResponse{
				ID: noteID,
			},
		},
		{
			name: "missing user ID",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			requestBody: PushRequest{
				Note: "Test note",
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
				Note: "Test note",
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
				Note: "Valid note content", // Make valid to pass JSON binding
			},
			urlParam: "",
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *note.PushParams) (uuid.UUID, error) {
					return uuid.Nil, note.ErrNoteIncorrectNoteText
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: response.Error{
				Messages: []string{"Invalid note text"},
			},
		},
		{
			name: "service returns not found error for update",
			setupContext: func(c *gin.Context) {
				c.Set("userID", userID)
			},
			requestBody: PushRequest{
				Note: "Test note",
			},
			urlParam: noteID.String(),
			mockSetup: func(m *mockService) {
				m.pushFunc = func(ctx context.Context, params *note.PushParams) (uuid.UUID, error) {
					return uuid.Nil, note.ErrNoteNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: response.Error{
				Messages: []string{"Note not found"},
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

			url := "/notes"
			if tt.urlParam != "" {
				url = fmt.Sprintf("/notes/%s", tt.urlParam)
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
