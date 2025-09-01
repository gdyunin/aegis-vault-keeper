package bankcard

import "errors"

// ErrInvalidCardNumber indicates the card number format is invalid (must be 13-19 digits).
var ErrInvalidCardNumber = errors.New("card number must contain 13–19 digits")

// ErrLuhnCheckFailed indicates the card number failed the Luhn checksum validation.
var ErrLuhnCheckFailed = errors.New("card number failed Luhn check")

// ErrEmptyCardHolder indicates the cardholder name is required but not provided.
var ErrEmptyCardHolder = errors.New("card holder cannot be empty")

// ErrInvalidExpiryMonth indicates the expiry month format is invalid.
var ErrInvalidExpiryMonth = errors.New("expiry month must be a valid 2-digit month (01–12)")

// ErrInvalidExpiryYear indicates the expiry year format is invalid.
var ErrInvalidExpiryYear = errors.New("expiry year must be a valid 4-digit year")

// ErrCardExpired indicates the card has already expired.
var ErrCardExpired = errors.New("card has expired")

// ErrInvalidCVV indicates the CVV format is invalid.
var ErrInvalidCVV = errors.New("CVV must contain 3 or 4 digits")

// ErrNewBankCardParamsValidation indicates validation failure during bank card creation.
var ErrNewBankCardParamsValidation = errors.New("new bank card parameters validation failed")
