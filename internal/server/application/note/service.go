package note

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/note"
	"github.com/google/uuid"
)

// Repository defines the interface for note data persistence operations.
type Repository interface {
	// Save persists a note entity using the provided parameters.
	Save(ctx context.Context, params repository.SaveParams) error

	// Load retrieves note entities using the provided parameters.
	Load(ctx context.Context, params repository.LoadParams) ([]*note.Note, error)
}

// Service provides note management business logic operations.
type Service struct {
	// r is the repository interface for note data persistence operations.
	r Repository
}

// NewService creates a new note service instance with the provided repository.
func NewService(r Repository) *Service {
	return &Service{r: r}
}

// Pull retrieves a specific note for the given user.
func (s *Service) Pull(ctx context.Context, params PullParams) (*Note, error) {
	notes, err := s.r.Load(ctx, repository.LoadParams{
		ID:     params.ID,
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", mapError(err))
	}
	if len(notes) == 0 {
		return nil, fmt.Errorf("note not found: %w", ErrNoteNotFound)
	}
	return newNoteFromDomain(notes[0]), nil
}

// List retrieves all notes for the specified user.
func (s *Service) List(ctx context.Context, params ListParams) ([]*Note, error) {
	notes, err := s.r.Load(ctx, repository.LoadParams{
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", mapError(err))
	}
	return newNotesFromDomain(notes), nil
}

// Push creates or updates a note for the specified user.
func (s *Service) Push(ctx context.Context, params *PushParams) (uuid.UUID, error) {
	n, err := note.NewNote(note.NewNoteParams{
		UserID:      params.UserID,
		Note:        params.Note,
		Description: params.Description,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create new note: %w", mapError(err))
	}

	if params.ID != uuid.Nil {
		if err := s.checkAccessToUpdate(ctx, params.ID, params.UserID); err != nil {
			return uuid.Nil, fmt.Errorf("access check for updating note failed: %w", err)
		}
		n.ID = params.ID
	}

	if err := s.r.Save(ctx, repository.SaveParams{Entity: n}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save note: %w", mapError(err))
	}
	return n.ID, nil
}

// checkAccessToUpdate verifies that the user has permission to update the specified note.
func (s *Service) checkAccessToUpdate(ctx context.Context, noteID, userID uuid.UUID) error {
	exists, err := s.Pull(ctx, PullParams{ID: noteID, UserID: userID})
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return fmt.Errorf("note for update not found: %w", err)
		}
		return fmt.Errorf("failed to pull existing note: %w", mapError(err))
	}
	if exists.UserID != userID {
		return fmt.Errorf("access denied to note: %w", ErrNoteAccessDenied)
	}
	return nil
}
