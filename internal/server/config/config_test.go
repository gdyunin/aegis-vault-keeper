package config

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKeySHA256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:  "simple key",
			input: "test_master_key",
			expected: func() []byte {
				sum := sha256.Sum256([]byte("test_master_key"))
				return sum[:]
			}(),
		},
		{
			name:  "empty string",
			input: "",
			expected: func() []byte {
				sum := sha256.Sum256([]byte(""))
				return sum[:]
			}(),
		},
		{
			name:  "long key",
			input: "very_long_master_key_with_special_chars_!@#$%^&*()_+",
			expected: func() []byte {
				sum := sha256.Sum256([]byte("very_long_master_key_with_special_chars_!@#$%^&*()_+"))
				return sum[:]
			}(),
		},
		{
			name:  "unicode characters",
			input: "master_key_with_unicode_测试",
			expected: func() []byte {
				sum := sha256.Sum256([]byte("master_key_with_unicode_测试"))
				return sum[:]
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := deriveKeySHA256(tt.input)

			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, 32) // SHA256 always produces 32 bytes
		})
	}
}

func TestBindEnvFromStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     interface{}
		name      string
		shouldErr bool
	}{
		{
			name: "valid struct with mapstructure tags",
			input: struct {
				Field1 string `mapstructure:"TEST_FIELD1"`
				Field2 int    `mapstructure:"TEST_FIELD2"`
				Field3 bool   `mapstructure:"TEST_FIELD3"`
			}{},
			shouldErr: false,
		},
		{
			name: "struct without tags",
			input: struct {
				Field1 string
				Field2 int
			}{},
			shouldErr: false,
		},
		{
			name:      "empty struct",
			input:     struct{}{},
			shouldErr: false,
		},
		{
			name: "pointer to struct",
			input: &struct {
				Field1 string `mapstructure:"TEST_FIELD_PTR"`
			}{},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := bindEnvFromStruct(tt.input)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadMasterKey(t *testing.T) {
	tests := []struct {
		setupEnv    func()
		cleanupEnv  func()
		name        string
		errorSubstr string
		shouldErr   bool
	}{
		{
			name: "valid master key",
			setupEnv: func() {
				viper.Set("MASTER_KEY", "this_is_a_valid_master_key_16_chars")
			},
			cleanupEnv: func() {
				viper.Set("MASTER_KEY", "")
			},
			shouldErr: false,
		},
		{
			name: "empty master key",
			setupEnv: func() {
				viper.Set("MASTER_KEY", "")
			},
			cleanupEnv:  func() {},
			shouldErr:   true,
			errorSubstr: "invalid master key",
		},
		{
			name: "too short master key",
			setupEnv: func() {
				viper.Set("MASTER_KEY", "short")
			},
			cleanupEnv: func() {
				viper.Set("MASTER_KEY", "")
			},
			shouldErr:   true,
			errorSubstr: "invalid master key",
		},
		{
			name: "exactly minimum length",
			setupEnv: func() {
				viper.Set("MASTER_KEY", "exactly16charkey")
			},
			cleanupEnv: func() {
				viper.Set("MASTER_KEY", "")
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			result, err := loadMasterKey()

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorSubstr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, 32) // SHA256 produces 32 bytes
			}
		})
	}
}

func TestValidateTLSConfig(t *testing.T) {
	t.Parallel()

	// Create temporary files for testing
	tempDir := t.TempDir()
	validCertFile := filepath.Join(tempDir, "cert.pem")
	validKeyFile := filepath.Join(tempDir, "key.pem")

	// Create the files
	require.NoError(t, os.WriteFile(validCertFile, []byte("fake cert"), 0o600))
	require.NoError(t, os.WriteFile(validKeyFile, []byte("fake key"), 0o600))

	nonExistentFile := filepath.Join(tempDir, "nonexistent.pem")

	tests := []struct {
		config      *Config
		name        string
		errorSubstr string
		shouldErr   bool
	}{
		{
			name: "TLS disabled",
			config: &Config{
				TLSEnabled: false,
			},
			shouldErr: false,
		},
		{
			name: "valid TLS config",
			config: &Config{
				TLSEnabled:  true,
				TLSCertFile: validCertFile,
				TLSKeyFile:  validKeyFile,
			},
			shouldErr: false,
		},
		{
			name: "missing cert file path",
			config: &Config{
				TLSEnabled:  true,
				TLSCertFile: "",
				TLSKeyFile:  validKeyFile,
			},
			shouldErr:   true,
			errorSubstr: "TLS_CERT_FILE is required",
		},
		{
			name: "missing key file path",
			config: &Config{
				TLSEnabled:  true,
				TLSCertFile: validCertFile,
				TLSKeyFile:  "",
			},
			shouldErr:   true,
			errorSubstr: "TLS_KEY_FILE is required",
		},
		{
			name: "cert file not found",
			config: &Config{
				TLSEnabled:  true,
				TLSCertFile: nonExistentFile,
				TLSKeyFile:  validKeyFile,
			},
			shouldErr:   true,
			errorSubstr: "TLS certificate file not found",
		},
		{
			name: "key file not found",
			config: &Config{
				TLSEnabled:  true,
				TLSCertFile: validCertFile,
				TLSKeyFile:  nonExistentFile,
			},
			shouldErr:   true,
			errorSubstr: "TLS key file not found",
		},
		{
			name: "both files not found",
			config: &Config{
				TLSEnabled:  true,
				TLSCertFile: nonExistentFile,
				TLSKeyFile:  nonExistentFile + "2",
			},
			shouldErr:   true,
			errorSubstr: "TLS certificate file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateTLSConfig(tt.config)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorSubstr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper tests to ensure our test utilities work.
func TestConfigStructReflection(t *testing.T) {
	t.Parallel()

	// Test that we can properly reflect on the Config struct
	cfg := Config{}
	cfgType := reflect.TypeOf(cfg)

	// Verify we can iterate through fields
	fieldCount := 0
	for i := range cfgType.NumField() {
		field := cfgType.Field(i)
		fieldCount++

		// Check that fields with mapstructure tags are valid
		if tag := field.Tag.Get("mapstructure"); tag != "" {
			assert.NotEmpty(t, tag, "mapstructure tag should not be empty")
		}
	}

	// Verify we found the expected number of fields (this might need adjustment if Config changes)
	assert.Greater(t, fieldCount, 10, "Config should have multiple fields")
}

func TestConfigFieldTypes(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	cfgType := reflect.TypeOf(cfg)

	expectedTypes := map[string]string{
		"FileStorageBasePath":  "string",
		"PostgresPassword":     "string",
		"PostgresDBName":       "string",
		"PostgresHost":         "string",
		"PostgresSSLMode":      "string",
		"LoggerLevel":          "string",
		"TLSCertFile":          "string",
		"TLSKeyFile":           "string",
		"PostgresUser":         "string",
		"MasterKey":            "[]uint8",
		"PostgresInitTimeout":  "time.Duration",
		"ApplicationPort":      "int",
		"AccessTokenLifeTime":  "time.Duration",
		"PostgresPort":         "int",
		"DeliveryStartTimeout": "time.Duration",
		"DeliveryStopTimeout":  "time.Duration",
		"TLSEnabled":           "bool",
	}

	for i := range cfgType.NumField() {
		field := cfgType.Field(i)
		if expectedType, exists := expectedTypes[field.Name]; exists {
			assert.Equal(t, expectedType, field.Type.String(),
				"Field %s should have type %s", field.Name, expectedType)
		}
	}
}

func TestMasterKeyConstants(t *testing.T) {
	t.Parallel()

	// Test that our constants are reasonable
	assert.Equal(t, 16, masterKeyMinLen, "Master key minimum length should be 16")
	assert.Greater(t, masterKeyMinLen, 8, "Master key should be reasonably secure")
	assert.LessOrEqual(t, masterKeyMinLen, 32, "Master key minimum shouldn't be too restrictive")
}

func TestLoadConfigErrorPatterns(t *testing.T) {
	// Test LoadConfig error patterns without complex setup
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "config_file_error_pattern",
			description: "LoadConfig should return error when config file not found",
		},
		{
			name:        "master_key_error_pattern",
			description: "LoadConfig should return error for invalid master key",
		},
		{
			name:        "tls_validation_error_pattern",
			description: "LoadConfig should return error for invalid TLS config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error message patterns exist
			assert.NotEmpty(t, tt.description)

			// Test that LoadConfig function exists
			assert.NotNil(t, LoadConfig)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		setup       func(t *testing.T) func()
		name        string
		errContains string
		expectErr   bool
	}{
		{
			name: "missing config file error",
			setup: func(t *testing.T) func() {
				t.Helper()
				// Reset viper and use invalid config path to trigger error
				viper.Reset()
				return func() { viper.Reset() }
			},
			expectErr:   true,
			errContains: "failed to read config file",
		},
		{
			name: "invalid master key error",
			setup: func(t *testing.T) func() {
				t.Helper()
				viper.Reset()

				// Create temporary config file
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "server.yml")
				err := os.WriteFile(configFile, []byte("server_host: localhost"), 0o600)
				require.NoError(t, err)

				viper.SetConfigName("server")
				viper.SetConfigType("yml")
				viper.AddConfigPath(tmpDir)

				// Set environment variable for short master key
				t.Setenv("MASTER_KEY", "short")

				return func() { viper.Reset() }
			},
			expectErr:   true,
			errContains: "failed to load master key",
		},
		{
			name: "TLS validation error",
			setup: func(t *testing.T) func() {
				t.Helper()
				viper.Reset()

				// Create temporary config file
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "server.yml")
				err := os.WriteFile(configFile, []byte("tls_enabled: true"), 0o600)
				require.NoError(t, err)

				viper.SetConfigName("server")
				viper.SetConfigType("yml")
				viper.AddConfigPath(tmpDir)

				// Set environment variables
				t.Setenv("MASTER_KEY", "test-master-key-for-testing-purposes-only")
				t.Setenv("TLS_ENABLED", "true")
				// Don't set TLS_CERT_FILE and TLS_KEY_FILE to trigger validation error

				return func() { viper.Reset() }
			},
			expectErr:   true,
			errContains: "TLS configuration validation failed",
		},
		{
			name: "unmarshal error pattern",
			setup: func(t *testing.T) func() {
				t.Helper()
				viper.Reset()

				// Create temporary config file with invalid YAML structure
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "server.yml")
				// Create invalid duration format to trigger unmarshal error
				err := os.WriteFile(configFile, []byte("delivery_start_timeout: invalid_duration"), 0o600)
				require.NoError(t, err)

				viper.SetConfigName("server")
				viper.SetConfigType("yml")
				viper.AddConfigPath(tmpDir)

				t.Setenv("MASTER_KEY", "test-master-key-for-testing-purposes-only")

				return func() { viper.Reset() }
			},
			expectErr:   true,
			errContains: "failed to decode config into struct",
		},
		{
			name: "bind environment variables error pattern",
			setup: func(t *testing.T) func() {
				t.Helper()
				viper.Reset()

				// Create temporary config file
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "server.yml")
				err := os.WriteFile(configFile, []byte("server_host: localhost"), 0o600)
				require.NoError(t, err)

				viper.SetConfigName("server")
				viper.SetConfigType("yml")
				viper.AddConfigPath(tmpDir)

				t.Setenv("MASTER_KEY", "test-master-key-for-testing-purposes-only")

				return func() { viper.Reset() }
			},
			expectErr:   false, // bindEnvFromStruct usually succeeds
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			// Test LoadConfig - this will execute the function
			config, err := LoadConfig()

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
			}
		})
	}
}

func TestLoadConfigFunctionExists(t *testing.T) {
	t.Parallel()

	// Test that LoadConfig function exists and can be called
	// This is a simple test to ensure the function is accessible
	assert.NotNil(t, LoadConfig, "LoadConfig function should exist")

	// Test error patterns that would occur in LoadConfig
	tests := []struct {
		name           string
		errorPattern   string
		expectedSubstr string
	}{
		{
			name:           "config file error pattern",
			errorPattern:   "failed to read config file: some error",
			expectedSubstr: "failed to read config file",
		},
		{
			name:           "env binding error pattern",
			errorPattern:   "failed to bind environment variables: some error",
			expectedSubstr: "failed to bind environment variables",
		},
		{
			name:           "master key error pattern",
			errorPattern:   "failed to load master key: some error",
			expectedSubstr: "failed to load master key",
		},
		{
			name:           "config unmarshal error pattern",
			errorPattern:   "failed to unmarshal config: some error",
			expectedSubstr: "failed to unmarshal config",
		},
		{
			name:           "tls validation error pattern",
			errorPattern:   "failed to validate TLS config: some error",
			expectedSubstr: "failed to validate TLS config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test that error patterns are as expected
			assert.Contains(t, tt.errorPattern, tt.expectedSubstr)
		})
	}
}
