package fxshow

import (
	"go.uber.org/fx"
)

// provideWithInterfaces registers a constructor with the container and declares which interfaces it implements.
// This generic helper reduces boilerplate when a type implements multiple interfaces.
func provideWithInterfaces[Impl any](constructor any, interfaces ...any) fx.Option {
	asOptions := make([]fx.Annotation, len(interfaces))
	for i, iface := range interfaces {
		asOptions[i] = fx.As(iface)
	}

	return fx.Provide(
		fx.Annotate(constructor, asOptions...),
	)
}
