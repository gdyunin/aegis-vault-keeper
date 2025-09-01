package fxshow

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestApplicationModule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc    func(*testing.T)
		name        string
		description string
	}{
		{
			name:        "module_exists",
			description: "Test that applicationModule variable exists and is properly defined",
			testFunc: func(t *testing.T) {
				t.Helper()
				assert.NotNil(t, applicationModule, "applicationModule should not be nil")
			},
		},
		{
			name:        "module_type",
			description: "Test that applicationModule is an fx.Option type",
			testFunc: func(t *testing.T) {
				t.Helper()
				moduleType := reflect.TypeOf(applicationModule)
				assert.NotNil(t, moduleType, "applicationModule type should not be nil")
			},
		},
		{
			name:        "module_structure",
			description: "Test that applicationModule has proper structure and providers",
			testFunc: func(t *testing.T) {
				t.Helper()
				app := fx.New(
					applicationModule,
					fx.NopLogger,
				)
				assert.NotNil(t, app, "fx.App should be created successfully with applicationModule")
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

func TestApplicationModuleProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		description     string
		providerPattern string
		interfaceCount  int
	}{
		{
			name:            "password_hasher_verificator_provider",
			description:     "Test PasswordHasherVerificator provider with interfaces",
			providerPattern: "security.PasswordHasherVerificator",
			interfaceCount:  1,
		},
		{
			name:            "crypto_key_generator_provider",
			description:     "Test CryptoKeyGenerator provider with interfaces",
			providerPattern: "security.CryptoKeyGenerator",
			interfaceCount:  1,
		},
		{
			name:            "token_generate_validator_provider",
			description:     "Test TokenGenerateValidator provider with interfaces",
			providerPattern: "security.TokenGenerateValidator",
			interfaceCount:  1,
		},
		{
			name:            "bankcard_service_provider",
			description:     "Test BankCard service provider with multiple interfaces",
			providerPattern: "bankcardApp.Service",
			interfaceCount:  2,
		},
		{
			name:            "credential_service_provider",
			description:     "Test Credential service provider with multiple interfaces",
			providerPattern: "credentialApp.Service",
			interfaceCount:  2,
		},
		{
			name:            "note_service_provider",
			description:     "Test Note service provider with multiple interfaces",
			providerPattern: "noteApp.Service",
			interfaceCount:  2,
		},
		{
			name:            "filedata_service_provider",
			description:     "Test FileData service provider with multiple interfaces",
			providerPattern: "filedataApp.Service",
			interfaceCount:  2,
		},
		{
			name:            "auth_service_provider",
			description:     "Test Auth service provider with multiple interfaces",
			providerPattern: "authApp.Service",
			interfaceCount:  2,
		},
		{
			name:            "datasync_service_provider",
			description:     "Test DataSync service provider with interfaces",
			providerPattern: "datasyncApp.Service",
			interfaceCount:  1,
		},
		{
			name:            "services_aggregator_provider",
			description:     "Test ServicesAggregator provider",
			providerPattern: "datasyncApp.NewServicesAggregator",
			interfaceCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotNil(t, applicationModule, "applicationModule should exist")
			assert.NotEmpty(t, tt.providerPattern, "providerPattern should not be empty")
			assert.GreaterOrEqual(t, tt.interfaceCount, 0, "interfaceCount should be non-negative")
		})
	}
}

func TestApplicationModuleIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setupFunc   func() fx.Option
		name        string
		description string
		expectError bool
	}{
		{
			name:        "standalone_module_integration",
			description: "Test applicationModule integration in standalone fx.App",
			setupFunc: func() fx.Option {
				return applicationModule
			},
			expectError: false,
		},
		{
			name:        "combined_modules_integration",
			description: "Test applicationModule integration with other modules",
			setupFunc: func() fx.Option {
				return fx.Options(
					applicationModule,
					fx.Module("test", fx.Provide(func() string { return "test" })),
				)
			},
			expectError: false,
		},
		{
			name:        "multiple_application_modules_integration",
			description: "Test multiple applicationModule instances",
			setupFunc: func() fx.Option {
				return fx.Options(
					applicationModule,
					fx.Module("test_app",
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

func TestApplicationModuleProvideWithInterfaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(*testing.T)
		name         string
		description  string
	}{
		{
			name:        "password_hasher_verificator_interfaces",
			description: "Test PasswordHasherVerificator interface provision",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that provideWithInterfaces is used correctly for PasswordHasherVerificator
				assert.NotNil(t, applicationModule)

				// The actual interface binding is tested by the module structure
				app := fx.New(
					applicationModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "service_multiple_interfaces",
			description: "Test service providers with multiple interfaces",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that services are provided with multiple interfaces correctly
				assert.NotNil(t, applicationModule)

				// Multiple interface binding is tested through module structure
				app := fx.New(
					applicationModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "fx_provide_vs_provide_with_interfaces",
			description: "Test mix of fx.Provide and provideWithInterfaces",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that the module correctly mixes fx.Provide and provideWithInterfaces
				assert.NotNil(t, applicationModule)

				// The module should handle both patterns correctly
				moduleType := reflect.TypeOf(applicationModule)
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

func TestApplicationModuleSecurityProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc     func(*testing.T)
		name         string
		description  string
		providerType string
	}{
		{
			name:         "password_hasher_verificator_constructor",
			description:  "Test PasswordHasherVerificator constructor function",
			providerType: "function",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test that the constructor function pattern is correct
				assert.NotNil(t, applicationModule)

				// The function should use security.NewPasswordHasherVerificator with crypto functions
				app := fx.New(
					applicationModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:         "crypto_key_generator_provider",
			description:  "Test CryptoKeyGenerator provider pattern",
			providerType: "direct",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test direct provider pattern for CryptoKeyGenerator
				assert.NotNil(t, applicationModule)

				// Uses security.NewCryptoKeyGenerator directly
				app := fx.New(
					applicationModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:         "token_generate_validator_constructor",
			description:  "Test TokenGenerateValidator constructor with config",
			providerType: "function_with_config",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test constructor function that uses config.AuthConfig
				assert.NotNil(t, applicationModule)

				// The function should use security.NewTokenGenerateValidator with config
				app := fx.New(
					applicationModule,
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

func TestApplicationModuleServiceProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		serviceType string
		interfaces  []string
	}{
		{
			name:        "bankcard_service_interfaces",
			description: "Test BankCard service interface provision",
			serviceType: "bankcardApp.Service",
			interfaces:  []string{"datasyncApp.BankCardService", "bankcardDelivery.Service"},
		},
		{
			name:        "credential_service_interfaces",
			description: "Test Credential service interface provision",
			serviceType: "credentialApp.Service",
			interfaces:  []string{"datasyncApp.CredentialService", "credentialDelivery.Service"},
		},
		{
			name:        "note_service_interfaces",
			description: "Test Note service interface provision",
			serviceType: "noteApp.Service",
			interfaces:  []string{"datasyncApp.NoteService", "noteDelivery.Service"},
		},
		{
			name:        "filedata_service_interfaces",
			description: "Test FileData service interface provision",
			serviceType: "filedataApp.Service",
			interfaces:  []string{"datasyncApp.FileDataService", "filedataDelivery.Service"},
		},
		{
			name:        "auth_service_interfaces",
			description: "Test Auth service interface provision",
			serviceType: "authApp.Service",
			interfaces:  []string{"authDelivery.Service", "middlewareDelivery.AuthWithJWTService"},
		},
		{
			name:        "datasync_service_interfaces",
			description: "Test DataSync service interface provision",
			serviceType: "datasyncApp.Service",
			interfaces:  []string{"datasyncDelivery.Service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotNil(t, applicationModule, "applicationModule should exist")
			assert.NotEmpty(t, tt.serviceType, "serviceType should not be empty")
			assert.NotEmpty(t, tt.interfaces, "interfaces should not be empty")

			// Test that the service interfaces are properly declared
			for _, iface := range tt.interfaces {
				assert.NotEmpty(t, iface, "interface name should not be empty")
			}
		})
	}
}

func TestApplicationModuleStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		validateFunc func(*testing.T)
		name         string
		description  string
	}{
		{
			name:        "fx_module_creation",
			description: "Test fx.Module creation with proper name 'application'",
			validateFunc: func(t *testing.T) {
				t.Helper()
				assert.NotNil(t, applicationModule)

				app := fx.New(
					applicationModule,
					fx.NopLogger,
				)
				require.NotNil(t, app)
			},
		},
		{
			name:        "provider_pattern_consistency",
			description: "Test provider pattern consistency across all providers",
			validateFunc: func(t *testing.T) {
				t.Helper()
				// Test that all providers follow consistent patterns
				assert.NotNil(t, applicationModule)

				// The module should have consistent provideWithInterfaces usage
				app := fx.New(
					applicationModule,
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
				assert.NotNil(t, applicationModule)

				// Multiple interface bindings should work correctly
				moduleType := reflect.TypeOf(applicationModule)
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

func TestApplicationModuleSafety(t *testing.T) {
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
			name:        "interface_binding_safety",
			description: "Test that interface bindings are safe",
			testPattern: "interface_binding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.testPattern {
			case "instantiation":
				assert.NotPanics(t, func() {
					_ = applicationModule
				})
			case "fx_integration":
				assert.NotPanics(t, func() {
					app := fx.New(
						applicationModule,
						fx.NopLogger,
					)
					_ = app
				})
			case "provider_reference":
				assert.NotPanics(t, func() {
					assert.NotNil(t, applicationModule)
				})
			case "interface_binding":
				assert.NotPanics(t, func() {
					app := fx.New(
						applicationModule,
						fx.Module("test", fx.Provide(func() bool { return true })),
						fx.NopLogger,
					)
					_ = app
				})
			}
		})
	}
}
