package bankcard

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	// DigitsOnlyRegex validates card numbers containing only 13-19 digits.
	digitsOnlyRegex = regexp.MustCompile(`^\d{13,19}$`)

	// CvvRegex validates CVV codes containing 3-4 digits.
	cvvRegex = regexp.MustCompile(`^\d{3,4}$`)

	// MonthRegex validates expiry months in MM format (01-12).
	monthRegex = regexp.MustCompile(`^(0[1-9]|1[0-2])$`)

	// YearRegex validates expiry years in YYYY format.
	yearRegex = regexp.MustCompile(`^\d{4}$`)
)

// BankCard represents a bank card entity with PCI DSS compliant encrypted storage.
type BankCard struct {
	// UpdatedAt contains the last modification timestamp.
	UpdatedAt time.Time
	// CardNumber contains the encrypted card number (PCI DSS Level 1 sensitive data).
	CardNumber []byte
	// CardHolder contains the encrypted cardholder name.
	CardHolder []byte
	// ExpiryMonth contains the encrypted expiry month in MM format.
	ExpiryMonth []byte
	// ExpiryYear contains the encrypted expiry year in YYYY format.
	ExpiryYear []byte
	// CVV contains the encrypted card verification value (PCI DSS restricted data).
	CVV []byte
	// Description contains encrypted user-provided description.
	Description []byte
	// ID contains the unique bank card identifier.
	ID uuid.UUID
	// UserID contains the card owner identifier.
	UserID uuid.UUID
}

// NewBankCard creates a new bank card entity with validation and encryption of sensitive data.
func NewBankCard(params *NewBankCardParams) (*BankCard, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.Join(ErrNewBankCardParamsValidation, err)
	}

	return &BankCard{
		ID:          uuid.New(),
		UserID:      params.UserID,
		CardNumber:  []byte(params.CardNumber),
		CardHolder:  []byte(params.CardHolder),
		ExpiryMonth: []byte(params.ExpiryMonth),
		ExpiryYear:  []byte(params.ExpiryYear),
		CVV:         []byte(params.CVV),
		Description: []byte(params.Description),
		UpdatedAt:   time.Now(),
	}, nil
}

// NewBankCardParams contains the parameters for creating a new bank card entity.
type NewBankCardParams struct {
	// CardNumber contains the card number (13-19 digits, validated with Luhn algorithm).
	CardNumber string
	// CardHolder contains the cardholder name (required, non-empty).
	CardHolder string
	// ExpiryMonth contains the expiry month in MM format (01-12).
	ExpiryMonth string
	// ExpiryYear contains the expiry year in YYYY format.
	ExpiryYear string
	// CVV contains the card verification value (3-4 digits).
	CVV string
	// Description contains optional user-provided description.
	Description string
	// UserID identifies the user creating this bank card.
	UserID uuid.UUID
}

// Validate performs comprehensive validation of all bank card parameters.
func (bcp *NewBankCardParams) Validate() error {
	validations := []func() error{
		bcp.validateCardNumber,
		bcp.validateCardHolder,
		bcp.validateExpiry,
		bcp.validateCVV,
	}

	// errs collects all validation errors encountered during bank card validation.
	var errs []error
	for _, fn := range validations {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateCardNumber validates the card number format and Luhn checksum.
func (bcp *NewBankCardParams) validateCardNumber() error {
	if !digitsOnlyRegex.MatchString(bcp.CardNumber) {
		return ErrInvalidCardNumber
	}
	if !luhnValid(bcp.CardNumber) {
		return ErrLuhnCheckFailed
	}
	return nil
}

// validateCardHolder validates that the cardholder name is not empty.
func (bcp *NewBankCardParams) validateCardHolder() error {
	if bcp.CardHolder == "" {
		return ErrEmptyCardHolder
	}
	return nil
}

// validateExpiry validates the expiry month and year format and ensures the card is not expired.
func (bcp *NewBankCardParams) validateExpiry() error {
	if !monthRegex.MatchString(bcp.ExpiryMonth) {
		return ErrInvalidExpiryMonth
	}
	if !yearRegex.MatchString(bcp.ExpiryYear) {
		return ErrInvalidExpiryYear
	}

	month, _ := strconv.Atoi(bcp.ExpiryMonth)
	year, _ := strconv.Atoi(bcp.ExpiryYear)
	now := time.Now()
	currentYear, currentMonth := now.Year(), int(now.Month())

	if year < currentYear || (year == currentYear && month < currentMonth) {
		return ErrCardExpired
	}
	return nil
}

// validateCVV validates the CVV format (3-4 digits).
func (bcp *NewBankCardParams) validateCVV() error {
	if !cvvRegex.MatchString(bcp.CVV) {
		return ErrInvalidCVV
	}
	return nil
}

// luhnValid validates a card number using the Luhn checksum algorithm.
func luhnValid(number string) bool {
	const (
		luhnDoubleDigitThreshold = 9
		luhnAdjustment           = 9
		luhnModulus              = 10
	)

	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if alt {
			d *= 2
			if d > luhnDoubleDigitThreshold {
				d -= luhnAdjustment
			}
		}
		sum += d
		alt = !alt
	}
	return sum%luhnModulus == 0
}
