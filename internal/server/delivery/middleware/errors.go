package middleware

import (
	"net/http"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
)

// MiddlewareErrRegistry defines error handling policies for middleware operations.
var MiddlewareErrRegistry = errutil.Registry{
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
}

// handleError processes middleware errors using the registry and returns appropriate HTTP response.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(MiddlewareErrRegistry, err, c)
}
