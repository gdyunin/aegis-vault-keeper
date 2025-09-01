package filedata

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFileDataService implements filedata service for testing.
type mockFileDataService struct {
	pullFunc func(ctx context.Context, params filedata.PullParams) (*filedata.FileData, error)
	listFunc func(ctx context.Context, params filedata.ListParams) ([]*filedata.FileData, error)
	pushFunc func(ctx context.Context, params *filedata.PushParams) (uuid.UUID, error)
}

func (m *mockFileDataService) Pull(
	ctx context.Context,
	params filedata.PullParams,
) (*filedata.FileData, error) {
	if m.pullFunc != nil {
		return m.pullFunc(ctx, params)
	}
	return &filedata.FileData{}, nil
}

func (m *mockFileDataService) List(
	ctx context.Context,
	params filedata.ListParams,
) ([]*filedata.FileData, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, params)
	}
	return []*filedata.FileData{}, nil
}

func (m *mockFileDataService) Push(ctx context.Context, params *filedata.PushParams) (uuid.UUID, error) {
	if m.pushFunc != nil {
		return m.pushFunc(ctx, params)
	}
	return uuid.New(), nil
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		svc  Service
		name string
	}{
		{
			name: "creates handler successfully",
			svc:  &mockFileDataService{},
		},
		{
			name: "creates handler with nil service",
			svc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(tt.svc)

			assert.NotNil(t, handler)
			assert.Equal(t, tt.svc, handler.s)
		})
	}
}

func TestHandler_Pull(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()
	testData := []byte("test file content")

	tests := []struct {
		setupContext   func(c *gin.Context)
		mockService    *mockFileDataService
		validateResp   func(t *testing.T, w *httptest.ResponseRecorder)
		name           string
		expectedStatus int
	}{
		{
			name: "successful pull",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
				c.Params = []gin.Param{{Key: "id", Value: fileID.String()}}
			},
			mockService: &mockFileDataService{
				pullFunc: func(ctx context.Context, params filedata.PullParams) (*filedata.FileData, error) {
					assert.Equal(t, fileID, params.ID)
					assert.Equal(t, userID, params.UserID)
					return &filedata.FileData{
						ID:          fileID,
						UserID:      userID,
						StorageKey:  "test.txt",
						Description: "Test file",
						Data:        testData,
						UpdatedAt:   time.Now(),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Contains(t, w.Header().Get("Content-Type"), "multipart/form-data")
				assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
				assert.Greater(t, len(w.Body.Bytes()), 0)
			},
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				c.Params = []gin.Param{{Key: "id", Value: fileID.String()}}
			},
			mockService:    &mockFileDataService{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid file ID",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
				c.Params = []gin.Param{{Key: "id", Value: "invalid-uuid"}}
			},
			mockService:    &mockFileDataService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
				c.Params = []gin.Param{{Key: "id", Value: fileID.String()}}
			},
			mockService: &mockFileDataService{
				pullFunc: func(ctx context.Context, params filedata.PullParams) (*filedata.FileData, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			handler := NewHandler(tt.mockService)
			handler.Pull(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w)
			}
		})
	}
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		setupContext   func(c *gin.Context)
		mockService    *mockFileDataService
		validateResp   func(t *testing.T, body []byte)
		name           string
		expectedStatus int
	}{
		{
			name: "successful list with files",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			mockService: &mockFileDataService{
				listFunc: func(ctx context.Context, params filedata.ListParams) ([]*filedata.FileData, error) {
					assert.Equal(t, userID, params.UserID)
					return []*filedata.FileData{
						{
							ID:          uuid.New(),
							UserID:      userID,
							StorageKey:  "test1.txt",
							Description: "Test file 1",
							UpdatedAt:   time.Now(),
						},
						{
							ID:          uuid.New(),
							UserID:      userID,
							StorageKey:  "test2.txt",
							Description: "Test file 2",
							UpdatedAt:   time.Now(),
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), "test1.txt")
				assert.Contains(t, string(body), "test2.txt")
			},
		},
		{
			name: "successful list with empty result",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			mockService: &mockFileDataService{
				listFunc: func(ctx context.Context, params filedata.ListParams) ([]*filedata.FileData, error) {
					return []*filedata.FileData{}, nil
				},
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			mockService:    &mockFileDataService{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "service error",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			mockService: &mockFileDataService{
				listFunc: func(ctx context.Context, params filedata.ListParams) ([]*filedata.FileData, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			handler := NewHandler(tt.mockService)
			handler.List(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandler_Push(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fileID := uuid.New()
	testContent := []byte("test file content")

	tests := []struct {
		setupContext   func(c *gin.Context)
		setupRequest   func() *http.Request
		mockService    *mockFileDataService
		validateResp   func(t *testing.T, body []byte)
		name           string
		expectedStatus int
	}{
		{
			name: "successful push new file",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Add file field
				part, err := writer.CreateFormFile("file", "test.txt")
				require.NoError(t, err)
				_, err = part.Write(testContent)
				require.NoError(t, err)

				// Add storage_key field
				err = writer.WriteField("storage_key", "custom_name.txt")
				require.NoError(t, err)

				// Add description field
				err = writer.WriteField("description", "Test description")
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/items/filedata", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			mockService: &mockFileDataService{
				pushFunc: func(ctx context.Context, params *filedata.PushParams) (uuid.UUID, error) {
					assert.Equal(t, userID, params.UserID)
					assert.Equal(t, uuid.Nil, params.ID) // New file
					assert.Equal(t, "custom_name.txt", params.StorageKey)
					assert.Equal(t, "Test description", params.Description)
					assert.Equal(t, testContent, params.Data)
					return fileID, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validateResp: func(t *testing.T, body []byte) {
				t.Helper()
				assert.Contains(t, string(body), fileID.String())
			},
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/items/filedata", strings.NewReader(""))
			},
			mockService:    &mockFileDataService{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "missing file",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				err := writer.WriteField("description", "No file field")
				require.NoError(t, err)
				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/items/filedata", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			mockService:    &mockFileDataService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				part, err := writer.CreateFormFile("file", "test.txt")
				require.NoError(t, err)
				_, err = part.Write(testContent)
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/items/filedata", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			mockService: &mockFileDataService{
				pushFunc: func(ctx context.Context, params *filedata.PushParams) (uuid.UUID, error) {
					return uuid.Nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = tt.setupRequest()

			tt.setupContext(c)

			handler := NewHandler(tt.mockService)
			handler.Push(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}
