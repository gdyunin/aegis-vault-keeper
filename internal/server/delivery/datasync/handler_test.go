package datasync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/datasync"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSyncService implements datasync service for testing.
type mockSyncService struct {
	pullFunc func(ctx context.Context, userID uuid.UUID) (*datasync.SyncPayload, error)
	pushFunc func(ctx context.Context, payload *datasync.SyncPayload) error
}

func (m *mockSyncService) Pull(ctx context.Context, userID uuid.UUID) (*datasync.SyncPayload, error) {
	if m.pullFunc != nil {
		return m.pullFunc(ctx, userID)
	}
	return &datasync.SyncPayload{}, nil
}

func (m *mockSyncService) Push(ctx context.Context, payload *datasync.SyncPayload) error {
	if m.pushFunc != nil {
		return m.pushFunc(ctx, payload)
	}
	return nil
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		svc  Service
		name string
	}{
		{
			name: "creates handler with service",
			svc:  nil, // Using nil for simplicity
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

func TestDataSyncErrRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedNonNil bool
	}{
		{
			name:           "registry is not nil",
			expectedNonNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectedNonNil {
				assert.NotNil(t, DataSyncErrRegistry)
			}
		})
	}
}

func TestHandler_Pull(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		setupContext   func(c *gin.Context)
		mockService    *mockSyncService
		name           string
		expectedStatus int
	}{
		{
			name: "successful pull",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			mockService: &mockSyncService{
				pullFunc: func(ctx context.Context, userID uuid.UUID) (*datasync.SyncPayload, error) {
					return &datasync.SyncPayload{}, nil
				},
			},
			expectedStatus: http.StatusNoContent, // Empty payload should return 204
		},
		{
			name: "successful pull with data",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			mockService: &mockSyncService{
				pullFunc: func(ctx context.Context, userID uuid.UUID) (*datasync.SyncPayload, error) {
					return &datasync.SyncPayload{
						UserID: userID,
						BankCards: []*bankcard.BankCard{
							{
								ID:          uuid.New(),
								UserID:      userID,
								CardNumber:  "4111111111111111",
								CVV:         "123",
								CardHolder:  "John Doe",
								ExpiryMonth: "12",
								ExpiryYear:  "2025",
								Description: "Test card",
							},
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK, // Non-empty payload should return 200
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				// don't set user_id
			},
			mockService:    &mockSyncService{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "service error",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			mockService: &mockSyncService{
				pullFunc: func(ctx context.Context, userID uuid.UUID) (*datasync.SyncPayload, error) {
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
		})
	}
}

func TestHandler_Push(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	payload := &SyncPayload{}

	tests := []struct {
		requestBody    interface{}
		setupContext   func(c *gin.Context)
		mockService    *mockSyncService
		name           string
		expectedStatus int
	}{
		{
			name: "successful push",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			requestBody: payload,
			mockService: &mockSyncService{
				pushFunc: func(ctx context.Context, payload *datasync.SyncPayload) error {
					return nil
				},
			},
			expectedStatus: http.StatusNoContent, // Push returns 204, not 200
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				// don't set user_id
			},
			requestBody:    payload,
			mockService:    &mockSyncService{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid JSON",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			requestBody:    "invalid json",
			mockService:    &mockSyncService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			setupContext: func(c *gin.Context) {
				c.Set(consts.CtxKeyUserID, userID)
			},
			requestBody: payload,
			mockService: &mockSyncService{
				pushFunc: func(ctx context.Context, payload *datasync.SyncPayload) error {
					return errors.New("service error")
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

			var reqBody []byte
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					reqBody = []byte(str)
				} else {
					var err error
					reqBody, err = json.Marshal(tt.requestBody)
					require.NoError(t, err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			tt.setupContext(c)

			handler := NewHandler(tt.mockService)
			handler.Push(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	tests := []struct {
		err          error
		name         string
		expectedCode int
	}{
		{
			name:         "generic error",
			err:          errors.New("generic error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, msgs := handleError(tt.err, c)

			assert.Equal(t, tt.expectedCode, code)
			assert.NotEmpty(t, msgs)
		})
	}
}
