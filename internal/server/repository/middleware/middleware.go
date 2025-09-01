package middleware

// Middleware defines a generic middleware function that wraps another function of type T.
// It follows the decorator pattern to add cross-cutting concerns like logging, encryption, etc.
type Middleware[T any] func(next T) T

// Chain applies multiple middleware functions to a base function in reverse order.
// The last middleware in the slice will be applied first (closest to the base function).
func Chain[T any](f T, mws ...Middleware[T]) T {
	for i := len(mws) - 1; i >= 0; i-- {
		f = mws[i](f)
	}
	return f
}
