package bankcard

import (
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputErr error
		wantErr  error
		name     string
	}{
		{
			name:     "nil_error",
			inputErr: nil,
			wantErr:  nil,
		},
		{
			name:     "domain_validation_error",
			inputErr: bankcard.ErrNewBankCardParamsValidation,
			wantErr:  ErrBankCardAppError,
		},
		{
			name:     "domain_invalid_card_number",
			inputErr: bankcard.ErrInvalidCardNumber,
			wantErr:  ErrBankCardInvalidCardNumber,
		},
		{
			name:     "domain_luhn_check_failed",
			inputErr: bankcard.ErrLuhnCheckFailed,
			wantErr:  ErrBankCardLuhnCheckFailed,
		},
		{
			name:     "domain_empty_card_holder",
			inputErr: bankcard.ErrEmptyCardHolder,
			wantErr:  ErrBankCardEmptyCardHolder,
		},
		{
			name:     "domain_invalid_expiry_month",
			inputErr: bankcard.ErrInvalidExpiryMonth,
			wantErr:  ErrBankCardInvalidExpiryMonth,
		},
		{
			name:     "domain_invalid_expiry_year",
			inputErr: bankcard.ErrInvalidExpiryYear,
			wantErr:  ErrBankCardInvalidExpiryYear,
		},
		{
			name:     "domain_card_expired",
			inputErr: bankcard.ErrCardExpired,
			wantErr:  ErrBankCardCardExpired,
		},
		{
			name:     "domain_invalid_cvv",
			inputErr: bankcard.ErrInvalidCVV,
			wantErr:  ErrBankCardInvalidCVV,
		},
		{
			name:     "unknown_error",
			inputErr: errors.New("unknown error"),
			wantErr:  ErrBankCardTechError, // mapError wraps this with "error after mapping"
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapError(tt.inputErr)

			if tt.wantErr == nil {
				assert.Nil(t, result)
			} else {
				require.Error(t, result)
				assert.Contains(t, result.Error(), "error after mapping")
				assert.ErrorIs(t, result, tt.wantErr)
			}
		})
	}
}

func TestMapFn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputErr error
		wantErr  error
		name     string
	}{
		{
			name:     "domain_validation_error",
			inputErr: bankcard.ErrNewBankCardParamsValidation,
			wantErr:  ErrBankCardAppError,
		},
		{
			name:     "domain_invalid_card_number",
			inputErr: bankcard.ErrInvalidCardNumber,
			wantErr:  ErrBankCardInvalidCardNumber,
		},
		{
			name:     "domain_luhn_check_failed",
			inputErr: bankcard.ErrLuhnCheckFailed,
			wantErr:  ErrBankCardLuhnCheckFailed,
		},
		{
			name:     "domain_empty_card_holder",
			inputErr: bankcard.ErrEmptyCardHolder,
			wantErr:  ErrBankCardEmptyCardHolder,
		},
		{
			name:     "domain_invalid_expiry_month",
			inputErr: bankcard.ErrInvalidExpiryMonth,
			wantErr:  ErrBankCardInvalidExpiryMonth,
		},
		{
			name:     "domain_invalid_expiry_year",
			inputErr: bankcard.ErrInvalidExpiryYear,
			wantErr:  ErrBankCardInvalidExpiryYear,
		},
		{
			name:     "domain_card_expired",
			inputErr: bankcard.ErrCardExpired,
			wantErr:  ErrBankCardCardExpired,
		},
		{
			name:     "domain_invalid_cvv",
			inputErr: bankcard.ErrInvalidCVV,
			wantErr:  ErrBankCardInvalidCVV,
		},
		{
			name:     "wrapped_domain_error",
			inputErr: errors.Join(errors.New("wrapper"), bankcard.ErrInvalidCardNumber),
			wantErr:  ErrBankCardInvalidCardNumber,
		},
		{
			name:     "unknown_error",
			inputErr: errors.New("unknown error"),
			wantErr:  ErrBankCardTechError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapFn(tt.inputErr)

			require.NotNil(t, result)
			assert.ErrorIs(t, result, tt.wantErr)

			// For unknown errors, check that it's joined with ErrBankCardTechError
			if tt.name == "unknown_error" {
				assert.ErrorIs(t, result, tt.inputErr)
			}
		})
	}
}
