package errutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test custom error types that implement unwrapping interfaces.
type testSingleUnwrapError struct {
	err error
	msg string
}

func (e *testSingleUnwrapError) Error() string {
	return e.msg
}

func (e *testSingleUnwrapError) Unwrap() error {
	return e.err
}

type testMultiUnwrapError struct {
	msg  string
	errs []error
}

func (e *testMultiUnwrapError) Error() string {
	return e.msg
}

func (e *testMultiUnwrapError) Unwrap() []error {
	return e.errs
}

func TestMapError(t *testing.T) {
	t.Parallel()

	baseErr1 := errors.New("base error 1")
	baseErr2 := errors.New("base error 2")
	baseErr3 := errors.New("base error 3")

	// Test mapping function that adds prefix
	addPrefix := func(err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("mapped: %w", err)
	}

	// Test mapping function that transforms to specific error
	transformToCustom := func(err error) error {
		if err == nil {
			return nil
		}
		return errors.New("custom error: " + err.Error())
	}

	// Test mapping function that returns nil for specific errors
	filterSpecific := func(err error) error {
		if err != nil && err.Error() == "base error 2" {
			return nil
		}
		return err
	}

	type args struct {
		mapFn MapFunc
		err   error
	}
	tests := []struct {
		args args
		want error
		name string
	}{
		{
			name: "nil_error_returns_nil",
			args: args{
				mapFn: addPrefix,
				err:   nil,
			},
			want: nil,
		},
		{
			name: "simple_error_mapping",
			args: args{
				mapFn: addPrefix,
				err:   baseErr1,
			},
			want: fmt.Errorf("mapped: %w", baseErr1),
		},
		{
			name: "transform_to_custom_error",
			args: args{
				mapFn: transformToCustom,
				err:   baseErr1,
			},
			want: errors.New("custom error: base error 1"),
		},
		{
			name: "single_wrapped_error",
			args: args{
				mapFn: addPrefix,
				err: &testSingleUnwrapError{
					msg: "wrapper error",
					err: baseErr1,
				},
			},
			want: fmt.Errorf("mapped: %w", baseErr1),
		},
		{
			name: "multiple_wrapped_errors",
			args: args{
				mapFn: addPrefix,
				err: &testMultiUnwrapError{
					msg:  "multi wrapper error",
					errs: []error{baseErr1, baseErr2},
				},
			},
			want: errors.Join(
				fmt.Errorf("mapped: %w", baseErr1),
				fmt.Errorf("mapped: %w", baseErr2),
			),
		},
		{
			name: "nested_single_wrapped_errors",
			args: args{
				mapFn: addPrefix,
				err: &testSingleUnwrapError{
					msg: "outer wrapper",
					err: &testSingleUnwrapError{
						msg: "inner wrapper",
						err: baseErr1,
					},
				},
			},
			want: fmt.Errorf("mapped: %w", baseErr1),
		},
		{
			name: "filter_specific_error",
			args: args{
				mapFn: filterSpecific,
				err:   baseErr2,
			},
			want: nil,
		},
		{
			name: "filter_in_multiple_wrapped_errors",
			args: args{
				mapFn: filterSpecific,
				err: &testMultiUnwrapError{
					msg:  "multi wrapper",
					errs: []error{baseErr1, baseErr2, baseErr3},
				},
			},
			want: errors.Join(baseErr1, nil, baseErr3),
		},
		{
			name: "empty_multiple_wrapped_errors",
			args: args{
				mapFn: addPrefix,
				err: &testMultiUnwrapError{
					msg:  "empty multi wrapper",
					errs: []error{},
				},
			},
			want: errors.Join(),
		},
		{
			name: "multiple_wrapped_with_nil_errors",
			args: args{
				mapFn: addPrefix,
				err: &testMultiUnwrapError{
					msg:  "multi wrapper with nils",
					errs: []error{baseErr1, nil, baseErr3},
				},
			},
			want: errors.Join(
				fmt.Errorf("mapped: %w", baseErr1),
				nil,
				fmt.Errorf("mapped: %w", baseErr3),
			),
		},
		{
			name: "standard_fmt_wrapped_error",
			args: args{
				mapFn: addPrefix,
				err:   fmt.Errorf("wrapper: %w", baseErr1),
			},
			want: fmt.Errorf("mapped: %w", baseErr1),
		},
		{
			name: "standard_errors_join",
			args: args{
				mapFn: addPrefix,
				err:   errors.Join(baseErr1, baseErr2),
			},
			want: errors.Join(
				fmt.Errorf("mapped: %w", baseErr1),
				fmt.Errorf("mapped: %w", baseErr2),
			),
		},
		{
			name: "complex_nested_structure",
			args: args{
				mapFn: addPrefix,
				err: &testMultiUnwrapError{
					msg: "outer multi",
					errs: []error{
						&testSingleUnwrapError{
							msg: "inner single",
							err: baseErr1,
						},
						&testMultiUnwrapError{
							msg:  "inner multi",
							errs: []error{baseErr2, baseErr3},
						},
					},
				},
			},
			want: errors.Join(
				fmt.Errorf("mapped: %w", baseErr1),
				errors.Join(
					fmt.Errorf("mapped: %w", baseErr2),
					fmt.Errorf("mapped: %w", baseErr3),
				),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := MapError(tt.args.mapFn, tt.args.err)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.want.Error(), got.Error())
		})
	}
}

func TestMapError_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mapFn   MapFunc
		err     error
		wantErr string
	}{
		{
			name: "map_function_panics",
			mapFn: func(err error) error {
				panic("mapping function panicked")
			},
			err: errors.New("test error"),
		},
		{
			name: "map_function_returns_nil_for_non_nil_input",
			mapFn: func(err error) error {
				return nil
			},
			err:     errors.New("test error"),
			wantErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.name == "map_function_panics" {
				assert.Panics(t, func() {
					_ = MapError(tt.mapFn, tt.err)
				})
				return
			}

			got := MapError(tt.mapFn, tt.err)
			if tt.wantErr == "" {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.wantErr, got.Error())
			}
		})
	}
}

func TestMapFunc(t *testing.T) {
	t.Parallel()

	// Test that MapFunc is a proper function type
	var mapFn MapFunc = func(err error) error {
		return fmt.Errorf("wrapped: %w", err)
	}

	testErr := errors.New("test error")
	result := mapFn(testErr)

	assert.Equal(t, "wrapped: test error", result.Error())
	assert.ErrorIs(t, result, testErr)
}

// Benchmark to ensure the function performs well with complex error structures.
func BenchmarkMapError(b *testing.B) {
	baseErr := errors.New("base error")
	complexErr := &testMultiUnwrapError{
		msg: "complex",
		errs: []error{
			&testSingleUnwrapError{msg: "single1", err: baseErr},
			&testSingleUnwrapError{msg: "single2", err: baseErr},
			&testMultiUnwrapError{
				msg:  "nested multi",
				errs: []error{baseErr, baseErr, baseErr},
			},
		},
	}

	mapFn := func(err error) error {
		return fmt.Errorf("mapped: %w", err)
	}

	b.ResetTimer()
	for range b.N {
		_ = MapError(mapFn, complexErr)
	}
}
