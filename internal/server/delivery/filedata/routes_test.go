package filedata

import (
	"context"
	"errors"
	"testing"

	appfiledata "github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Simple mock for route testing.
type mockService struct{}

func (m *mockService) Pull(
	ctx context.Context,
	params appfiledata.PullParams,
) (*appfiledata.FileData, error) {
	return nil, errors.New("mock error")
}

func (m *mockService) List(
	ctx context.Context,
	params appfiledata.ListParams,
) ([]*appfiledata.FileData, error) {
	return nil, nil
}
func (m *mockService) Push(ctx context.Context, params *appfiledata.PushParams) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		setupHandler func() *Handler
		verifyRoutes func(*testing.T, *gin.Engine)
		name         string
	}{
		{
			name: "routes registered successfully",
			setupHandler: func() *Handler {
				return &Handler{s: &mockService{}}
			},
			verifyRoutes: func(t *testing.T, router *gin.Engine) {
				t.Helper()
				routes := router.Routes()

				// Check that routes exist with correct methods and paths
				routeMap := make(map[string]bool)
				for _, route := range routes {
					key := route.Method + " " + route.Path
					routeMap[key] = true
				}

				// Verify essential routes are registered
				assert.True(t, routeMap["GET /filedata/:id"], "GET /:id route should be registered")
				assert.True(t, routeMap["GET /filedata/"], "GET / route should be registered")
				assert.True(t, routeMap["POST /filedata/"], "POST / route should be registered")
				assert.True(t, routeMap["PUT /filedata/:id"], "PUT /:id route should be registered")

				// Verify we have at least the expected number of routes
				assert.GreaterOrEqual(t, len(routes), 4, "Should have at least 4 routes registered")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := tt.setupHandler()
			router := gin.New()

			// Call the actual RegisterRoutes function
			RegisterRoutes(router, handler)

			tt.verifyRoutes(t, router)
		})
	}
}
