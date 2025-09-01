package bankcard

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/google/uuid"
)

// BankCard represents a bank card data transfer object for the application layer.
type BankCard struct {
	// UpdatedAt specifies when the bank card was last updated.
	UpdatedAt time.Time
	// CardNumber contains the bank card number.
	CardNumber string
	// CardHolder contains the name of the card holder.
	CardHolder string
	// ExpiryMonth contains the expiry month of the card.
	ExpiryMonth string
	// ExpiryYear contains the expiry year of the card.
	ExpiryYear string
	// CVV contains the card verification value.
	CVV string
	// Description contains an optional description of the card.
	Description string
	// ID is the unique identifier of the bank card.
	ID uuid.UUID
	// UserID is the identifier of the user who owns the card.
	UserID uuid.UUID
}

// newBankCardFromDomain converts a domain bank card entity to application DTO.
func newBankCardFromDomain(bc *bankcard.BankCard) *BankCard {
	if bc == nil {
		return nil
	}
	return &BankCard{
		ID:          bc.ID,
		UserID:      bc.UserID,
		CardNumber:  string(bc.CardNumber),
		CardHolder:  string(bc.CardHolder),
		ExpiryMonth: string(bc.ExpiryMonth),
		ExpiryYear:  string(bc.ExpiryYear),
		CVV:         string(bc.CVV),
		Description: string(bc.Description),
		UpdatedAt:   bc.UpdatedAt,
	}
}

// newBankCardsFromDomain converts a slice of domain bank card entities to application DTOs.
func newBankCardsFromDomain(bcs []*bankcard.BankCard) []*BankCard {
	result := make([]*BankCard, 0, len(bcs))
	for _, bc := range bcs {
		result = append(result, newBankCardFromDomain(bc))
	}
	return result
}

// PullParams contains parameters for retrieving a specific bank card.
type PullParams struct {
	// ID is the unique identifier of the bank card to retrieve.
	ID uuid.UUID
	// UserID is the identifier of the user who owns the card.
	UserID uuid.UUID
}

// ListParams contains parameters for listing bank cards.
type ListParams struct {
	// UserID is the identifier of the user whose cards to list.
	UserID uuid.UUID
}

// PushParams contains parameters for creating or updating a bank card.
type PushParams struct {
	// CardNumber contains the bank card number.
	CardNumber string
	// CardHolder contains the name of the card holder.
	CardHolder string
	// ExpiryMonth contains the expiry month of the card.
	ExpiryMonth string
	// ExpiryYear contains the expiry year of the card.
	ExpiryYear string
	// CVV contains the card verification value.
	CVV string
	// Description contains an optional description of the card.
	Description string
	// ID is the unique identifier of the bank card.
	ID uuid.UUID
	// UserID is the identifier of the user who owns the card.
	UserID uuid.UUID
}
