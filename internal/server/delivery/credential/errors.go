package credential

import (
	"net/http"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
)

// CredentialErrRegistry maps credential application errors to HTTP responses.
// Each rule defines status codes, public messages, logging behavior, and error classification.
var CredentialErrRegistry = errutil.Registry{

	{
		ErrorIn: app.ErrCredentialTechError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusInternalServerError,
			PublicMsg:  http.StatusText(http.StatusInternalServerError),
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassTech,
		},
	},

	{
		ErrorIn: app.ErrCredentialAccessDenied,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusForbidden,
			PublicMsg:  "Access to this credential is denied",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassAuth,
		},
	},

	{
		ErrorIn: app.ErrCredentialNotFound,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusNotFound,
			PublicMsg:  "Credential not found",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassGeneric,
		},
	},

	{
		ErrorIn: app.ErrCredentialIncorrectLogin,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid login",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrCredentialIncorrectPassword,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid password",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},

	{
		ErrorIn: app.ErrCredentialAppError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid parameters",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
}

// handleError processes credential application errors using the registry.
// Returns HTTP status code and error messages for response.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(CredentialErrRegistry, err, c)
}
