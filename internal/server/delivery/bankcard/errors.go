package bankcard

import (
	"net/http"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
)

// BankCardErrRegistry maps bank card application errors to HTTP responses.
// Each rule defines status codes, public messages, logging behavior, and error classification.
var BankCardErrRegistry = errutil.Registry{

	{
		ErrorIn: app.ErrBankCardTechError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusInternalServerError,
			PublicMsg:  http.StatusText(http.StatusInternalServerError),
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassTech,
		},
	},

	{
		ErrorIn: app.ErrBankCardAccessDenied,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusForbidden,
			PublicMsg:  "Access to this bank card is denied",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassAuth,
		},
	},

	{
		ErrorIn: app.ErrBankCardNotFound,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusNotFound,
			PublicMsg:  "Bank card not found",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassGeneric,
		},
	},
	{
		ErrorIn: app.ErrBankCardInvalidCardNumber,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Card number must contain 13–19 digits",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrBankCardLuhnCheckFailed,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Card number failed Luhn check",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrBankCardEmptyCardHolder,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Card holder must not be empty",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrBankCardInvalidExpiryMonth,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Expiry month must be a valid 2-digit month (01–12)",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrBankCardInvalidExpiryYear,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Expiry year must be a valid 4-digit year",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrBankCardCardExpired,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Card has expired",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrBankCardInvalidCVV,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "CVV must contain 3 or 4 digits",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},

	{
		ErrorIn: app.ErrBankCardAppError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid parameters",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
}

// handleError processes bank card application errors using the registry.
// Returns HTTP status code and error messages for response.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(BankCardErrRegistry, err, c)
}
