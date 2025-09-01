package bankcard

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
		setupHandler   func() *Handler
		validateFunc   func(t *testing.T, router *gin.Engine)
		name           string
		expectedRoutes []string
	}{
		{
			name: "routes registered correctly",
			setupHandler: func() *Handler {
				mockService := &mockBankCardService{}
				return NewHandler(mockService)
			},
			expectedRoutes: []string{
				"GET /bankcards",
				"GET /bankcards/:id",
				"POST /bankcards",
				"PUT /bankcards/:id",
			},
			validateFunc: func(t *testing.T, router *gin.Engine) {
				t.Helper()
				routes := router.Routes()
				assert.Len(t, routes, 4)

				// Check that all routes are registered
				methodPaths := make(map[string]string)
				for _, route := range routes {
					methodPaths[route.Method+" "+route.Path] = route.Handler
				}

				assert.Contains(t, methodPaths, "GET /bankcards")
				assert.Contains(t, methodPaths, "GET /bankcards/:id")
				assert.Contains(t, methodPaths, "POST /bankcards")
				assert.Contains(t, methodPaths, "PUT /bankcards/:id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			rootGroup := router.Group("")

			handler := tt.setupHandler()
			require.NotNil(t, handler)

			// Execute
			RegisterRoutes(rootGroup, handler)

			// Validate
			tt.validateFunc(t, router)
		})
	}
}

func TestRegisterRoutes_Integration(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rootGroup := router.Group("/api")

	mockService := &mockBankCardService{}
	handler := NewHandler(mockService)

	// Execute
	RegisterRoutes(rootGroup, handler)

	// Validate routes are accessible
	routes := router.Routes()
	require.Len(t, routes, 4)

	// Check specific route paths
	var listFound, pullFound, postFound, putFound bool
	for _, route := range routes {
		switch {
		case route.Method == http.MethodGet && route.Path == "/api/bankcards":
			listFound = true
		case route.Method == http.MethodGet && route.Path == "/api/bankcards/:id":
			pullFound = true
		case route.Method == http.MethodPost && route.Path == "/api/bankcards":
			postFound = true
		case route.Method == http.MethodPut && route.Path == "/api/bankcards/:id":
			putFound = true
		}
	}

	assert.True(t, listFound, "List route should be registered")
	assert.True(t, pullFound, "Pull route should be registered")
	assert.True(t, postFound, "Post route should be registered")
	assert.True(t, putFound, "Put route should be registered")
}

func TestRegisterRoutes_WithDifferentBasePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		basePath string
		expected []string
	}{
		{
			name:     "root path",
			basePath: "",
			expected: []string{"/bankcards", "/bankcards/:id"},
		},
		{
			name:     "api v1 path",
			basePath: "/api/v1",
			expected: []string{"/api/v1/bankcards", "/api/v1/bankcards/:id"},
		},
		{
			name:     "nested path",
			basePath: "/app/api",
			expected: []string{"/app/api/bankcards", "/app/api/bankcards/:id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			rootGroup := router.Group(tt.basePath)

			mockService := &mockBankCardService{}
			handler := NewHandler(mockService)

			// Execute
			RegisterRoutes(rootGroup, handler)

			// Validate
			routes := router.Routes()
			require.Len(t, routes, 4)

			actualPaths := make([]string, len(routes))
			for i, route := range routes {
				actualPaths[i] = route.Path
			}

			for _, expectedPath := range tt.expected {
				assert.Contains(t, actualPaths, expectedPath)
			}
		})
	}
}

func TestRegisterRoutes_HandlerMethods(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rootGroup := router.Group("")

	mockService := &mockBankCardService{}
	handler := NewHandler(mockService)

	// Execute
	RegisterRoutes(rootGroup, handler)

	// Validate that handler methods are properly set
	routes := router.Routes()
	require.Len(t, routes, 4)

	methodCounts := make(map[string]int)
	for _, route := range routes {
		// Verify that routes have handlers set
		assert.NotEmpty(t, route.Handler, "Route %s should have a handler", route.Path)

		// Count methods
		methodCounts[route.Method]++

		// Verify HTTP methods and paths
		switch {
		case route.Method == http.MethodGet && route.Path == "/bankcards":
			// List endpoint
		case route.Method == http.MethodGet && route.Path == "/bankcards/:id":
			// Pull endpoint
		case route.Method == http.MethodPost && route.Path == "/bankcards":
			// Create endpoint
		case route.Method == http.MethodPut && route.Path == "/bankcards/:id":
			// Update endpoint
		default:
			t.Errorf("Unexpected route: %s %s", route.Method, route.Path)
		}
	}

	// Verify method distribution
	assert.Equal(t, 2, methodCounts["GET"], "Should have 2 GET routes")
	assert.Equal(t, 1, methodCounts["POST"], "Should have 1 POST route")
	assert.Equal(t, 1, methodCounts["PUT"], "Should have 1 PUT route")
}
