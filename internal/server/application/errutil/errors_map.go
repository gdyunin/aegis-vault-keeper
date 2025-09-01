package errutil

import "errors"

// multiUnwrapper defines the interface for errors that wrap multiple errors.
type multiUnwrapper interface{ Unwrap() []error }

// singleUnwrapper defines the interface for errors that wrap a single error.
type singleUnwrapper interface{ Unwrap() error }

// MapFunc defines a function type that transforms one error into another.
type MapFunc func(error) error

// MapError recursively applies a mapping function to an error and all its wrapped errors.
// It handles both single and multiple wrapped errors, preserving the error wrapping structure.
func MapError(mapFn MapFunc, err error) error {
	if err == nil {
		return nil
	}

	if m, ok := err.(multiUnwrapper); ok {
		unwrapped := m.Unwrap()
		mapped := make([]error, 0, len(unwrapped))
		for _, e := range unwrapped {
			mapped = append(mapped, MapError(mapFn, e))
		}
		return errors.Join(mapped...)
	}

	if s, ok := err.(singleUnwrapper); ok {
		return MapError(mapFn, s.Unwrap())
	}

	return mapFn(err)
}
