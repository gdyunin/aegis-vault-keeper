package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// masterKeyMinLen defines the minimum required length for the master encryption key.
const masterKeyMinLen = 16

// Config contains all configuration parameters for the AegisVaultKeeper server application.
type Config struct {
	// FileStorageBasePath specifies the base directory for file storage operations.
	FileStorageBasePath string `mapstructure:"FILE_STORAGE_BASE_PATH"`
	// PostgresPassword contains the database user password (sensitive data).
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	// PostgresDBName specifies the target PostgreSQL database name.
	PostgresDBName string `mapstructure:"POSTGRES_DB_NAME"`
	// PostgresHost specifies the PostgreSQL server hostname or IP address.
	PostgresHost string `mapstructure:"POSTGRES_HOST"`
	// PostgresSSLMode specifies the SSL connection mode (disable, require, verify-ca, verify-full).
	PostgresSSLMode string `mapstructure:"POSTGRES_SSL_MODE"`
	// LoggerLevel specifies the logging level (debug, info, warn, error).
	LoggerLevel string `mapstructure:"LOGGER_LEVEL"`
	// TLSCertFile specifies the path to the TLS certificate file.
	TLSCertFile string `mapstructure:"TLS_CERT_FILE"`
	// TLSKeyFile specifies the path to the TLS private key file.
	TLSKeyFile string `mapstructure:"TLS_KEY_FILE"`
	// PostgresUser specifies the database username for authentication.
	PostgresUser string `mapstructure:"POSTGRES_USER"`
	// MasterKey contains the derived encryption key for data protection (highly sensitive).
	MasterKey []byte
	// PostgresInitTimeout specifies the maximum duration for database initialization.
	PostgresInitTimeout time.Duration `mapstructure:"POSTGRES_INIT_TIMEOUT"`
	// ApplicationPort specifies the HTTP server listening port.
	ApplicationPort int `mapstructure:"APPLICATION_PORT"`
	// AccessTokenLifeTime specifies the JWT token validity duration.
	AccessTokenLifeTime time.Duration `mapstructure:"ACCESS_TOKEN_LIFETIME"`
	// PostgresPort specifies the PostgreSQL server port number.
	PostgresPort int `mapstructure:"POSTGRES_PORT"`
	// DeliveryStartTimeout specifies the maximum duration for HTTP server startup.
	DeliveryStartTimeout time.Duration `mapstructure:"DELIVERY_START_TIMEOUT"`
	// DeliveryStopTimeout specifies the maximum duration for HTTP server shutdown.
	DeliveryStopTimeout time.Duration `mapstructure:"DELIVERY_STOP_TIMEOUT"`
	// TLSEnabled determines whether HTTPS should be used instead of HTTP.
	TLSEnabled bool `mapstructure:"TLS_ENABLED"`
}

// LoadConfig loads and validates the server configuration from environment variables and files.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("server")
	viper.SetConfigType("yml")
	viper.AddConfigPath("/app/config")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := bindEnvFromStruct(Config{}); err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// cfg holds the unmarshalled configuration structure.
	var cfg Config
	if err := viper.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.StringToTimeDurationHookFunc(),
	)); err != nil {
		return nil, fmt.Errorf("failed to decode config into struct: %w", err)
	}

	mk, err := loadMasterKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load master key: %w", err)
	}
	cfg.MasterKey = mk

	if err := validateTLSConfig(&cfg); err != nil {
		return nil, fmt.Errorf("TLS configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// bindEnvFromStruct binds environment variables to viper based on mapstructure tags in the struct.
func bindEnvFromStruct(structType interface{}) error {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := range t.NumField() {
		field := t.Field(i)
		if key := field.Tag.Get("mapstructure"); key != "" {
			if err := viper.BindEnv(key); err != nil {
				return fmt.Errorf("failed to bind environment variable %q: %w", key, err)
			}
		}
	}
	return nil
}

// loadMasterKey loads and validates the master encryption key from environment variables.
// The key is validated for minimum length before being derived using SHA256.
func loadMasterKey() ([]byte, error) {
	masterKey := viper.GetString("MASTER_KEY")
	if masterKey == "" || len(masterKey) < masterKeyMinLen {
		return nil, fmt.Errorf("invalid master key: it must be at least %d characters long", masterKeyMinLen)
	}
	return deriveKeySHA256(masterKey), nil
}

// deriveKeySHA256 derives a 32-byte encryption key from the master key using SHA256.
func deriveKeySHA256(masterKey string) []byte {
	sum := sha256.Sum256([]byte(masterKey))
	return sum[:]
}

// validateTLSConfig validates TLS configuration when TLS is enabled.
// Checks that required certificate and key files are specified and exist.
func validateTLSConfig(cfg *Config) error {
	if !cfg.TLSEnabled {
		return nil
	}

	if cfg.TLSCertFile == "" {
		return errors.New("TLS_CERT_FILE is required when TLS is enabled")
	}

	if cfg.TLSKeyFile == "" {
		return errors.New("TLS_KEY_FILE is required when TLS is enabled")
	}

	if _, err := os.Stat(cfg.TLSCertFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS certificate file not found: %s", cfg.TLSCertFile)
	}

	if _, err := os.Stat(cfg.TLSKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS key file not found: %s", cfg.TLSKeyFile)
	}

	return nil
}
