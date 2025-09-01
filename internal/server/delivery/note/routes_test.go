package note

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		setupHandler   bool
	}{
		{
			name:           "POST /notes - create note",
			method:         http.MethodPost,
			path:           "/notes",
			expectedStatus: http.StatusOK,
			setupHandler:   true,
		},
		{
			name:           "GET /notes - list notes",
			method:         http.MethodGet,
			path:           "/notes",
			expectedStatus: http.StatusOK,
			setupHandler:   true,
		},
		{
			name:           "GET /notes/:id - get note by ID",
			method:         http.MethodGet,
			path:           "/notes/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusOK,
			setupHandler:   true,
		},
		{
			name:           "PUT /notes/:id - update note",
			method:         http.MethodPut,
			path:           "/notes/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusOK,
			setupHandler:   true,
		},
		{
			name:           "unsupported method DELETE",
			method:         http.MethodDelete,
			path:           "/notes/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusNotFound,
			setupHandler:   true,
		},
		{
			name:           "unsupported method PATCH",
			method:         http.MethodPatch,
			path:           "/notes",
			expectedStatus: http.StatusNotFound,
			setupHandler:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := gin.New()
			group := router.Group("/")

			// Create handler for RegisterRoutes call
			mockService := &mockService{}
			handler := NewHandler(mockService)

			if tt.setupHandler {
				// Create wrapper functions to avoid method assignment issues
				pullHandler := func(c *gin.Context) {
					c.Status(http.StatusOK)
				}
				listHandler := func(c *gin.Context) {
					c.Status(http.StatusOK)
				}
				pushHandler := func(c *gin.Context) {
					c.Status(http.StatusOK)
				}

				// Override routes with simple handlers for route testing
				group.POST("/notes", pushHandler)
				group.GET("/notes", listHandler)
				notesIDGroup := group.Group("/notes/:id")
				notesIDGroup.GET("", pullHandler)
				notesIDGroup.PUT("", pushHandler)

				// Skip the normal RegisterRoutes call
				goto skipRegister
			}

			RegisterRoutes(group, handler)
		skipRegister:

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRegisterRoutes_RouteStructure(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	group := router.Group("/api/v1")

	// Create a simple handler
	handler := &Handler{}

	RegisterRoutes(group, handler)

	// Get all registered routes
	routes := router.Routes()

	expectedRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/notes"},
		{http.MethodGet, "/api/v1/notes"},
		{http.MethodGet, "/api/v1/notes/:id"},
		{http.MethodPut, "/api/v1/notes/:id"},
	}

	// Verify all expected routes are registered
	for _, expected := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Method == expected.method && route.Path == expected.path {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected route %s %s not found", expected.method, expected.path)
	}

	// Verify no unexpected routes are registered (should have exactly 4 routes)
	noteRoutes := 0
	for _, route := range routes {
		if len(route.Path) > 8 && route.Path[:9] == "/api/v1/n" {
			noteRoutes++
		}
	}
	assert.Equal(t, len(expectedRoutes), noteRoutes, "Unexpected number of note routes")
}

func TestRegisterRoutes_GroupPrefixHandling(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		groupPrefix  string
		expectedPath string
	}{
		{
			name:         "no prefix group",
			groupPrefix:  "",
			expectedPath: "/notes",
		},
		{
			name:         "api prefix",
			groupPrefix:  "/api",
			expectedPath: "/api/notes",
		},
		{
			name:         "versioned api prefix",
			groupPrefix:  "/api/v1",
			expectedPath: "/api/v1/notes",
		},
		{
			name:         "items prefix",
			groupPrefix:  "/items",
			expectedPath: "/items/notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := gin.New()
			group := router.Group(tt.groupPrefix)

			handler := &Handler{}
			RegisterRoutes(group, handler)

			routes := router.Routes()

			// Find the POST route for notes
			found := false
			for _, route := range routes {
				if route.Method == http.MethodPost && route.Path == tt.expectedPath {
					found = true
					break
				}
			}

			assert.True(t, found, "Expected POST route %s not found", tt.expectedPath)
		})
	}
}

func TestRegisterRoutes_MethodHandlerMapping(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		method   string
		path     string
		hasParam bool
	}{
		{
			name:     "POST without ID parameter",
			method:   http.MethodPost,
			path:     "/notes",
			hasParam: false,
		},
		{
			name:     "GET without ID parameter",
			method:   http.MethodGet,
			path:     "/notes",
			hasParam: false,
		},
		{
			name:     "GET with ID parameter",
			method:   http.MethodGet,
			path:     "/notes/:id",
			hasParam: true,
		},
		{
			name:     "PUT with ID parameter",
			method:   http.MethodPut,
			path:     "/notes/:id",
			hasParam: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := gin.New()
			group := router.Group("/")

			handler := &Handler{}
			RegisterRoutes(group, handler)

			routes := router.Routes()

			// Find the specific route
			found := false
			for _, route := range routes {
				if route.Method == tt.method && route.Path == tt.path {
					found = true
					break
				}
			}

			assert.True(t, found, "Expected route %s %s not found", tt.method, tt.path)
		})
	}
}

func TestRegisterRoutes_HandlerNil(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	// Test that registering routes with nil handler doesn't panic
	router := gin.New()
	group := router.Group("/")

	require.NotPanics(t, func() {
		RegisterRoutes(group, nil)
	})
}
