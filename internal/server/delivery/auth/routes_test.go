package auth

import (
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
				mockService := &mockAuthService{}
				return NewHandler(mockService)
			},
			expectedRoutes: []string{
				"POST /auth/register",
				"POST /auth/login",
			},
			validateFunc: func(t *testing.T, router *gin.Engine) {
				t.Helper()
				routes := router.Routes()
				assert.Len(t, routes, 2)

				// Check that both routes are registered
				methodPaths := make(map[string]string)
				for _, route := range routes {
					methodPaths[route.Method+" "+route.Path] = route.Handler
				}

				assert.Contains(t, methodPaths, "POST /auth/register")
				assert.Contains(t, methodPaths, "POST /auth/login")
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

	mockService := &mockAuthService{}
	handler := NewHandler(mockService)

	// Execute
	RegisterRoutes(rootGroup, handler)

	// Validate routes are accessible
	routes := router.Routes()
	require.Len(t, routes, 2)

	// Check specific route paths
	var registerFound, loginFound bool
	for _, route := range routes {
		switch route.Path {
		case "/api/auth/register":
			assert.Equal(t, "POST", route.Method)
			registerFound = true
		case "/api/auth/login":
			assert.Equal(t, "POST", route.Method)
			loginFound = true
		}
	}

	assert.True(t, registerFound, "Register route should be registered")
	assert.True(t, loginFound, "Login route should be registered")
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
			expected: []string{"/auth/register", "/auth/login"},
		},
		{
			name:     "api v1 path",
			basePath: "/api/v1",
			expected: []string{"/api/v1/auth/register", "/api/v1/auth/login"},
		},
		{
			name:     "nested path",
			basePath: "/app/api",
			expected: []string{"/app/api/auth/register", "/app/api/auth/login"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			rootGroup := router.Group(tt.basePath)

			mockService := &mockAuthService{}
			handler := NewHandler(mockService)

			// Execute
			RegisterRoutes(rootGroup, handler)

			// Validate
			routes := router.Routes()
			require.Len(t, routes, 2)

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

	mockService := &mockAuthService{}
	handler := NewHandler(mockService)

	// Execute
	RegisterRoutes(rootGroup, handler)

	// Validate that handler methods are properly set
	routes := router.Routes()
	require.Len(t, routes, 2)

	for _, route := range routes {
		// Verify that routes have handlers set
		assert.NotEmpty(t, route.Handler, "Route %s should have a handler", route.Path)

		// Verify HTTP methods
		switch route.Path {
		case "/auth/register":
			assert.Equal(t, "POST", route.Method)
		case "/auth/login":
			assert.Equal(t, "POST", route.Method)
		default:
			t.Errorf("Unexpected route path: %s", route.Path)
		}
	}
}
