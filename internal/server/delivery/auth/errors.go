package auth

import (
	"net/http"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
)

// AuthErrRegistry defines error handling policies for authentication-related errors.
// Each entry maps application errors to HTTP status codes and public messages.
var AuthErrRegistry = errutil.Registry{
	{
		ErrorIn: app.ErrAuthTechError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusInternalServerError,
			PublicMsg:  http.StatusText(http.StatusInternalServerError),
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassTech,
		},
	},
	{
		ErrorIn: app.ErrAuthWrongLoginOrPassword,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusUnauthorized,
			PublicMsg:  "The provided login or password is incorrect",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassAuth,
		},
	},
	{
		ErrorIn: app.ErrAuthInvalidAccessToken,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusUnauthorized,
			PublicMsg:  "Your access token is invalid or has expired. Please log in",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassAuth,
		},
	},
	{
		ErrorIn: app.ErrAuthIncorrectLogin,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "The login provided is not valid",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrAuthIncorrectPassword,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "The password provided is not valid",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrAuthUserAlreadyExists,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusConflict,
			PublicMsg:  "User with this login already exists",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
	{
		ErrorIn: app.ErrAuthAppError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "The parameters provided are invalid",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
}

// handleError processes authentication errors using the registry and returns appropriate status code and messages.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(AuthErrRegistry, err, c)
}
