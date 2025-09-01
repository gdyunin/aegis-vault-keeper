package fxshow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// Test interfaces and implementations for testing.
type TestInterface1 interface {
	Method1() string
}

type TestInterface2 interface {
	Method2() int
}

type TestImpl struct {
	value string
}

func (t *TestImpl) Method1() string {
	return t.value
}

func (t *TestImpl) Method2() int {
	return len(t.value)
}

func NewTestImpl(value string) *TestImpl {
	return &TestImpl{value: value}
}

func TestProvideWithInterfaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		constructor  any
		validateFunc func(t *testing.T, option fx.Option)
		name         string
		interfaces   []any
		expectPanic  bool
	}{
		{
			name:        "single_interface",
			constructor: NewTestImpl,
			interfaces:  []any{new(TestInterface1)},
			expectPanic: false,
			validateFunc: func(t *testing.T, option fx.Option) {
				t.Helper()
				// Validate that the option is not nil
				assert.NotNil(t, option)

				// Check that it's a provide option by examining its type
				optionType := reflect.TypeOf(option)
				assert.NotNil(t, optionType)
			},
		},
		{
			name:        "multiple_interfaces",
			constructor: NewTestImpl,
			interfaces:  []any{new(TestInterface1), new(TestInterface2)},
			expectPanic: false,
			validateFunc: func(t *testing.T, option fx.Option) {
				t.Helper()
				assert.NotNil(t, option)

				optionType := reflect.TypeOf(option)
				assert.NotNil(t, optionType)
			},
		},
		{
			name:        "no_interfaces",
			constructor: NewTestImpl,
			interfaces:  []any{},
			expectPanic: false,
			validateFunc: func(t *testing.T, option fx.Option) {
				t.Helper()
				assert.NotNil(t, option)
			},
		},
		{
			name:        "nil_constructor",
			constructor: nil,
			interfaces:  []any{new(TestInterface1)},
			expectPanic: false, // fx.Provide will handle this, but provideWithInterfaces shouldn't panic
			validateFunc: func(t *testing.T, option fx.Option) {
				t.Helper()
				assert.NotNil(t, option)
			},
		},
		{
			name:         "nil_interface",
			constructor:  NewTestImpl,
			interfaces:   []any{nil},
			expectPanic:  true, // fx.As will panic with nil interface
			validateFunc: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectPanic {
				assert.Panics(t, func() {
					provideWithInterfaces[*TestImpl](tt.constructor, tt.interfaces...)
				})
				return
			}

			var option fx.Option
			assert.NotPanics(t, func() {
				option = provideWithInterfaces[*TestImpl](tt.constructor, tt.interfaces...)
			})

			if tt.validateFunc != nil {
				tt.validateFunc(t, option)
			}
		})
	}
}

func TestProvideWithInterfaces_Integration(t *testing.T) {
	t.Parallel()

	// Integration test that validates the fx.Option actually works with fx.App
	t.Run("valid_integration", func(t *testing.T) {
		t.Parallel()

		var receivedImpl1 TestInterface1
		var receivedImpl2 TestInterface2

		app := fx.New(
			fx.Provide(func() string { return "test_value" }),
			provideWithInterfaces[*TestImpl](
				NewTestImpl,
				new(TestInterface1),
				new(TestInterface2),
			),
			fx.Invoke(func(i1 TestInterface1, i2 TestInterface2) {
				receivedImpl1 = i1
				receivedImpl2 = i2
			}),
			fx.NopLogger,
		)

		ctx := context.Background()
		err := app.Start(ctx)
		require.NoError(t, err)

		err = app.Stop(ctx)
		require.NoError(t, err)

		// Verify that the implementations were injected correctly
		require.NotNil(t, receivedImpl1)
		require.NotNil(t, receivedImpl2)

		assert.Equal(t, "test_value", receivedImpl1.Method1())
		assert.Equal(t, 10, receivedImpl2.Method2()) // len("test_value") = 10
	})

	t.Run("no_interfaces_integration", func(t *testing.T) {
		t.Parallel()

		var receivedImpl *TestImpl

		app := fx.New(
			fx.Provide(func() string { return "no_interfaces_test" }),
			provideWithInterfaces[*TestImpl](
				NewTestImpl,
				// No interfaces provided
			),
			fx.Invoke(func(impl *TestImpl) {
				receivedImpl = impl
			}),
			fx.NopLogger,
		)

		ctx := context.Background()
		err := app.Start(ctx)
		require.NoError(t, err)

		err = app.Stop(ctx)
		require.NoError(t, err)

		// Verify that the concrete implementation was injected
		require.NotNil(t, receivedImpl)
		assert.Equal(t, "no_interfaces_test", receivedImpl.Method1())
		assert.Equal(t, 18, receivedImpl.Method2()) // len("no_interfaces_test") = 18
	})
}

// Benchmark the performance of provideWithInterfaces.
func BenchmarkProvideWithInterfaces(b *testing.B) {
	interfaces := []any{new(TestInterface1), new(TestInterface2)}

	b.ResetTimer()
	for range b.N {
		_ = provideWithInterfaces[*TestImpl](NewTestImpl, interfaces...)
	}
}

func BenchmarkProvideWithInterfaces_NoInterfaces(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_ = provideWithInterfaces[*TestImpl](NewTestImpl)
	}
}

func BenchmarkProvideWithInterfaces_ManyInterfaces(b *testing.B) {
	// Create many interface pointers to test performance with many interfaces
	interfaces := make([]any, 10)
	for i := range interfaces {
		interfaces[i] = new(TestInterface1)
	}

	b.ResetTimer()
	for range b.N {
		_ = provideWithInterfaces[*TestImpl](NewTestImpl, interfaces...)
	}
}
