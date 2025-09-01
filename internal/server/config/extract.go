package config

import (
	"strconv"
	"time"
)

// DBConfig contains database connection configuration extracted from the main config.
type DBConfig struct {
	// Host specifies the PostgreSQL server hostname or IP address.
	Host string
	// User specifies the database username for authentication.
	User string
	// Password contains the database user password (sensitive data).
	Password string
	// DBName specifies the target PostgreSQL database name.
	DBName string
	// SSLMode specifies the SSL connection mode (disable, require, verify-ca, verify-full).
	SSLMode string
	// Port specifies the PostgreSQL server port number.
	Port int
	// Timeout specifies the maximum duration for database initialization.
	Timeout time.Duration
}

// ExtractDBConfig extracts database-specific configuration from the main config.
func ExtractDBConfig(cfg *Config) *DBConfig {
	return &DBConfig{
		Host:     cfg.PostgresHost,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		DBName:   cfg.PostgresDBName,
		SSLMode:  cfg.PostgresSSLMode,
		Port:     cfg.PostgresPort,
		Timeout:  cfg.PostgresInitTimeout,
	}
}

// LoggerConfig contains logging configuration extracted from the main config.
type LoggerConfig struct {
	// Level specifies the logging level (debug, info, warn, error).
	Level string
}

// ExtractLoggerConfig extracts logging-specific configuration from the main config.
func ExtractLoggerConfig(cfg *Config) *LoggerConfig {
	return &LoggerConfig{
		Level: cfg.LoggerLevel,
	}
}

// AuthConfig contains authentication configuration extracted from the main config.
type AuthConfig struct {
	// MasterKey contains the derived encryption key for data protection (highly sensitive).
	MasterKey []byte
	// AccessTokenLifeTime specifies the JWT token validity duration.
	AccessTokenLifeTime time.Duration
}

// ExtractAuthConfig extracts authentication-specific configuration from the main config.
func ExtractAuthConfig(cfg *Config) *AuthConfig {
	return &AuthConfig{
		MasterKey:           cfg.MasterKey,
		AccessTokenLifeTime: cfg.AccessTokenLifeTime,
	}
}

// DeliveryConfig contains HTTP server configuration extracted from the main config.
type DeliveryConfig struct {
	// Address specifies the HTTP server listening address and port.
	Address string
	// TLSCertFile specifies the path to the TLS certificate file.
	TLSCertFile string
	// TLSKeyFile specifies the path to the TLS private key file.
	TLSKeyFile string
	// StartTimeout specifies the maximum duration for HTTP server startup.
	StartTimeout time.Duration
	// StopTimeout specifies the maximum duration for HTTP server shutdown.
	StopTimeout time.Duration
	// TLSEnabled determines whether HTTPS should be used instead of HTTP.
	TLSEnabled bool
}

// ExtractDeliveryConfig extracts HTTP delivery-specific configuration from the main config.
func ExtractDeliveryConfig(cfg *Config) *DeliveryConfig {
	return &DeliveryConfig{
		Address:      ":" + strconv.Itoa(cfg.ApplicationPort),
		StartTimeout: cfg.DeliveryStartTimeout,
		StopTimeout:  cfg.DeliveryStopTimeout,
		TLSEnabled:   cfg.TLSEnabled,
		TLSCertFile:  cfg.TLSCertFile,
		TLSKeyFile:   cfg.TLSKeyFile,
	}
}

// FileStorageConfig contains file storage configuration extracted from the main config.
type FileStorageConfig struct {
	// BasePath specifies the base directory for file storage operations.
	BasePath string
}

// ExtractFileStorageConfig extracts file storage-specific configuration from the main config.
func ExtractFileStorageConfig(cfg *Config) *FileStorageConfig {
	return &FileStorageConfig{
		BasePath: cfg.FileStorageBasePath,
	}
}
