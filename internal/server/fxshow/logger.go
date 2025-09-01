package fxshow

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/config"
	"github.com/gdyunin/aegis-vault-keeper/pkg/logging"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// loggerModule provides logger dependencies.
// Configures structured logging with appropriate levels.
var loggerModule = fx.Module("logger",
	fx.Provide(
		func(cfg *config.LoggerConfig) *zap.SugaredLogger {
			return logging.NewLogger(cfg.Level)
		},
	),
)
