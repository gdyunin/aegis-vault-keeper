package fxshow

import (
	"go.uber.org/fx"
)

// BuildApp constructs and configures the complete application using dependency injection.
// Returns a configured fx.App with all modules wired together.
func BuildApp() *fx.App {
	return fx.New(
		configModule,
		loggerModule,
		repositoryModule,
		applicationModule,
		deliveryModule,
		fx.Invoke(
			runDatabaseClient,
			runHTTPServer,
		),
	)
}
