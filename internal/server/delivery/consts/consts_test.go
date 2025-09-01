package consts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "HeaderXRequestID",
			got:  HeaderXRequestID,
			want: "X-Request-Id",
		},
		{
			name: "CtxKeyUserID",
			got:  CtxKeyUserID,
			want: "userID",
		},
		{
			name: "ErrorMessageInvalidParameters",
			got:  ErrorMessageInvalidParameters,
			want: "Invalid or missing request parameters",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.got)
		})
	}
}
