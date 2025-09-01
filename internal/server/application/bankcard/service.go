package bankcard

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/bankcard"
	"github.com/google/uuid"
)

// Repository defines the interface for bank card data persistence operations.
type Repository interface {
	// Save persists bank card data using the provided parameters.
	Save(ctx context.Context, params repository.SaveParams) error

	// Load retrieves bank card data using the provided parameters.
	Load(ctx context.Context, params repository.LoadParams) ([]*bankcard.BankCard, error)
}

// Service provides bank card business logic operations.
type Service struct {
	// r is the repository interface for bank card data persistence operations.
	r Repository
}

// NewService creates a new bank card service instance with the provided repository.
func NewService(r Repository) *Service {
	return &Service{r: r}
}

// Pull retrieves a specific bank card for the given user and card ID.
func (s *Service) Pull(ctx context.Context, params PullParams) (*BankCard, error) {
	cards, err := s.r.Load(ctx, repository.LoadParams{
		ID:     params.ID,
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load bank cards: %w", mapError(err))
	}
	if len(cards) == 0 {
		return nil, fmt.Errorf("bank card not found: %w", ErrBankCardNotFound)
	}
	return newBankCardFromDomain(cards[0]), nil
}

// List retrieves all bank cards for the specified user.
func (s *Service) List(ctx context.Context, params ListParams) ([]*BankCard, error) {
	cards, err := s.r.Load(ctx, repository.LoadParams{
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load bank cards: %w", mapError(err))
	}
	return newBankCardsFromDomain(cards), nil
}

// Push creates or updates a bank card with the provided parameters.
func (s *Service) Push(ctx context.Context, params *PushParams) (uuid.UUID, error) {
	card, err := bankcard.NewBankCard(&bankcard.NewBankCardParams{
		UserID:      params.UserID,
		CardNumber:  params.CardNumber,
		CardHolder:  params.CardHolder,
		ExpiryMonth: params.ExpiryMonth,
		ExpiryYear:  params.ExpiryYear,
		CVV:         params.CVV,
		Description: params.Description,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create bank card: %w", mapError(err))
	}

	if params.ID != uuid.Nil {
		if err := s.checkAccessToUpdate(ctx, params.ID, params.UserID); err != nil {
			return uuid.Nil, fmt.Errorf("access check for updating bank card failed: %w", err)
		}
		card.ID = params.ID
	}

	if err := s.r.Save(ctx, repository.SaveParams{Entity: card}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save bank card: %w", mapError(err))
	}
	return card.ID, nil
}

// checkAccessToUpdate verifies that a user has permission to update a specific bank card.
func (s *Service) checkAccessToUpdate(ctx context.Context, cardID, userID uuid.UUID) error {
	exists, err := s.Pull(ctx, PullParams{ID: cardID, UserID: userID})
	if err != nil {
		if errors.Is(err, ErrBankCardNotFound) {
			return fmt.Errorf("bank card for update not found: %w", err)
		}
		return fmt.Errorf("failed to pull existing bank card: %w", mapError(err))
	}
	if exists.UserID != userID {
		return fmt.Errorf("access denied to bank card: %w", ErrBankCardAccessDenied)
	}
	return nil
}
