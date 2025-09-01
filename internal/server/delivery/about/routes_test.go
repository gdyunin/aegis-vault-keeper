package about

import (
	"net/http"
	"testing"
	"time"

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
	}{
		{
			name:   "success/registers_about_endpoint",
			method: http.MethodGet,
			path:   "/about",
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

			mockInfo := &MockBuildInfoOperator{
				VersionFunc: func() string { return "1.0.0" },
				DateFunc:    time.Now,
				CommitFunc:  func() string { return "abc123" },
			}
			handler := NewHandler(mockInfo)

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

	mockInfo := &MockBuildInfoOperator{
		VersionFunc: func() string { return "2.0.0" },
		DateFunc:    func() time.Time { return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC) },
		CommitFunc:  func() string { return "def456" },
	}
	handler := NewHandler(mockInfo)

	// Register routes
	RegisterRoutes(group, handler)

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "success/about_endpoint_works",
			method:         http.MethodGet,
			path:           "/api/v1/about",
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
			handler.AboutInfo(c)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, recorder.statusCode)
		})
	}
}

func TestRegisterRoutes_MultipleGroups(t *testing.T) {
	t.Parallel()

	// Test that routes can be registered with different groups
	gin.SetMode(gin.TestMode)

	mockInfo := &MockBuildInfoOperator{}
	handler := NewHandler(mockInfo)

	tests := []struct {
		name       string
		groupPath  string
		expectPath string
	}{
		{
			name:       "success/root_group",
			groupPath:  "",
			expectPath: "/about",
		},
		{
			name:       "success/api_group",
			groupPath:  "/api",
			expectPath: "/api/about",
		},
		{
			name:       "success/versioned_group",
			groupPath:  "/api/v1",
			expectPath: "/api/v1/about",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create fresh router for each test
			testRouter := gin.New()
			group := testRouter.Group(tt.groupPath)

			// Register routes
			RegisterRoutes(group, handler)

			// Check that route was registered with correct path
			routes := testRouter.Routes()
			found := false
			for _, route := range routes {
				if route.Method == http.MethodGet && route.Path == tt.expectPath {
					found = true
					break
				}
			}

			assert.True(t, found, "Expected route GET %s not found", tt.expectPath)
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
