package filedata

import (
	"net/http"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
)

// FileDataErrRegistry defines error handling policies for file data operations.
var FileDataErrRegistry = errutil.Registry{

	{
		ErrorIn: app.ErrFileTechError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusInternalServerError,
			PublicMsg:  http.StatusText(http.StatusInternalServerError),
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassTech,
		},
	},

	{
		ErrorIn: app.ErrFileAccessDenied,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusForbidden,
			PublicMsg:  "Access to this file is denied",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassAuth,
		},
	},

	{
		ErrorIn: app.ErrFileNotFound,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusNotFound,
			PublicMsg:  "File not found",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassGeneric,
		},
	},

	{
		ErrorIn: app.ErrFileIncorrectStorageKey,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid storage key",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},

	{
		ErrorIn: app.ErrFileIncorrectHashSum,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid hash sum",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},

	{
		ErrorIn: app.ErrFileDataRequired,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "File data is required",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},

	{
		ErrorIn: app.ErrRollBackFileSaveFailed,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusInternalServerError,
			PublicMsg:  http.StatusText(http.StatusInternalServerError),
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassTech,
		},
	},

	{
		ErrorIn: app.ErrFileAppError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid parameters",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
}

// handleError processes file data errors using the registry and returns appropriate HTTP response.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(FileDataErrRegistry, err, c)
}
