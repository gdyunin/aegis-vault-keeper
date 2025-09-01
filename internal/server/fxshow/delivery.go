package fxshow

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/buildinfo"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/common"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/config"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// deliveryModule provides all HTTP delivery layer dependencies.
// Configures route registry, middleware registry, and HTTP server.
var deliveryModule = fx.Module("delivery",
	provideWithInterfaces[*common.BuildInfoOperator](
		func() *common.BuildInfoOperator {
			return common.NewBuildInfoOperator(buildinfo.Version, buildinfo.Date, buildinfo.Commit)
		},
		new(delivery.BuildInfoOperator),
	),
	provideWithInterfaces[*delivery.RouteRegistry](
		delivery.NewRouteRegistry,
		new(delivery.RouteConfigurator),
	),
	provideWithInterfaces[*delivery.MiddlewareRegistry](
		delivery.NewMiddlewareRegistry,
		new(delivery.MiddlewareConfigurator),
	),
	fx.Provide(
		func(
			cfg *config.DeliveryConfig,
			logger *zap.SugaredLogger,
			rc delivery.RouteConfigurator,
			mc delivery.MiddlewareConfigurator,
		) *delivery.HTTPServer {
			return delivery.NewHTTPServer(
				logger.Named("hhtp-server"),
				rc,
				mc,
				cfg.Address,
				cfg.StartTimeout,
				cfg.StopTimeout,
				cfg.TLSEnabled,
				cfg.TLSCertFile,
				cfg.TLSKeyFile,
			)
		},
	),
)

// runHTTPServer registers HTTP server lifecycle hooks with fx.
func runHTTPServer(lc fx.Lifecycle, s *delivery.HTTPServer) {
	lc.Append(fx.Hook{
		OnStart: s.Start,
		OnStop:  s.Stop,
	})
}
