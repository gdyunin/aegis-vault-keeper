package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		base     func(int) int
		mws      []Middleware[func(int) int]
		input    int
		expected int
	}{
		{
			name:     "no middleware",
			base:     func(x int) int { return x * 2 },
			mws:      nil,
			input:    5,
			expected: 10,
		},
		{
			name: "single middleware",
			base: func(x int) int { return x * 2 },
			mws: []Middleware[func(int) int]{
				func(next func(int) int) func(int) int {
					return func(x int) int { return next(x) + 1 }
				},
			},
			input:    5,
			expected: 11, // (5 * 2) + 1
		},
		{
			name: "multiple middleware",
			base: func(x int) int { return x },
			mws: []Middleware[func(int) int]{
				func(next func(int) int) func(int) int {
					return func(x int) int { return next(x) * 2 }
				},
				func(next func(int) int) func(int) int {
					return func(x int) int { return next(x) + 3 }
				},
			},
			input:    5,
			expected: 16, // ((5 + 3) * 2) = 16 (applied in reverse order)
		},
		{
			name: "complex middleware chain",
			base: func(x int) int { return x },
			mws: []Middleware[func(int) int]{
				func(next func(int) int) func(int) int {
					return func(x int) int { return next(x) + 10 }
				},
				func(next func(int) int) func(int) int {
					return func(x int) int { return next(x) * 3 }
				},
				func(next func(int) int) func(int) int {
					return func(x int) int { return next(x) - 1 }
				},
			},
			input:    2,
			expected: 13, // ((2 - 1) * 3) + 10 = 13
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chained := Chain(tt.base, tt.mws...)
			result := chained(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiddleware_StringFunction(t *testing.T) {
	t.Parallel()

	// Test middleware with string functions
	baseFunc := func(s string) string { return s }

	upperMiddleware := func(next func(string) string) func(string) string {
		return func(s string) string {
			result := next(s)
			return "UPPER(" + result + ")"
		}
	}

	prefixMiddleware := func(next func(string) string) func(string) string {
		return func(s string) string {
			result := next(s)
			return "PREFIX_" + result
		}
	}

	chained := Chain(baseFunc, upperMiddleware, prefixMiddleware)
	result := chained("test")

	// Applied in reverse order: PREFIX_UPPER(test)
	assert.Equal(t, "UPPER(PREFIX_test)", result)
}

func TestMiddleware_EmptyMiddlewareSlice(t *testing.T) {
	t.Parallel()

	baseFunc := func(x int) int { return x * 5 }
	emptyMws := []Middleware[func(int) int]{}

	chained := Chain(baseFunc, emptyMws...)
	result := chained(3)

	assert.Equal(t, 15, result)
}
