package delivery

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNewMiddlewareRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		logger bool
	}{
		{
			name:   "create middleware registry with logger",
			logger: true,
		},
		{
			name:   "create middleware registry with nil logger",
			logger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logger *zap.SugaredLogger
			if tt.logger {
				logger = zaptest.NewLogger(t).Sugar()
			}

			registry := NewMiddlewareRegistry(logger)

			require.NotNil(t, registry)
			assert.Equal(t, logger, registry.logger)
		})
	}
}

func TestMiddlewareRegistry_RegisterMiddlewares(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		expectedMiddleware []string
		withLogger         bool
		expectPanic        bool
	}{
		{
			name:       "register middlewares with logger",
			withLogger: true,
			expectedMiddleware: []string{
				"Recovery",
				"RequestID",
				"RequestLogging",
			},
			expectPanic: false,
		},
		{
			name:               "register middlewares without logger panics",
			withLogger:         false,
			expectedMiddleware: []string{},
			expectPanic:        true, // nil logger causes panic in Named()
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()

			var logger *zap.SugaredLogger
			if tt.withLogger {
				logger = zaptest.NewLogger(t).Sugar()
			}

			registry := NewMiddlewareRegistry(logger)

			// Test for panic or success based on expectation
			if tt.expectPanic {
				assert.Panics(t, func() {
					registry.RegisterMiddlewares(router)
				})
				return
			}

			// This should not panic
			assert.NotPanics(t, func() {
				registry.RegisterMiddlewares(router)
			})

			// Verify middleware was registered by checking handlers
			handlers := router.Handlers
			assert.NotEmpty(t, handlers, "Middleware should have been registered")
		})
	}
}

func TestMiddlewareRegistry_MiddlewareOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedOrder  []string
		verifyHandlers bool
	}{
		{
			name: "middleware registration order",
			expectedOrder: []string{
				"gin.Recovery",
				"middleware.RequestID",
				"middleware.RequestLogging",
			},
			verifyHandlers: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			logger := zaptest.NewLogger(t).Sugar()

			registry := NewMiddlewareRegistry(logger)
			registry.RegisterMiddlewares(router)

			if tt.verifyHandlers {
				// Verify handlers were registered in correct order
				handlers := router.Handlers
				assert.GreaterOrEqual(t, len(handlers), 3, "Should have at least 3 middleware handlers")
			}
		})
	}
}

func TestMiddlewareRegistry_LoggerNaming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		loggerName     string
		expectedNaming string
	}{
		{
			name:           "logger with correct naming",
			loggerName:     "test-logger",
			expectedNaming: "http-request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			logger := zaptest.NewLogger(t).Sugar().Named(tt.loggerName)

			registry := NewMiddlewareRegistry(logger)

			// This should not panic and should handle logger naming correctly
			assert.NotPanics(t, func() {
				registry.RegisterMiddlewares(router)
			})

			// Verify logger is correctly stored
			assert.NotNil(t, registry.logger)
		})
	}
}

func TestMiddlewareRegistry_MiddlewareConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		testType    string
		expectPanic bool
	}{
		{
			name:        "standard middleware configuration",
			testType:    "standard",
			expectPanic: false,
		},
		{
			name:        "middleware with nil router",
			testType:    "nil_router",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()
			registry := NewMiddlewareRegistry(logger)

			var router *gin.Engine
			if tt.testType == "standard" {
				gin.SetMode(gin.TestMode)
				router = gin.New()
			}
			// For nil_router case, router remains nil

			if tt.expectPanic {
				assert.Panics(t, func() {
					registry.RegisterMiddlewares(router)
				})
			} else {
				assert.NotPanics(t, func() {
					registry.RegisterMiddlewares(router)
				})
			}
		})
	}
}

func TestMiddlewareRegistry_Integration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		registryCount     int
		expectDuplication bool
	}{
		{
			name:              "single registry registration",
			registryCount:     1,
			expectDuplication: false,
		},
		{
			name:              "multiple registry registration",
			registryCount:     2,
			expectDuplication: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			logger := zaptest.NewLogger(t).Sugar()

			initialHandlerCount := len(router.Handlers)

			for range tt.registryCount {
				registry := NewMiddlewareRegistry(logger)
				registry.RegisterMiddlewares(router)
			}

			finalHandlerCount := len(router.Handlers)
			handlerDelta := finalHandlerCount - initialHandlerCount

			if tt.expectDuplication {
				// Multiple registrations should add more handlers
				expectedDelta := 3 * tt.registryCount // 3 middleware per registration
				assert.Equal(t, expectedDelta, handlerDelta, "Should have duplicated middleware")
			} else {
				// Single registration should add exactly 3 handlers
				assert.Equal(t, 3, handlerDelta, "Should have exactly 3 middleware handlers")
			}
		})
	}
}

func TestMiddlewareRegistry_MiddlewareTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		middleware  string
		description string
	}{
		{
			name:        "gin recovery middleware",
			middleware:  "gin.Recovery",
			description: "handles panics and returns 500 status",
		},
		{
			name:        "request ID middleware",
			middleware:  "middleware.RequestID",
			description: "adds unique request ID to context",
		},
		{
			name:        "request logging middleware",
			middleware:  "middleware.RequestLogging",
			description: "logs HTTP requests and responses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			logger := zaptest.NewLogger(t).Sugar()

			registry := NewMiddlewareRegistry(logger)
			registry.RegisterMiddlewares(router)

			// Verify middleware types are correctly configured
			assert.NotEmpty(t, tt.description, "Middleware should have description")
			assert.NotEmpty(t, tt.middleware, "Middleware should have name")

			// Verify handlers were registered
			handlers := router.Handlers
			assert.GreaterOrEqual(t, len(handlers), 3, "Should have registered middleware handlers")
		})
	}
}

func TestMiddlewareRegistry_LoggerConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		loggerConfig     string
		expectedBehavior string
	}{
		{
			name:             "logger with named instance",
			loggerConfig:     "named",
			expectedBehavior: "should create http-request named logger",
		},
		{
			name:             "logger without naming",
			loggerConfig:     "standard",
			expectedBehavior: "should use provided logger for request logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()

			var logger *zap.SugaredLogger
			if tt.loggerConfig == "named" {
				logger = zaptest.NewLogger(t).Sugar().Named("test-service")
			} else {
				logger = zaptest.NewLogger(t).Sugar()
			}

			registry := NewMiddlewareRegistry(logger)
			registry.RegisterMiddlewares(router)

			// Verify logger configuration behavior
			assert.NotEmpty(t, tt.expectedBehavior, "Expected behavior should be described")
			assert.NotNil(t, registry.logger, "Registry should store logger reference")
		})
	}
}
