package note

import (
	"net/http"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
)

// NoteErrRegistry defines error handling policies for note operations.
var NoteErrRegistry = errutil.Registry{

	{
		ErrorIn: app.ErrNoteTechError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusInternalServerError,
			PublicMsg:  http.StatusText(http.StatusInternalServerError),
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassTech,
		},
	},

	{
		ErrorIn: app.ErrNoteAccessDenied,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusForbidden,
			PublicMsg:  "Access to this note is denied",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassAuth,
		},
	},

	{
		ErrorIn: app.ErrNoteNotFound,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusNotFound,
			PublicMsg:  "Note not found",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassGeneric,
		},
	},

	{
		ErrorIn: app.ErrNoteIncorrectNoteText,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid note text",
			LogIt:      false,
			AllowMerge: true,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},

	{
		ErrorIn: app.ErrNoteAppError,
		HandlePolicy: errutil.Policy{
			StatusCode: http.StatusBadRequest,
			PublicMsg:  "Invalid parameters",
			LogIt:      false,
			AllowMerge: false,
			ErrorClass: errutil.ErrorClassValidation,
		},
	},
}

// handleError processes note errors using the registry and returns appropriate HTTP response.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(NoteErrRegistry, err, c)
}
