package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	// Set gin to test mode to reduce noise
	gin.SetMode(gin.TestMode)

	t.Run("adds_request_id_header", func(t *testing.T) {
		t.Parallel()

		// Create a test handler that checks if the header was set
		var receivedRequestID string
		testHandler := func(c *gin.Context) {
			receivedRequestID = c.Request.Header.Get(consts.HeaderXRequestID)
			c.Status(http.StatusOK)
		}

		// Setup router with RequestID middleware
		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", testHandler)

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify request ID was set
		assert.NotEmpty(t, receivedRequestID)

		// Verify it's a valid UUID
		parsedUUID, err := uuid.Parse(receivedRequestID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, parsedUUID)
	})

	t.Run("generates_different_request_ids", func(t *testing.T) {
		t.Parallel()

		var requestIDs []string
		testHandler := func(c *gin.Context) {
			requestID := c.Request.Header.Get(consts.HeaderXRequestID)
			requestIDs = append(requestIDs, requestID)
			c.Status(http.StatusOK)
		}

		// Setup router
		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", testHandler)

		// Make multiple requests
		numRequests := 5
		for range numRequests {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify we got the expected number of request IDs
		assert.Len(t, requestIDs, numRequests)

		// Verify all request IDs are different
		uniqueIDs := make(map[string]bool)
		for _, id := range requestIDs {
			assert.NotEmpty(t, id)

			// Check if this ID was already seen
			assert.False(t, uniqueIDs[id], "Request ID %s was generated twice", id)
			uniqueIDs[id] = true

			// Verify it's a valid UUID
			_, err := uuid.Parse(id)
			assert.NoError(t, err, "Request ID %s is not a valid UUID", id)
		}
	})

	t.Run("calls_next_handler", func(t *testing.T) {
		t.Parallel()

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		}

		// Setup router
		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", testHandler)

		// Create and execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Verify the handler was called
		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("works_with_existing_header", func(t *testing.T) {
		t.Parallel()

		var finalRequestID string
		testHandler := func(c *gin.Context) {
			finalRequestID = c.Request.Header.Get(consts.HeaderXRequestID)
			c.Status(http.StatusOK)
		}

		// Setup router
		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", testHandler)

		// Create request with existing X-Request-Id header
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		existingRequestID := "existing-request-id"
		req.Header.Set(consts.HeaderXRequestID, existingRequestID)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify the middleware overwrote the existing header with a new UUID
		assert.NotEqual(t, existingRequestID, finalRequestID)
		assert.NotEmpty(t, finalRequestID)

		// Verify the new value is a valid UUID
		_, err := uuid.Parse(finalRequestID)
		require.NoError(t, err)
	})

	t.Run("middleware_chain", func(t *testing.T) {
		t.Parallel()

		var requestIDFromMiddleware1 string
		var requestIDFromMiddleware2 string
		var requestIDFromHandler string

		middleware1 := func(c *gin.Context) {
			requestIDFromMiddleware1 = c.Request.Header.Get(consts.HeaderXRequestID)
			c.Next()
		}

		middleware2 := func(c *gin.Context) {
			requestIDFromMiddleware2 = c.Request.Header.Get(consts.HeaderXRequestID)
			c.Next()
		}

		testHandler := func(c *gin.Context) {
			requestIDFromHandler = c.Request.Header.Get(consts.HeaderXRequestID)
			c.Status(http.StatusOK)
		}

		// Setup router with middleware chain
		router := gin.New()
		router.Use(RequestID())
		router.Use(middleware1)
		router.Use(middleware2)
		router.GET("/test", testHandler)

		// Create and execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Verify all middlewares and handler received the same request ID
		assert.NotEmpty(t, requestIDFromMiddleware1)
		assert.Equal(t, requestIDFromMiddleware1, requestIDFromMiddleware2)
		assert.Equal(t, requestIDFromMiddleware1, requestIDFromHandler)

		// Verify it's a valid UUID
		_, err := uuid.Parse(requestIDFromMiddleware1)
		require.NoError(t, err)
	})

	t.Run("different_http_methods", func(t *testing.T) {
		t.Parallel()

		var requestIDs []string
		testHandler := func(c *gin.Context) {
			requestID := c.Request.Header.Get(consts.HeaderXRequestID)
			requestIDs = append(requestIDs, requestID)
			c.Status(http.StatusOK)
		}

		// Setup router
		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", testHandler)
		router.POST("/test", testHandler)
		router.PUT("/test", testHandler)
		router.DELETE("/test", testHandler)

		// Test different HTTP methods
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify we got request IDs for all methods
		assert.Len(t, requestIDs, len(methods))

		// Verify all are unique and valid UUIDs
		uniqueIDs := make(map[string]bool)
		for i, id := range requestIDs {
			assert.NotEmpty(t, id, "Request ID for method %s is empty", methods[i])
			assert.False(t, uniqueIDs[id], "Request ID %s was generated twice", id)
			uniqueIDs[id] = true

			_, err := uuid.Parse(id)
			assert.NoError(t, err, "Request ID %s is not a valid UUID", id)
		}
	})
}

// Benchmark the RequestID middleware performance.
func BenchmarkRequestID(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for range b.N {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}
}
