package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractDBConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config   *Config
		expected *DBConfig
		name     string
	}{
		{
			name: "complete database config",
			config: &Config{
				PostgresHost:        "localhost",
				PostgresUser:        "testuser",
				PostgresPassword:    "testpass",
				PostgresDBName:      "testdb",
				PostgresSSLMode:     "disable",
				PostgresPort:        5432,
				PostgresInitTimeout: 30 * time.Second,
			},
			expected: &DBConfig{
				Host:     "localhost",
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "disable",
				Port:     5432,
				Timeout:  30 * time.Second,
			},
		},
		{
			name:   "empty database config",
			config: &Config{},
			expected: &DBConfig{
				Host:     "",
				User:     "",
				Password: "",
				DBName:   "",
				SSLMode:  "",
				Port:     0,
				Timeout:  0,
			},
		},
		{
			name: "production-like database config",
			config: &Config{
				PostgresHost:        "db.example.com",
				PostgresUser:        "produser",
				PostgresPassword:    "verysecurepassword",
				PostgresDBName:      "aegis_vault_keeper_prod",
				PostgresSSLMode:     "require",
				PostgresPort:        5432,
				PostgresInitTimeout: 60 * time.Second,
			},
			expected: &DBConfig{
				Host:     "db.example.com",
				User:     "produser",
				Password: "verysecurepassword",
				DBName:   "aegis_vault_keeper_prod",
				SSLMode:  "require",
				Port:     5432,
				Timeout:  60 * time.Second,
			},
		},
		{
			name: "ssl modes variations",
			config: &Config{
				PostgresHost:    "localhost",
				PostgresSSLMode: "verify-full",
				PostgresPort:    5433,
			},
			expected: &DBConfig{
				Host:    "localhost",
				SSLMode: "verify-full",
				Port:    5433,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractDBConfig(tt.config)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractLoggerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config   *Config
		expected *LoggerConfig
		name     string
	}{
		{
			name: "debug level",
			config: &Config{
				LoggerLevel: "debug",
			},
			expected: &LoggerConfig{
				Level: "debug",
			},
		},
		{
			name: "info level",
			config: &Config{
				LoggerLevel: "info",
			},
			expected: &LoggerConfig{
				Level: "info",
			},
		},
		{
			name: "warn level",
			config: &Config{
				LoggerLevel: "warn",
			},
			expected: &LoggerConfig{
				Level: "warn",
			},
		},
		{
			name: "error level",
			config: &Config{
				LoggerLevel: "error",
			},
			expected: &LoggerConfig{
				Level: "error",
			},
		},
		{
			name: "empty level",
			config: &Config{
				LoggerLevel: "",
			},
			expected: &LoggerConfig{
				Level: "",
			},
		},
		{
			name: "uppercase level",
			config: &Config{
				LoggerLevel: "DEBUG",
			},
			expected: &LoggerConfig{
				Level: "DEBUG",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractLoggerConfig(tt.config)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAuthConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config   *Config
		expected *AuthConfig
		name     string
	}{
		{
			name: "complete auth config",
			config: &Config{
				MasterKey:           []byte("secret_master_key_32_bytes_long"),
				AccessTokenLifeTime: 24 * time.Hour,
			},
			expected: &AuthConfig{
				MasterKey:           []byte("secret_master_key_32_bytes_long"),
				AccessTokenLifeTime: 24 * time.Hour,
			},
		},
		{
			name:   "empty auth config",
			config: &Config{},
			expected: &AuthConfig{
				MasterKey:           nil,
				AccessTokenLifeTime: 0,
			},
		},
		{
			name: "short token lifetime",
			config: &Config{
				MasterKey:           []byte("key"),
				AccessTokenLifeTime: 15 * time.Minute,
			},
			expected: &AuthConfig{
				MasterKey:           []byte("key"),
				AccessTokenLifeTime: 15 * time.Minute,
			},
		},
		{
			name: "long token lifetime",
			config: &Config{
				MasterKey:           []byte("test_key"),
				AccessTokenLifeTime: 7 * 24 * time.Hour,
			},
			expected: &AuthConfig{
				MasterKey:           []byte("test_key"),
				AccessTokenLifeTime: 7 * 24 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractAuthConfig(tt.config)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractDeliveryConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config   *Config
		expected *DeliveryConfig
		name     string
	}{
		{
			name: "complete delivery config with TLS",
			config: &Config{
				ApplicationPort:      8080,
				DeliveryStartTimeout: 30 * time.Second,
				DeliveryStopTimeout:  10 * time.Second,
				TLSEnabled:           true,
				TLSCertFile:          "/path/to/cert.pem",
				TLSKeyFile:           "/path/to/key.pem",
			},
			expected: &DeliveryConfig{
				Address:      ":8080",
				StartTimeout: 30 * time.Second,
				StopTimeout:  10 * time.Second,
				TLSEnabled:   true,
				TLSCertFile:  "/path/to/cert.pem",
				TLSKeyFile:   "/path/to/key.pem",
			},
		},
		{
			name: "HTTP only config",
			config: &Config{
				ApplicationPort:      3000,
				DeliveryStartTimeout: 15 * time.Second,
				DeliveryStopTimeout:  5 * time.Second,
				TLSEnabled:           false,
			},
			expected: &DeliveryConfig{
				Address:      ":3000",
				StartTimeout: 15 * time.Second,
				StopTimeout:  5 * time.Second,
				TLSEnabled:   false,
				TLSCertFile:  "",
				TLSKeyFile:   "",
			},
		},
		{
			name: "default port zero",
			config: &Config{
				ApplicationPort: 0,
			},
			expected: &DeliveryConfig{
				Address:      ":0",
				StartTimeout: 0,
				StopTimeout:  0,
				TLSEnabled:   false,
				TLSCertFile:  "",
				TLSKeyFile:   "",
			},
		},
		{
			name: "high port number",
			config: &Config{
				ApplicationPort: 65535,
			},
			expected: &DeliveryConfig{
				Address:      ":65535",
				StartTimeout: 0,
				StopTimeout:  0,
				TLSEnabled:   false,
				TLSCertFile:  "",
				TLSKeyFile:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractDeliveryConfig(tt.config)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFileStorageConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config   *Config
		expected *FileStorageConfig
		name     string
	}{
		{
			name: "absolute path",
			config: &Config{
				FileStorageBasePath: "/var/data/aegis_vault_keeper",
			},
			expected: &FileStorageConfig{
				BasePath: "/var/data/aegis_vault_keeper",
			},
		},
		{
			name: "relative path",
			config: &Config{
				FileStorageBasePath: "./data",
			},
			expected: &FileStorageConfig{
				BasePath: "./data",
			},
		},
		{
			name: "empty path",
			config: &Config{
				FileStorageBasePath: "",
			},
			expected: &FileStorageConfig{
				BasePath: "",
			},
		},
		{
			name: "home directory path",
			config: &Config{
				FileStorageBasePath: "~/aegis_vault_keeper",
			},
			expected: &FileStorageConfig{
				BasePath: "~/aegis_vault_keeper",
			},
		},
		{
			name: "windows path",
			config: &Config{
				FileStorageBasePath: "C:\\data\\aegis_vault_keeper",
			},
			expected: &FileStorageConfig{
				BasePath: "C:\\data\\aegis_vault_keeper",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractFileStorageConfig(tt.config)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractedConfigStructures(t *testing.T) {
	t.Parallel()

	// Test that all extraction functions create independent structures
	mainConfig := &Config{
		PostgresHost:         "localhost",
		PostgresUser:         "user",
		PostgresPassword:     "pass",
		PostgresDBName:       "db",
		PostgresSSLMode:      "disable",
		PostgresPort:         5432,
		PostgresInitTimeout:  30 * time.Second,
		LoggerLevel:          "info",
		MasterKey:            []byte("test_key"),
		AccessTokenLifeTime:  time.Hour,
		ApplicationPort:      8080,
		DeliveryStartTimeout: 30 * time.Second,
		DeliveryStopTimeout:  10 * time.Second,
		TLSEnabled:           true,
		TLSCertFile:          "cert.pem",
		TLSKeyFile:           "key.pem",
		FileStorageBasePath:  "/data",
	}

	// Extract all configs
	dbConfig := ExtractDBConfig(mainConfig)
	loggerConfig := ExtractLoggerConfig(mainConfig)
	authConfig := ExtractAuthConfig(mainConfig)
	deliveryConfig := ExtractDeliveryConfig(mainConfig)
	fileStorageConfig := ExtractFileStorageConfig(mainConfig)

	// Verify all are non-nil
	assert.NotNil(t, dbConfig)
	assert.NotNil(t, loggerConfig)
	assert.NotNil(t, authConfig)
	assert.NotNil(t, deliveryConfig)
	assert.NotNil(t, fileStorageConfig)

	// Verify independence by modifying extracted configs
	dbConfig.Host = "modified"
	loggerConfig.Level = "modified"
	authConfig.AccessTokenLifeTime = time.Minute
	deliveryConfig.Address = "modified"
	fileStorageConfig.BasePath = "modified"

	// Original config should be unchanged
	assert.Equal(t, "localhost", mainConfig.PostgresHost)
	assert.Equal(t, "info", mainConfig.LoggerLevel)
	assert.Equal(t, time.Hour, mainConfig.AccessTokenLifeTime)
	assert.Equal(t, 8080, mainConfig.ApplicationPort)
	assert.Equal(t, "/data", mainConfig.FileStorageBasePath)

	// Re-extract to verify original values are preserved
	newDBConfig := ExtractDBConfig(mainConfig)
	assert.Equal(t, "localhost", newDBConfig.Host)
	assert.NotEqual(t, "modified", newDBConfig.Host)
}
