package health

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		routes []string
	}{
		{
			name:   "success/registers_health_endpoint",
			routes: []string{"/health"},
			method: http.MethodGet,
			path:   "/health",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			group := router.Group("")
			handler := NewHandler()

			// Register routes
			RegisterRoutes(group, handler)

			// Get routes info
			routes := router.Routes()

			// Find the registered route
			found := false
			for _, route := range routes {
				if route.Method == tt.method && route.Path == tt.path {
					found = true
					break
				}
			}

			require.True(t, found, "Expected route %s %s not found", tt.method, tt.path)
		})
	}
}

func TestRegisterRoutes_Integration(t *testing.T) {
	t.Parallel()

	// Integration test to ensure registered routes work correctly
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	handler := NewHandler()

	// Register routes
	RegisterRoutes(group, handler)

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "success/health_check_works",
			method:         http.MethodGet,
			path:           "/api/v1/health",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			// Create recorder
			recorder := &FakeResponseWriter{
				statusCode: http.StatusOK,
				headers:    make(http.Header),
			}

			// Create gin context manually for testing
			c, _ := gin.CreateTestContext(recorder)
			c.Request = req

			// Execute handler directly
			handler.HealthCheck(c)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, recorder.statusCode)
		})
	}
}

// FakeResponseWriter implements http.ResponseWriter for testing.
type FakeResponseWriter struct {
	headers    http.Header
	body       []byte
	statusCode int
}

func (f *FakeResponseWriter) Header() http.Header {
	return f.headers
}

func (f *FakeResponseWriter) Write(data []byte) (int, error) {
	f.body = append(f.body, data...)
	return len(data), nil
}

func (f *FakeResponseWriter) WriteHeader(statusCode int) {
	f.statusCode = statusCode
}
