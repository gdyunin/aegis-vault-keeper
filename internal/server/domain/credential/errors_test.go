package credential

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
			name: "ErrNewCredentialParamsValidation",
			err:  ErrNewCredentialParamsValidation,
			want: "invalid parameters for new credential",
		},
		{
			name: "ErrIncorrectLogin",
			err:  ErrIncorrectLogin,
			want: "incorrect login",
		},
		{
			name: "ErrIncorrectPassword",
			err:  ErrIncorrectPassword,
			want: "incorrect password",
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
