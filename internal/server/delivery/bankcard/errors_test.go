package bankcard

import (
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBankCardErrRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		errorIn        error
		expectedPolicy errutil.Policy
		found          bool
	}{
		{
			name:    "tech error",
			errorIn: bankcard.ErrBankCardTechError,
			expectedPolicy: errutil.Policy{
				StatusCode: 500,
				PublicMsg:  "Internal Server Error",
				LogIt:      true,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassTech,
			},
			found: true,
		},
		{
			name:    "not found error",
			errorIn: bankcard.ErrBankCardNotFound,
			expectedPolicy: errutil.Policy{
				StatusCode: 404,
				PublicMsg:  "Bank card not found",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassGeneric,
			},
			found: true,
		},
		{
			name:    "access denied error",
			errorIn: bankcard.ErrBankCardAccessDenied,
			expectedPolicy: errutil.Policy{
				StatusCode: 403,
				PublicMsg:  "Access to this bank card is denied",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassAuth,
			},
			found: true,
		},
		{
			name:    "invalid card number error",
			errorIn: bankcard.ErrBankCardInvalidCardNumber,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "Card number must contain 13–19 digits",
				LogIt:      false,
				AllowMerge: true,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "empty card holder error",
			errorIn: bankcard.ErrBankCardEmptyCardHolder,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "Card holder must not be empty",
				LogIt:      false,
				AllowMerge: true,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "card expired error",
			errorIn: bankcard.ErrBankCardCardExpired,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "Card has expired",
				LogIt:      false,
				AllowMerge: true,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "app error",
			errorIn: bankcard.ErrBankCardAppError,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "Invalid parameters",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "unknown error",
			errorIn: errors.New("unknown error"),
			found:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Find the error in registry
			var found bool
			var actualPolicy errutil.Policy

			for _, entry := range BankCardErrRegistry {
				if errors.Is(tt.errorIn, entry.ErrorIn) {
					found = true
					actualPolicy = entry.HandlePolicy
					break
				}
			}

			assert.Equal(t, tt.found, found, "Error should be found in registry: %v", tt.found)

			if tt.found {
				assert.Equal(t, tt.expectedPolicy.StatusCode, actualPolicy.StatusCode)
				assert.Equal(t, tt.expectedPolicy.PublicMsg, actualPolicy.PublicMsg)
				assert.Equal(t, tt.expectedPolicy.LogIt, actualPolicy.LogIt)
				assert.Equal(t, tt.expectedPolicy.AllowMerge, actualPolicy.AllowMerge)
				assert.Equal(t, tt.expectedPolicy.ErrorClass, actualPolicy.ErrorClass)
			}
		})
	}
}

func TestBankCardErrRegistry_AllErrorsCovered(t *testing.T) {
	t.Parallel()

	// List of all bankcard errors that should be in the registry
	expectedErrors := []error{
		bankcard.ErrBankCardTechError,
		bankcard.ErrBankCardNotFound,
		bankcard.ErrBankCardAccessDenied,
		bankcard.ErrBankCardInvalidCardNumber,
		bankcard.ErrBankCardLuhnCheckFailed,
		bankcard.ErrBankCardEmptyCardHolder,
		bankcard.ErrBankCardInvalidExpiryMonth,
		bankcard.ErrBankCardInvalidExpiryYear,
		bankcard.ErrBankCardCardExpired,
		bankcard.ErrBankCardInvalidCVV,
		bankcard.ErrBankCardAppError,
	}

	// Check that all expected errors are in the registry
	for _, expectedErr := range expectedErrors {
		found := false
		for _, entry := range BankCardErrRegistry {
			if errors.Is(expectedErr, entry.ErrorIn) {
				found = true
				break
			}
		}
		assert.True(t, found, "Error %v should be in the registry", expectedErr)
	}

	// Verify registry has expected number of entries
	assert.Len(t, BankCardErrRegistry, len(expectedErrors),
		"Registry should contain exactly %d entries", len(expectedErrors))
}

func TestBankCardErrRegistry_StatusCodes(t *testing.T) {
	t.Parallel()

	statusCodeTests := []struct {
		error      error
		statusCode int
	}{
		{bankcard.ErrBankCardTechError, 500},
		{bankcard.ErrBankCardNotFound, 404},
		{bankcard.ErrBankCardAccessDenied, 403},
		{bankcard.ErrBankCardInvalidCardNumber, 400},
		{bankcard.ErrBankCardLuhnCheckFailed, 400},
		{bankcard.ErrBankCardEmptyCardHolder, 400},
		{bankcard.ErrBankCardInvalidExpiryMonth, 400},
		{bankcard.ErrBankCardInvalidExpiryYear, 400},
		{bankcard.ErrBankCardCardExpired, 400},
		{bankcard.ErrBankCardInvalidCVV, 400},
		{bankcard.ErrBankCardAppError, 400},
	}

	for _, tt := range statusCodeTests {
		t.Run(tt.error.Error(), func(t *testing.T) {
			t.Parallel()

			for _, entry := range BankCardErrRegistry {
				if errors.Is(tt.error, entry.ErrorIn) {
					assert.Equal(t, tt.statusCode, entry.HandlePolicy.StatusCode)
					return
				}
			}
			t.Errorf("Error %v not found in registry", tt.error)
		})
	}
}

func TestBankCardErrRegistry_ErrorClasses(t *testing.T) {
	t.Parallel()

	errorClassTests := []struct {
		error      error
		errorClass errutil.ErrorClass
	}{
		{bankcard.ErrBankCardTechError, errutil.ErrorClassTech},
		{bankcard.ErrBankCardNotFound, errutil.ErrorClassGeneric},
		{bankcard.ErrBankCardAccessDenied, errutil.ErrorClassAuth},
		{bankcard.ErrBankCardInvalidCardNumber, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardLuhnCheckFailed, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardEmptyCardHolder, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardInvalidExpiryMonth, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardInvalidExpiryYear, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardCardExpired, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardInvalidCVV, errutil.ErrorClassValidation},
		{bankcard.ErrBankCardAppError, errutil.ErrorClassValidation},
	}

	for _, tt := range errorClassTests {
		t.Run(tt.error.Error(), func(t *testing.T) {
			t.Parallel()

			for _, entry := range BankCardErrRegistry {
				if errors.Is(tt.error, entry.ErrorIn) {
					assert.Equal(t, tt.errorClass, entry.HandlePolicy.ErrorClass)
					return
				}
			}
			t.Errorf("Error %v not found in registry", tt.error)
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputError         error
		name               string
		expectedMessages   []string
		expectedStatusCode int
	}{
		{
			name:               "tech error",
			inputError:         bankcard.ErrBankCardTechError,
			expectedStatusCode: 500,
			expectedMessages:   []string{"Internal Server Error"},
		},
		{
			name:               "not found error",
			inputError:         bankcard.ErrBankCardNotFound,
			expectedStatusCode: 404,
			expectedMessages:   []string{"Bank card not found"},
		},
		{
			name:               "access denied error",
			inputError:         bankcard.ErrBankCardAccessDenied,
			expectedStatusCode: 403,
			expectedMessages:   []string{"Access to this bank card is denied"},
		},
		{
			name:               "invalid card number error",
			inputError:         bankcard.ErrBankCardInvalidCardNumber,
			expectedStatusCode: 400,
			expectedMessages:   []string{"Card number must contain 13–19 digits"},
		},
		{
			name:               "empty card holder error",
			inputError:         bankcard.ErrBankCardEmptyCardHolder,
			expectedStatusCode: 400,
			expectedMessages:   []string{"Card holder must not be empty"},
		},
		{
			name:               "card expired error",
			inputError:         bankcard.ErrBankCardCardExpired,
			expectedStatusCode: 400,
			expectedMessages:   []string{"Card has expired"},
		},
		{
			name:               "app error",
			inputError:         bankcard.ErrBankCardAppError,
			expectedStatusCode: 400,
			expectedMessages:   []string{"Invalid parameters"},
		},
		{
			name:               "unknown error",
			inputError:         errors.New("unknown error"),
			expectedStatusCode: 500,
			expectedMessages:   []string{"Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup gin context
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(nil)

			// Execute
			statusCode, messages := handleError(tt.inputError, c)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, statusCode)
			assert.Equal(t, tt.expectedMessages, messages)
		})
	}
}

func TestHandleError_WithWrappedErrors(t *testing.T) {
	t.Parallel()

	// Test with wrapped bankcard errors
	wrappedNotFoundError := errors.New("database query failed: " + bankcard.ErrBankCardNotFound.Error())

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	statusCode, messages := handleError(wrappedNotFoundError, c)

	// Should still handle the error as unknown since it's not directly wrapped with errors.Wrap
	assert.Equal(t, 500, statusCode)
	assert.Equal(t, []string{"Internal Server Error"}, messages)
}
