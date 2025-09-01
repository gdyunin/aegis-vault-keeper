package bankcard

import (
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/errutil"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
)

// Bank card application error definitions.
var (
	// ErrBankCardAppError indicates a general bank card application error.
	ErrBankCardAppError = errors.New("bank card application error")

	// ErrBankCardTechError indicates a technical error in the bank card system.
	ErrBankCardTechError = errors.New("bank card technical error")

	// ErrBankCardInvalidCardNumber indicates the card number format is invalid.
	ErrBankCardInvalidCardNumber = errors.New("card number must contain 13–19 digits")

	// ErrBankCardLuhnCheckFailed indicates the card number failed Luhn algorithm validation.
	ErrBankCardLuhnCheckFailed = errors.New("card number failed Luhn check")

	// ErrBankCardEmptyCardHolder indicates the card holder name is empty.
	ErrBankCardEmptyCardHolder = errors.New("card holder cannot be empty")

	// ErrBankCardInvalidExpiryMonth indicates the expiry month format is invalid.
	ErrBankCardInvalidExpiryMonth = errors.New("expiry month must be a valid 2-digit month (01–12)")

	// ErrBankCardInvalidExpiryYear indicates the expiry year format is invalid.
	ErrBankCardInvalidExpiryYear = errors.New("expiry year must be a valid 4-digit year")

	// ErrBankCardCardExpired indicates the card has passed its expiry date.
	ErrBankCardCardExpired = errors.New("card has expired")

	// ErrBankCardInvalidCVV indicates the CVV format is invalid.
	ErrBankCardInvalidCVV = errors.New("CVV must contain 3 or 4 digits")

	// ErrBankCardNotFound indicates the requested bank card was not found.
	ErrBankCardNotFound = errors.New("bank card not found")

	// ErrBankCardAccessDenied indicates access to the bank card is not permitted.
	ErrBankCardAccessDenied = errors.New("access to this bank card is denied")
)

// mapError maps domain errors to application-level errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	mapped := errutil.MapError(mapFn, err)
	if mapped != nil {
		return fmt.Errorf("error after mapping: %w", mapped)
	}
	return nil
}

// mapFn provides the actual error mapping logic for different bank card error types.
func mapFn(err error) error {
	switch {
	case errors.Is(err, bankcard.ErrNewBankCardParamsValidation):
		return ErrBankCardAppError
	case errors.Is(err, bankcard.ErrInvalidCardNumber):
		return ErrBankCardInvalidCardNumber
	case errors.Is(err, bankcard.ErrLuhnCheckFailed):
		return ErrBankCardLuhnCheckFailed
	case errors.Is(err, bankcard.ErrEmptyCardHolder):
		return ErrBankCardEmptyCardHolder
	case errors.Is(err, bankcard.ErrInvalidExpiryMonth):
		return ErrBankCardInvalidExpiryMonth
	case errors.Is(err, bankcard.ErrInvalidExpiryYear):
		return ErrBankCardInvalidExpiryYear
	case errors.Is(err, bankcard.ErrCardExpired):
		return ErrBankCardCardExpired
	case errors.Is(err, bankcard.ErrInvalidCVV):
		return ErrBankCardInvalidCVV
	default:
		return errors.Join(ErrBankCardTechError, err)
	}
}
