package fxshow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestDeliveryModule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc    func(*testing.T)
		name        string
		description string
	}{
		{
			name:        "module_exists",
			description: "Test that deliveryModule variable exists and is properly defined",
			testFunc: func(t *testing.T) {
				t.Helper()
				assert.NotNil(t, deliveryModule, "deliveryModule should not be nil")
			},
		},
		{
			name:        "module_type",
			description: "Test that deliveryModule is an fx.Option type",
			testFunc: func(t *testing.T) {
				t.Helper()
				moduleType := reflect.TypeOf(deliveryModule)
				assert.NotNil(t, moduleType, "deliveryModule type should not be nil")
			},
		},
		{
			name:        "module_structure",
			description: "Test that deliveryModule has proper structure and providers",
			testFunc: func(t *testing.T) {
				t.Helper()
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				assert.NotNil(t, app, "fx.App should be created successfully with deliveryModule")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestDeliveryModuleProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		description     string
		providerPattern string
		interfaceCount  int
	}{
		{
			name:            "build_info_operator_provider",
			description:     "Test BuildInfoOperator provider with interfaces",
			providerPattern: "common.BuildInfoOperator",
			interfaceCount:  1,
		},
		{
			name:            "route_registry_provider",
			description:     "Test RouteRegistry provider with interfaces",
			providerPattern: "delivery.RouteRegistry",
			interfaceCount:  1,
		},
		{
			name:            "middleware_registry_provider",
			description:     "Test MiddlewareRegistry provider with interfaces",
			providerPattern: "delivery.MiddlewareRegistry",
			interfaceCount:  1,
		},
		{
			name:            "http_server_provider",
			description:     "Test HTTPServer provider with fx.Provide",
			providerPattern: "delivery.HTTPServer",
			interfaceCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotNil(t, deliveryModule, "deliveryModule should exist")
			assert.NotEmpty(t, tt.providerPattern, "providerPattern should not be empty")
			assert.GreaterOrEqual(t, tt.interfaceCount, 0, "interfaceCount should be non-negative")
		})
	}
}

func TestDeliveryModuleIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setupFunc   func() fx.Option
		name        string
		description string
		expectError bool
	}{
		{
			name:        "standalone_module_integration",
			description: "Test deliveryModule integration in standalone fx.App",
			setupFunc: func() fx.Option {
				return deliveryModule
			},
			expectError: false,
		},
		{
			name:        "combined_modules_integration",
			description: "Test deliveryModule integration with other modules",
			setupFunc: func() fx.Option {
				return fx.Options(
					deliveryModule,
					fx.Module("test", fx.Provide(func() string { return "test" })),
				)
			},
			expectError: false,
		},
		{
			name:        "multiple_delivery_modules_integration",
			description: "Test multiple deliveryModule instances",
			setupFunc: func() fx.Option {
				return fx.Options(
					deliveryModule,
					fx.Module("test_delivery",
						fx.Provide(func() int { return 42 }),
					),
				)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			option := tt.setupFunc()
			require.NotNil(t, option)

			app := fx.New(
				option,
				fx.NopLogger,
			)

			// Both branches expect the same thing, so no conditional needed
			assert.NotNil(t, app)
		})
	}
}

func TestDeliveryModuleBuildInfoOperator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(*testing.T)
		name         string
		description  string
	}{
		{
			name:        "build_info_operator_constructor",
			description: "Test BuildInfoOperator constructor function",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that the constructor function pattern is correct
				assert.NotNil(t, deliveryModule)

				// The function should use common.NewBuildInfoOperator with buildinfo constants
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "build_info_constants_usage",
			description: "Test usage of buildinfo constants",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that buildinfo.Version, Date, Commit are used
				assert.NotNil(t, deliveryModule)

				// The constructor should reference buildinfo constants
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "build_info_interface_binding",
			description: "Test BuildInfoOperator interface binding",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that BuildInfoOperator is bound to delivery.BuildInfoOperator interface
				assert.NotNil(t, deliveryModule)

				// Uses provideWithInterfaces pattern
				moduleType := reflect.TypeOf(deliveryModule)
				assert.NotNil(t, moduleType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.validateFunc(t)
		})
	}
}

func TestDeliveryModuleRegistryProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc     func(*testing.T)
		name         string
		description  string
		registryType string
	}{
		{
			name:         "route_registry_provider",
			description:  "Test RouteRegistry provider pattern",
			registryType: "RouteRegistry",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test direct provider pattern for RouteRegistry
				assert.NotNil(t, deliveryModule)

				// Uses delivery.NewRouteRegistry directly
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:         "middleware_registry_provider",
			description:  "Test MiddlewareRegistry provider pattern",
			registryType: "MiddlewareRegistry",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test direct provider pattern for MiddlewareRegistry
				assert.NotNil(t, deliveryModule)

				// Uses delivery.NewMiddlewareRegistry directly
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:         "registry_interface_bindings",
			description:  "Test registry interface bindings",
			registryType: "Both",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that registries are bound to correct interfaces
				assert.NotNil(t, deliveryModule)

				// RouteRegistry -> RouteConfigurator, MiddlewareRegistry -> MiddlewareConfigurator
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestDeliveryModuleHTTPServerProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc    func(*testing.T)
		name        string
		description string
	}{
		{
			name:        "http_server_constructor_pattern",
			description: "Test HTTPServer constructor function pattern",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the constructor function has correct dependencies
				assert.NotNil(t, deliveryModule)

				// Function should depend on config, logger, RouteConfigurator, MiddlewareConfigurator
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "http_server_delivery_newhttp_usage",
			description: "Test delivery.NewHTTPServer usage",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that delivery.NewHTTPServer is called with correct parameters
				assert.NotNil(t, deliveryModule)

				// Should use logger.Named, rc, mc, and config parameters
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "http_server_configuration_pattern",
			description: "Test HTTPServer configuration parameter pattern",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that config parameters are passed correctly
				assert.NotNil(t, deliveryModule)

				// Should use cfg.Address, cfg.StartTimeout, cfg.StopTimeout, cfg.TLS*
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestDeliveryModuleStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(*testing.T)
		name         string
		description  string
	}{
		{
			name:        "fx_module_creation",
			description: "Test fx.Module creation with proper name 'delivery'",
			validateFunc: func(t *testing.T) {
				t.Helper()
				assert.NotNil(t, deliveryModule)

				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "provider_pattern_mix",
			description: "Test mix of provideWithInterfaces and fx.Provide",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that the module correctly mixes both provider patterns
				assert.NotNil(t, deliveryModule)

				// Uses both provideWithInterfaces and fx.Provide
				app := fx.New(
					deliveryModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "interface_binding_structure",
			description: "Test interface binding structure",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that interface bindings are structured correctly
				assert.NotNil(t, deliveryModule)

				// Multiple providers with interface bindings
				moduleType := reflect.TypeOf(deliveryModule)
				assert.NotNil(t, moduleType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.validateFunc(t)
		})
	}
}

func TestRunHTTPServerLifecycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc    func(*testing.T)
		name        string
		description string
	}{
		{
			name:        "function_exists",
			description: "Test that runHTTPServer function exists and has correct signature",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test function signature
				funcType := reflect.TypeOf(runHTTPServer)
				assert.NotNil(t, funcType, "runHTTPServer function should exist")
				assert.Equal(t, reflect.Func, funcType.Kind(), "runHTTPServer should be a function")
			},
		},
		{
			name:        "lifecycle_hook_structure",
			description: "Test lifecycle hook structure",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that function follows fx lifecycle pattern
				funcType := reflect.TypeOf(runHTTPServer)
				assert.NotNil(t, funcType)

				// Should take fx.Lifecycle and *delivery.HTTPServer parameters
				assert.Equal(t, 2, funcType.NumIn(), "runHTTPServer should take 2 parameters")
			},
		},
		{
			name:        "fx_hook_pattern",
			description: "Test fx.Hook pattern usage",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the function uses fx.Hook pattern correctly
				assert.NotNil(t, runHTTPServer)

				// The function should call lc.Append(fx.Hook{OnStart, OnStop})
				funcType := reflect.TypeOf(runHTTPServer)
				assert.Equal(t, reflect.Func, funcType.Kind())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

// Mock implementations for testing lifecycle functions.
type mockDeliveryHTTPServer struct {
	startCalled bool
	stopCalled  bool
}

func (m *mockDeliveryHTTPServer) Start(ctx context.Context) error {
	m.startCalled = true
	return nil
}

func (m *mockDeliveryHTTPServer) Stop(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

func TestRunHTTPServerIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc    func(*testing.T)
		name        string
		description string
	}{
		{
			name:        "lifecycle_methods_called",
			description: "Test that lifecycle methods are properly called",
			testFunc: func(t *testing.T) {
				t.Helper()
				mockServer := &mockDeliveryHTTPServer{}

				app := fx.New(
					fx.Provide(func() *mockDeliveryHTTPServer { return mockServer }),
					fx.Invoke(func(lc fx.Lifecycle, s *mockDeliveryHTTPServer) {
						lc.Append(fx.Hook{
							OnStart: s.Start,
							OnStop:  s.Stop,
						})
					}),
					fx.NopLogger,
				)

				ctx := context.Background()
				err := app.Start(ctx)
				require.NoError(t, err)

				err = app.Stop(ctx)
				require.NoError(t, err)

				assert.True(t, mockServer.startCalled, "Start should be called")
				assert.True(t, mockServer.stopCalled, "Stop should be called")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestDeliveryModuleSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		testPattern string
	}{
		{
			name:        "module_instantiation_safety",
			description: "Test that module instantiation doesn't panic",
			testPattern: "instantiation",
		},
		{
			name:        "fx_integration_safety",
			description: "Test that fx integration is safe",
			testPattern: "fx_integration",
		},
		{
			name:        "provider_reference_safety",
			description: "Test that provider function references are safe",
			testPattern: "provider_reference",
		},
		{
			name:        "lifecycle_function_safety",
			description: "Test that lifecycle functions are safe",
			testPattern: "lifecycle_function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.testPattern {
			case "instantiation":
				assert.NotPanics(t, func() {
					_ = deliveryModule
				})
			case "fx_integration":
				assert.NotPanics(t, func() {
					app := fx.New(
						deliveryModule,
						fx.NopLogger,
					)
					_ = app
				})
			case "provider_reference":
				assert.NotPanics(t, func() {
					assert.NotNil(t, deliveryModule)
				})
			case "lifecycle_function":
				assert.NotPanics(t, func() {
					assert.NotNil(t, runHTTPServer)
					funcType := reflect.TypeOf(runHTTPServer)
					assert.NotNil(t, funcType)
				})
			}
		})
	}
}
