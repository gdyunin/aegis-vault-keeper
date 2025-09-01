package note

import "errors"

// ErrNewNoteParamsValidation indicates that note creation parameters failed validation.
var ErrNewNoteParamsValidation = errors.New("new note parameters validation failed")

// ErrIncorrectNoteText indicates that the provided note text is invalid or empty.
var ErrIncorrectNoteText = errors.New("incorrect note text")
