package filedata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrIncorrectStorageKey",
			err:  ErrIncorrectStorageKey,
			want: "invalid storage key",
		},
		{
			name: "ErrIncorrectHashSum",
			err:  ErrIncorrectHashSum,
			want: "invalid hash sum",
		},
		{
			name: "ErrNewFileParamsValidation",
			err:  ErrNewFileParamsValidation,
			want: "new file parameters validation failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}
