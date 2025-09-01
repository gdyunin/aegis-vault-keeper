package note

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Note represents a text note with optional description.
type Note struct {
	// UpdatedAt contains the timestamp when the note was last modified.
	UpdatedAt time.Time
	// Note contains the encrypted note content.
	Note []byte
	// Description contains the encrypted note description.
	Description []byte
	// ID uniquely identifies this note.
	ID uuid.UUID
	// UserID identifies the user who owns this note.
	UserID uuid.UUID
}

// NewNote creates a new note with the provided parameters after validation.
func NewNote(params NewNoteParams) (*Note, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.Join(ErrNewNoteParamsValidation, err)
	}

	n := Note{
		ID:          uuid.New(),
		UserID:      params.UserID,
		Note:        []byte(params.Note),
		Description: []byte(params.Description),
		UpdatedAt:   time.Now(),
	}
	return &n, nil
}

// NewNoteParams contains parameters for creating a new note.
type NewNoteParams struct {
	// Note contains the text content of the note (required).
	Note string
	// Description contains an optional description for the note.
	Description string
	// UserID identifies the user who will own this note.
	UserID uuid.UUID
}

// Validate checks that the note creation parameters are valid.
func (np *NewNoteParams) Validate() error {
	validations := []func() error{
		np.validateNote,
	}

	// errs collects all validation errors encountered during note validation.
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

// validateNote ensures that the note text is not empty.
func (np *NewNoteParams) validateNote() error {
	if np.Note == "" {
		return ErrIncorrectNoteText
	}
	return nil
}
