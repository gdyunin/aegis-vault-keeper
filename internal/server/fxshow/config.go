package fxshow

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/config"
	"go.uber.org/fx"
)

// configModule provides all configuration-related dependencies.
// Loads configuration and extracts specialized config structs for different components.
var configModule = fx.Module("config",
	fx.Provide(
		config.LoadConfig,
		config.ExtractAuthConfig,
		config.ExtractDBConfig,
		config.ExtractLoggerConfig,
		config.ExtractDeliveryConfig,
		config.ExtractFileStorageConfig,
	),
)
