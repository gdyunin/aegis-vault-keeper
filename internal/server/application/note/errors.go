package note

import (
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/errutil"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
)

// Note error definitions.
var (
	// ErrNoteAppError indicates a general note application error.
	ErrNoteAppError = errors.New("note application error")

	// ErrNoteTechError indicates a technical error in the note system.
	ErrNoteTechError = errors.New("note technical error")

	// ErrNoteIncorrectNoteText indicates incorrect note text was provided.
	ErrNoteIncorrectNoteText = errors.New("incorrect note text")

	// ErrNoteNotFound indicates the requested note was not found.
	ErrNoteNotFound = errors.New("note not found")

	// ErrNoteAccessDenied indicates access to the note is not permitted.
	ErrNoteAccessDenied = errors.New("access to this note is denied")
)

// mapError maps domain and repository errors to application-level errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	mapped := errutil.MapError(mapFn, err)
	if mapped != nil {
		return fmt.Errorf("note error mapping failed: %w", mapped)
	}
	return nil
}

// mapFn provides the actual error mapping logic for different error types.
func mapFn(err error) error {
	switch {
	case errors.Is(err, note.ErrNewNoteParamsValidation):
		return ErrNoteAppError
	case errors.Is(err, note.ErrIncorrectNoteText):
		return ErrNoteIncorrectNoteText
	default:
		return errors.Join(ErrNoteTechError, err)
	}
}
