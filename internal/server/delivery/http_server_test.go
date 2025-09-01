package delivery

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// Mock implementations for testing.
type mockRouteConfigurator struct {
	registerRoutesFunc func(router *gin.Engine)
}

func (m *mockRouteConfigurator) RegisterRoutes(router *gin.Engine) {
	if m.registerRoutesFunc != nil {
		m.registerRoutesFunc(router)
	}
}

type mockMiddlewareConfigurator struct {
	registerMiddlewaresFunc func(router *gin.Engine)
}

func (m *mockMiddlewareConfigurator) RegisterMiddlewares(router *gin.Engine) {
	if m.registerMiddlewaresFunc != nil {
		m.registerMiddlewaresFunc(router)
	}
}

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		addr         string
		certFile     string
		keyFile      string
		startTimeout time.Duration
		stopTimeout  time.Duration
		tlsEnabled   bool
	}{
		{
			name:         "HTTP server configuration",
			addr:         ":8080",
			startTimeout: 5 * time.Second,
			stopTimeout:  10 * time.Second,
			tlsEnabled:   false,
			certFile:     "",
			keyFile:      "",
		},
		{
			name:         "HTTPS server configuration",
			addr:         ":8443",
			startTimeout: 10 * time.Second,
			stopTimeout:  15 * time.Second,
			tlsEnabled:   true,
			certFile:     "/path/to/cert.pem",
			keyFile:      "/path/to/key.pem",
		},
		{
			name:         "custom address and timeouts",
			addr:         "localhost:3000",
			startTimeout: 30 * time.Second,
			stopTimeout:  60 * time.Second,
			tlsEnabled:   false,
			certFile:     "",
			keyFile:      "",
		},
		{
			name:         "zero timeouts",
			addr:         ":0",
			startTimeout: 0,
			stopTimeout:  0,
			tlsEnabled:   false,
			certFile:     "",
			keyFile:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()

			rc := &mockRouteConfigurator{
				registerRoutesFunc: func(router *gin.Engine) {
					router.GET("/test", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"status": "ok"})
					})
				},
			}

			mc := &mockMiddlewareConfigurator{
				registerMiddlewaresFunc: func(router *gin.Engine) {
					router.Use(gin.Recovery())
				},
			}

			server := NewHTTPServer(
				logger,
				rc,
				mc,
				tt.addr,
				tt.startTimeout,
				tt.stopTimeout,
				tt.tlsEnabled,
				tt.certFile,
				tt.keyFile,
			)

			require.NotNil(t, server)
			assert.NotNil(t, server.l)
			assert.NotNil(t, server.server)
			assert.Equal(t, tt.addr, server.server.Addr)
			assert.Equal(t, tt.startTimeout, server.startTimeout)
			assert.Equal(t, tt.stopTimeout, server.stopTimeout)
			assert.Equal(t, tt.tlsEnabled, server.tlsEnabled)
			assert.Equal(t, tt.certFile, server.certFile)
			assert.Equal(t, tt.keyFile, server.keyFile)
			assert.NotNil(t, server.server.Handler)
		})
	}
}

func TestHTTPServer_Start(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		addr         string
		errContains  string
		startTimeout time.Duration
		tlsEnabled   bool
		expectErr    bool
	}{
		{
			name:         "start HTTP server with invalid address",
			addr:         "invalid-address",
			startTimeout: 100 * time.Millisecond,
			tlsEnabled:   false,
			expectErr:    true,
			errContains:  "HTTP server start failed",
		},
		{
			name:         "start HTTPS server without certificates",
			addr:         ":0",
			startTimeout: 100 * time.Millisecond,
			tlsEnabled:   true,
			expectErr:    true,
			errContains:  "HTTP server start failed",
		},
		{
			name:         "start with very short timeout",
			addr:         ":0",
			startTimeout: 1 * time.Nanosecond,
			tlsEnabled:   false,
			expectErr:    false, // Timeout doesn't necessarily cause error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()

			rc := &mockRouteConfigurator{}
			mc := &mockMiddlewareConfigurator{}

			server := NewHTTPServer(
				logger,
				rc,
				mc,
				tt.addr,
				tt.startTimeout,
				10*time.Second, // stopTimeout
				tt.tlsEnabled,
				"nonexistent-cert.pem",
				"nonexistent-key.pem",
			)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := server.Start(ctx)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else if err == nil {
				// For successful starts, we need to stop the server
				stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer stopCancel()
				_ = server.Stop(stopCtx)
			}
		})
	}
}

func TestHTTPServer_Stop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		stopTimeout time.Duration
		expectErr   bool
	}{
		{
			name:        "stop with normal timeout",
			stopTimeout: 5 * time.Second,
			expectErr:   false,
		},
		{
			name:        "stop with zero timeout",
			stopTimeout: 0,
			expectErr:   false,
		},
		{
			name:        "stop with very short timeout",
			stopTimeout: 1 * time.Nanosecond,
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()
			rc := &mockRouteConfigurator{}
			mc := &mockMiddlewareConfigurator{}

			server := NewHTTPServer(
				logger,
				rc,
				mc,
				":0",
				100*time.Millisecond,
				tt.stopTimeout,
				false,
				"",
				"",
			)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := server.Stop(ctx)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				// Stop should not error even if server wasn't started
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPServer_GetProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedResult string
		tlsEnabled     bool
	}{
		{
			name:           "HTTP protocol",
			tlsEnabled:     false,
			expectedResult: "HTTP",
		},
		{
			name:           "HTTPS protocol",
			tlsEnabled:     true,
			expectedResult: "HTTPS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()
			rc := &mockRouteConfigurator{}
			mc := &mockMiddlewareConfigurator{}

			server := NewHTTPServer(
				logger,
				rc,
				mc,
				":8080",
				5*time.Second,
				10*time.Second,
				tt.tlsEnabled,
				"cert.pem",
				"key.pem",
			)

			protocol := server.getProtocol()
			assert.Equal(t, tt.expectedResult, protocol)
		})
	}
}

func TestHTTPServer_StartCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		errContains  string
		startTimeout time.Duration
		sendError    bool
		cancelCtx    bool
		expectErr    bool
	}{
		{
			name:         "successful start check",
			startTimeout: 100 * time.Millisecond,
			sendError:    false,
			cancelCtx:    false,
			expectErr:    false,
		},
		{
			name:         "start check with error",
			startTimeout: 100 * time.Millisecond,
			sendError:    true,
			cancelCtx:    false,
			expectErr:    true,
			errContains:  "server startup error",
		},
		{
			name:         "start check with context cancellation",
			startTimeout: 1 * time.Second,
			sendError:    false,
			cancelCtx:    true,
			expectErr:    true,
			errContains:  "start cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()
			rc := &mockRouteConfigurator{}
			mc := &mockMiddlewareConfigurator{}

			server := NewHTTPServer(
				logger,
				rc,
				mc,
				":0",
				tt.startTimeout,
				5*time.Second,
				false,
				"",
				"",
			)

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			if tt.cancelCtx {
				// Cancel context immediately for testing cancellation
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			errChan := make(chan error, 1)
			if tt.sendError {
				go func() {
					time.Sleep(10 * time.Millisecond)
					errChan <- assert.AnError
				}()
			}

			err := server.startCheck(ctx, errChan)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPServer_Listen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		testType   string
		tlsEnabled bool
	}{
		{
			name:       "HTTP listen",
			tlsEnabled: false,
			testType:   "method signature",
		},
		{
			name:       "HTTPS listen",
			tlsEnabled: true,
			testType:   "method signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t).Sugar()
			rc := &mockRouteConfigurator{}
			mc := &mockMiddlewareConfigurator{}

			server := NewHTTPServer(
				logger,
				rc,
				mc,
				":0",
				100*time.Millisecond,
				5*time.Second,
				tt.tlsEnabled,
				"cert.pem",
				"key.pem",
			)

			// Test method signature exists
			assert.NotNil(t, server.listen)
			assert.NotNil(t, server.listenHTTP)
			assert.NotNil(t, server.listenHTTPS)
		})
	}
}

func TestRouteConfigurator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedRoutes []string
		routeCount     int
	}{
		{
			name:       "single route registration",
			routeCount: 1,
			expectedRoutes: []string{
				"/test",
			},
		},
		{
			name:       "multiple route registration",
			routeCount: 3,
			expectedRoutes: []string{
				"/users",
				"/posts",
				"/comments",
			},
		},
		{
			name:           "no routes registration",
			routeCount:     0,
			expectedRoutes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registeredRoutes := make([]string, 0)

			rc := &mockRouteConfigurator{
				registerRoutesFunc: func(router *gin.Engine) {
					for _, route := range tt.expectedRoutes {
						registeredRoutes = append(registeredRoutes, route)
						router.GET(route, func(c *gin.Context) {
							c.JSON(http.StatusOK, gin.H{"status": "ok"})
						})
					}
				},
			}

			router := gin.New()
			rc.RegisterRoutes(router)

			assert.Len(t, registeredRoutes, tt.routeCount)
			for i, expectedRoute := range tt.expectedRoutes {
				if i < len(registeredRoutes) {
					assert.Equal(t, expectedRoute, registeredRoutes[i])
				}
			}
		})
	}
}

func TestMiddlewareConfigurator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		expectedMiddleware []string
		middlewareCount    int
	}{
		{
			name:            "single middleware registration",
			middlewareCount: 1,
			expectedMiddleware: []string{
				"recovery",
			},
		},
		{
			name:            "multiple middleware registration",
			middlewareCount: 3,
			expectedMiddleware: []string{
				"logger",
				"recovery",
				"cors",
			},
		},
		{
			name:               "no middleware registration",
			middlewareCount:    0,
			expectedMiddleware: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registeredMiddleware := make([]string, 0)

			mc := &mockMiddlewareConfigurator{
				registerMiddlewaresFunc: func(router *gin.Engine) {
					for _, middleware := range tt.expectedMiddleware {
						registeredMiddleware = append(registeredMiddleware, middleware)
						// Mock middleware registration
						router.Use(func(c *gin.Context) {
							c.Next()
						})
					}
				},
			}

			router := gin.New()
			mc.RegisterMiddlewares(router)

			assert.Len(t, registeredMiddleware, tt.middlewareCount)
			for i, expectedMw := range tt.expectedMiddleware {
				if i < len(registeredMiddleware) {
					assert.Equal(t, expectedMw, registeredMiddleware[i])
				}
			}
		})
	}
}

func TestGinMode(t *testing.T) {
	tests := []struct {
		name         string
		expectedMode string
	}{
		{
			name:         "gin mode is set",
			expectedMode: gin.Mode(), // Get whatever mode is currently set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that gin mode can be retrieved
			mode := gin.Mode()
			assert.NotEmpty(t, mode, "Gin mode should be set")

			// Verify it's one of the valid modes
			validModes := []string{gin.DebugMode, gin.ReleaseMode, gin.TestMode}
			assert.Contains(t, validModes, mode, "Should be a valid gin mode")
		})
	}
}
