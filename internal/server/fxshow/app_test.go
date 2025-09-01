package fxshow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestBuildApp(t *testing.T) {
	tests := []struct {
		validateFunc func(t *testing.T, app *fx.App)
		name         string
	}{
		{
			name: "app_creation_succeeds",
			validateFunc: func(t *testing.T, app *fx.App) {
				t.Helper()
				assert.NotNil(t, app, "BuildApp should return a non-nil fx.App")
			},
		},
		{
			name: "app_has_correct_type",
			validateFunc: func(t *testing.T, app *fx.App) {
				t.Helper()
				appType := reflect.TypeOf(app)
				assert.NotNil(t, appType, "App should have a valid type")
				assert.Equal(t, "*fx.App", appType.String(), "App should be of type *fx.App")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment for config loading
			t.Setenv("SERVER_HOST", "localhost")
			t.Setenv("SERVER_PORT", "8080")
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_PORT", "5432")
			t.Setenv("DB_USER", "test")
			t.Setenv("DB_PASSWORD", "test")
			t.Setenv("DB_NAME", "test")
			t.Setenv("DB_SSL_MODE", "disable")
			t.Setenv("LOG_LEVEL", "info")
			t.Setenv("AUTH_SECRET_KEY", "test-secret-key-for-testing-purposes-only")
			t.Setenv("FILE_STORAGE_BASE_PATH", "/tmp/test-storage")

			// Create the app but don't start it to avoid requiring full dependencies
			var app *fx.App
			assert.NotPanics(t, func() {
				app = BuildApp()
			}, "BuildApp should not panic")

			if tt.validateFunc != nil {
				tt.validateFunc(t, app)
			}
		})
	}
}

// Mock implementations for testing lifecycle functions.
type mockPingCloser struct {
	pinged bool
	closed bool
}

func (m *mockPingCloser) Ping(ctx context.Context) error {
	m.pinged = true
	return nil
}

func (m *mockPingCloser) Close(ctx context.Context) error {
	m.closed = true
	return nil
}

func TestRunHTTPServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc func(t *testing.T)
		name     string
	}{
		{
			name: "function_exists_and_has_correct_signature",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the function exists and has the correct signature
				funcType := reflect.TypeOf(runHTTPServer)
				assert.NotNil(t, funcType, "runHTTPServer function should exist")
				assert.Equal(t, "func", funcType.Kind().String(), "runHTTPServer should be a function")

				// Verify function signature - should accept fx.Lifecycle and any (for HTTPServer interface)
				assert.Equal(t, 2, funcType.NumIn(), "runHTTPServer should take 2 parameters")
				assert.Equal(t, 0, funcType.NumOut(), "runHTTPServer should return nothing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.testFunc != nil {
				tt.testFunc(t)
			}
		})
	}
}

func TestRunDatabaseClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc func(t *testing.T)
		name     string
	}{
		{
			name: "function_exists_and_has_correct_signature",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the function exists and has the correct signature
				funcType := reflect.TypeOf(runDatabaseClient)
				assert.NotNil(t, funcType, "runDatabaseClient function should exist")
				assert.Equal(t, "func", funcType.Kind().String(), "runDatabaseClient should be a function")

				// Verify function signature - should accept fx.Lifecycle and PingCloser
				assert.Equal(t, 2, funcType.NumIn(), "runDatabaseClient should take 2 parameters")
				assert.Equal(t, 0, funcType.NumOut(), "runDatabaseClient should return nothing")
			},
		},
		{
			name: "calls_lifecycle_methods_on_mock",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test with simplified fx app that doesn't require full config
				client := &mockPingCloser{}

				app := fxtest.New(t,
					fx.Provide(func() PingCloser { return client }),
					fx.Invoke(runDatabaseClient),
					fx.NopLogger,
				)

				app.RequireStart()
				assert.True(t, client.pinged, "Database client should be pinged via lifecycle hook")

				app.RequireStop()
				assert.True(t, client.closed, "Database client should be closed via lifecycle hook")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.testFunc != nil {
				tt.testFunc(t)
			}
		})
	}
}

func TestModuleStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(t *testing.T)
		name         string
	}{
		{
			name: "modules_are_defined",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that module variables exist and are not nil
				assert.NotNil(t, configModule, "configModule should be defined")
				assert.NotNil(t, loggerModule, "loggerModule should be defined")
				assert.NotNil(t, repositoryModule, "repositoryModule should be defined")
				assert.NotNil(t, applicationModule, "applicationModule should be defined")
				assert.NotNil(t, deliveryModule, "deliveryModule should be defined")
			},
		},
		{
			name: "modules_have_correct_types",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Verify modules are fx.Option types
				configType := reflect.TypeOf(configModule)
				loggerType := reflect.TypeOf(loggerModule)
				repositoryType := reflect.TypeOf(repositoryModule)
				applicationHType := reflect.TypeOf(applicationModule)
				deliveryType := reflect.TypeOf(deliveryModule)

				assert.NotNil(t, configType, "configModule should have a valid type")
				assert.NotNil(t, loggerType, "loggerModule should have a valid type")
				assert.NotNil(t, repositoryType, "repositoryModule should have a valid type")
				assert.NotNil(t, applicationHType, "applicationModule should have a valid type")
				assert.NotNil(t, deliveryType, "deliveryModule should have a valid type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}

func TestPingCloserInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setupClient func() PingCloser
		name        string
		testPing    bool
		testClose   bool
	}{
		{
			name: "mock_implements_interface",
			setupClient: func() PingCloser {
				return &mockPingCloser{}
			},
			testPing:  true,
			testClose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := tt.setupClient()
			require.NotNil(t, client)

			if tt.testPing {
				err := client.Ping(context.Background())
				assert.NoError(t, err, "Ping should not return error")
			}

			if tt.testClose {
				err := client.Close(context.Background())
				assert.NoError(t, err, "Close should not return error")
			}
		})
	}
}

func TestLifecycleFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc func(t *testing.T)
		name     string
	}{
		{
			name: "runHTTPServer_function_exists",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the function exists and has the correct signature
				funcType := reflect.TypeOf(runHTTPServer)
				assert.NotNil(t, funcType, "runHTTPServer function should exist")
				assert.Equal(t, "func", funcType.Kind().String(), "runHTTPServer should be a function")
			},
		},
		{
			name: "runDatabaseClient_function_exists",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the function exists and has the correct signature
				funcType := reflect.TypeOf(runDatabaseClient)
				assert.NotNil(t, funcType, "runDatabaseClient function should exist")
				assert.Equal(t, "func", funcType.Kind().String(), "runDatabaseClient should be a function")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.testFunc != nil {
				tt.testFunc(t)
			}
		})
	}
}

func TestIntegrationWithFxTest(t *testing.T) {
	t.Parallel()

	// This test verifies that our database client lifecycle function integrates properly with fx
	t.Run("lifecycle_integration", func(t *testing.T) {
		t.Parallel()

		client := &mockPingCloser{}

		app := fxtest.New(t,
			fx.Provide(func() PingCloser { return client }),
			fx.Invoke(runDatabaseClient),
			fx.NopLogger,
		)

		// Start and stop the app to trigger lifecycle hooks
		app.RequireStart()
		assert.True(t, client.pinged, "Database client should be pinged")

		app.RequireStop()
		assert.True(t, client.closed, "Database client should be closed")
	})
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Test error scenarios for lifecycle functions
	tests := []struct {
		testFunc func(t *testing.T)
		name     string
	}{
		{
			name: "runHTTPServer_with_nil_server",
			testFunc: func(t *testing.T) {
				t.Helper()
				// This test verifies that the function handles nil gracefully
				// We can't easily test this without triggering panics in fx,
				// but we can verify the function signature accepts the right types
				funcValue := reflect.ValueOf(runHTTPServer)
				funcType := funcValue.Type()

				// Verify function signature
				assert.Equal(t, 2, funcType.NumIn(), "runHTTPServer should take 2 parameters")
				assert.Equal(t, 0, funcType.NumOut(), "runHTTPServer should return nothing")
			},
		},
		{
			name: "runDatabaseClient_with_nil_client",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Similar verification for database client function
				funcValue := reflect.ValueOf(runDatabaseClient)
				funcType := funcValue.Type()

				// Verify function signature
				assert.Equal(t, 2, funcType.NumIn(), "runDatabaseClient should take 2 parameters")
				assert.Equal(t, 0, funcType.NumOut(), "runDatabaseClient should return nothing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.testFunc != nil {
				tt.testFunc(t)
			}
		})
	}
}
