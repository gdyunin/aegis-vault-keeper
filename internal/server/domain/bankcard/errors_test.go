package bankcard

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
			name: "ErrInvalidCardNumber",
			err:  ErrInvalidCardNumber,
			want: "card number must contain 13–19 digits",
		},
		{
			name: "ErrLuhnCheckFailed",
			err:  ErrLuhnCheckFailed,
			want: "card number failed Luhn check",
		},
		{
			name: "ErrEmptyCardHolder",
			err:  ErrEmptyCardHolder,
			want: "card holder cannot be empty",
		},
		{
			name: "ErrInvalidExpiryMonth",
			err:  ErrInvalidExpiryMonth,
			want: "expiry month must be a valid 2-digit month (01–12)",
		},
		{
			name: "ErrInvalidExpiryYear",
			err:  ErrInvalidExpiryYear,
			want: "expiry year must be a valid 4-digit year",
		},
		{
			name: "ErrCardExpired",
			err:  ErrCardExpired,
			want: "card has expired",
		},
		{
			name: "ErrInvalidCVV",
			err:  ErrInvalidCVV,
			want: "CVV must contain 3 or 4 digits",
		},
		{
			name: "ErrNewBankCardParamsValidation",
			err:  ErrNewBankCardParamsValidation,
			want: "new bank card parameters validation failed",
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
