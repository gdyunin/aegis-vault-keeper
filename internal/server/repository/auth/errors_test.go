package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrUserNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedText string
	}{
		{
			name:         "error message is correct",
			expectedText: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ErrUserNotFound

			assert.Error(t, err)
			assert.Equal(t, tt.expectedText, err.Error())
		})
	}
}

func TestErrUserAlreadyExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedText string
	}{
		{
			name:         "error message is correct",
			expectedText: "user already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ErrUserAlreadyExists

			assert.Error(t, err)
			assert.Equal(t, tt.expectedText, err.Error())
		})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		checkType   func(*testing.T, error)
		description string
	}{
		{
			name: "ErrUserNotFound is distinct error",
			err:  ErrUserNotFound,
			checkType: func(t *testing.T, err error) {
				t.Helper()
				assert.NotEqual(t, ErrUserAlreadyExists, err)
				assert.Contains(t, err.Error(), "not found")
			},
			description: "should be distinct from other errors",
		},
		{
			name: "ErrUserAlreadyExists is distinct error",
			err:  ErrUserAlreadyExists,
			checkType: func(t *testing.T, err error) {
				t.Helper()
				assert.NotEqual(t, ErrUserNotFound, err)
				assert.Contains(t, err.Error(), "already exists")
			},
			description: "should be distinct from other errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.checkType(t, tt.err)
		})
	}
}

func TestErrorComparisons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err1     error
		err2     error
		name     string
		shouldEq bool
	}{
		{
			name:     "same error instances are equal",
			err1:     ErrUserNotFound,
			err2:     ErrUserNotFound,
			shouldEq: true,
		},
		{
			name:     "different error instances are not equal",
			err1:     ErrUserNotFound,
			err2:     ErrUserAlreadyExists,
			shouldEq: false,
		},
		{
			name:     "same error instances are equal - already exists",
			err1:     ErrUserAlreadyExists,
			err2:     ErrUserAlreadyExists,
			shouldEq: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.shouldEq {
				assert.Equal(t, tt.err1, tt.err2)
			} else {
				assert.NotEqual(t, tt.err1, tt.err2)
			}
		})
	}
}
