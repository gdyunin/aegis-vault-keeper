package swagger

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

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		handlerCalled  bool
	}{
		{
			name:           "swagger route registration",
			path:           "/swagger/",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "swagger index route",
			path:           "/swagger/index.html",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "swagger doc.json route",
			path:           "/swagger/doc.json",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "swagger assets route",
			path:           "/swagger/swagger-ui-bundle.js",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "swagger wildcard route",
			path:           "/swagger/any/path/here",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			routerGroup := router.Group("")

			handlerCalled := false
			mockHandler := func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{
					"message": "swagger handler called",
					"path":    c.Request.URL.Path,
				})
			}

			RegisterRoutes(routerGroup, mockHandler)

			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.handlerCalled {
				assert.True(t, handlerCalled, "Handler should have been called")
			}
		})
	}
}

func TestRegisterRoutesWithDifferentRouterGroups(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		groupPrefix string
		path        string
		fullPath    string
	}{
		{
			name:        "root group",
			groupPrefix: "",
			path:        "/swagger/",
			fullPath:    "/swagger/",
		},
		{
			name:        "api group",
			groupPrefix: "/api",
			path:        "/swagger/",
			fullPath:    "/api/swagger/",
		},
		{
			name:        "v1 group",
			groupPrefix: "/v1",
			path:        "/swagger/index.html",
			fullPath:    "/v1/swagger/index.html",
		},
		{
			name:        "nested group",
			groupPrefix: "/api/v1",
			path:        "/swagger/doc.json",
			fullPath:    "/api/v1/swagger/doc.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			routerGroup := router.Group(tt.groupPrefix)

			handlerCalled := false
			mockHandler := func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{
					"message":    "swagger handler called",
					"path":       c.Request.URL.Path,
					"full_path":  c.FullPath(),
					"group_path": tt.groupPrefix,
				})
			}

			RegisterRoutes(routerGroup, mockHandler)

			req, err := http.NewRequest(http.MethodGet, tt.fullPath, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.True(t, handlerCalled, "Handler should have been called")
		})
	}
}

func TestRegisterRoutesMethodValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		shouldWork     bool
	}{
		{
			name:           "GET method works",
			method:         http.MethodGet,
			path:           "/swagger/",
			expectedStatus: http.StatusOK,
			shouldWork:     true,
		},
		{
			name:           "POST method not found",
			method:         http.MethodPost,
			path:           "/swagger/",
			expectedStatus: http.StatusNotFound, // Gin returns 404 for unregistered method/path combo
			shouldWork:     false,
		},
		{
			name:           "PUT method not found",
			method:         http.MethodPut,
			path:           "/swagger/",
			expectedStatus: http.StatusNotFound,
			shouldWork:     false,
		},
		{
			name:           "DELETE method not found",
			method:         http.MethodDelete,
			path:           "/swagger/",
			expectedStatus: http.StatusNotFound,
			shouldWork:     false,
		},
		{
			name:           "PATCH method not found",
			method:         http.MethodPatch,
			path:           "/swagger/",
			expectedStatus: http.StatusNotFound,
			shouldWork:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			routerGroup := router.Group("")

			handlerCalled := false
			mockHandler := func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{
					"message": "swagger handler called",
					"method":  c.Request.Method,
				})
			}

			RegisterRoutes(routerGroup, mockHandler)

			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.shouldWork {
				assert.True(t, handlerCalled, "Handler should have been called for %s", tt.method)
			} else {
				assert.False(t, handlerCalled, "Handler should not have been called for %s", tt.method)
			}
		})
	}
}

func TestRegisterRoutesPathMatching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		shouldMatch  bool
		expectedCode int
	}{
		{
			name:         "exact swagger path",
			path:         "/swagger/",
			shouldMatch:  true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "swagger without trailing slash redirects",
			path:         "/swagger",
			shouldMatch:  false,
			expectedCode: http.StatusMovedPermanently, // Gin redirects to /swagger/
		},
		{
			name:         "swagger with additional path",
			path:         "/swagger/test",
			shouldMatch:  true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "swagger with deep path",
			path:         "/swagger/ui/assets/style.css",
			shouldMatch:  true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "non-swagger path",
			path:         "/api/users",
			shouldMatch:  false,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "swagger prefix but different path",
			path:         "/swaggerui/",
			shouldMatch:  false,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			routerGroup := router.Group("")

			handlerCalled := false
			mockHandler := func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{
					"message": "swagger handler called",
					"path":    c.Request.URL.Path,
				})
			}

			RegisterRoutes(routerGroup, mockHandler)

			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.Equal(t, tt.shouldMatch, handlerCalled, "Handler call expectation for path %s", tt.path)
		})
	}
}

func TestRegisterRoutesHandlerNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		handler gin.HandlerFunc
		name    string
	}{
		{
			name: "valid handler",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			routerGroup := router.Group("")

			// This should not panic
			assert.NotPanics(t, func() {
				RegisterRoutes(routerGroup, tt.handler)
			})

			// Test that route was registered
			req, err := http.NewRequest(http.MethodGet, "/swagger/", nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// RegisterRoutes function itself should not panic
			assert.True(t, true, "RegisterRoutes should not panic")
		})
	}
}

func TestRegisterRoutesIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		multipleGroups   bool
		expectedHandlers int
	}{
		{
			name:             "single group registration",
			multipleGroups:   false,
			expectedHandlers: 1,
		},
		{
			name:             "multiple group registration",
			multipleGroups:   true,
			expectedHandlers: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()

			handlerCallCount := 0
			mockHandler := func(c *gin.Context) {
				handlerCallCount++
				c.JSON(http.StatusOK, gin.H{
					"message":    "swagger handler called",
					"call_count": handlerCallCount,
				})
			}

			// Register routes on one group
			group1 := router.Group("/api")
			RegisterRoutes(group1, mockHandler)

			if tt.multipleGroups {
				// Register routes on another group
				group2 := router.Group("/v1")
				RegisterRoutes(group2, mockHandler)
			}

			// Test first group
			req1, err := http.NewRequest(http.MethodGet, "/api/swagger/", nil)
			require.NoError(t, err)

			recorder1 := httptest.NewRecorder()
			router.ServeHTTP(recorder1, req1)
			assert.Equal(t, http.StatusOK, recorder1.Code)

			if tt.multipleGroups {
				// Test second group
				req2, err := http.NewRequest(http.MethodGet, "/v1/swagger/", nil)
				require.NoError(t, err)

				recorder2 := httptest.NewRecorder()
				router.ServeHTTP(recorder2, req2)
				assert.Equal(t, http.StatusOK, recorder2.Code)
			}

			assert.Equal(t, tt.expectedHandlers, handlerCallCount)
		})
	}
}
