package credential

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/credential"
	"github.com/google/uuid"
)

// Repository defines the interface for credential data persistence operations.
type Repository interface {
	// Save persists a credential entity using the provided parameters.
	Save(ctx context.Context, params repository.SaveParams) error

	// Load retrieves credential entities using the provided parameters.
	Load(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error)
}

// Service provides credential management business logic operations.
type Service struct {
	// r is the repository interface for credential data persistence operations.
	r Repository
}

// NewService creates a new credential service instance with the provided repository.
func NewService(r Repository) *Service {
	return &Service{r: r}
}

// Pull retrieves a specific credential for the given user.
func (s *Service) Pull(ctx context.Context, params PullParams) (*Credential, error) {
	creds, err := s.r.Load(ctx, repository.LoadParams{
		ID:     params.ID,
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", mapError(err))
	}
	if len(creds) == 0 {
		return nil, fmt.Errorf("credential not found: %w", ErrCredentialNotFound)
	}
	return newCredentialFromDomain(creds[0]), nil
}

// List retrieves all credentials for the specified user.
func (s *Service) List(ctx context.Context, params ListParams) ([]*Credential, error) {
	creds, err := s.r.Load(ctx, repository.LoadParams{
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", mapError(err))
	}
	return newCredentialsFromDomain(creds), nil
}

// Push creates or updates a credential for the specified user.
func (s *Service) Push(ctx context.Context, params *PushParams) (uuid.UUID, error) {
	cred, err := credential.NewCredential(credential.NewCredentialParams{
		UserID:      params.UserID,
		Login:       params.Login,
		Password:    params.Password,
		Description: params.Description,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create credential: %w", mapError(err))
	}

	if params.ID != uuid.Nil {
		if err := s.checkAccessToUpdate(ctx, params.ID, params.UserID); err != nil {
			return uuid.Nil, fmt.Errorf("access check for updating credential failed: %w", err)
		}
		cred.ID = params.ID
	}

	if err := s.r.Save(ctx, repository.SaveParams{Entity: cred}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save credential: %w", mapError(err))
	}
	return cred.ID, nil
}

// checkAccessToUpdate verifies that the user has permission to update the specified credential.
func (s *Service) checkAccessToUpdate(ctx context.Context, credID, userID uuid.UUID) error {
	exists, err := s.Pull(ctx, PullParams{ID: credID, UserID: userID})
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return fmt.Errorf("credential for update not found: %w", err)
		}
		return fmt.Errorf("failed to pull existing credential: %w", mapError(err))
	}
	if exists.UserID != userID {
		return fmt.Errorf("access denied to credential: %w", ErrCredentialAccessDenied)
	}
	return nil
}
