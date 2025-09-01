package bankcard

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/google/uuid"
)

// BankCard represents a bank card entity for API transfer.
type BankCard struct {
	// UpdatedAt contains the timestamp when this card was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty"   example:"2023-12-01T10:00:00Z"`
	// CardNumber contains the 13-19 digit payment card number (PCI DSS sensitive data).
	CardNumber string `json:"card_number,omitempty"  example:"4242424242424242"`
	// CardHolder contains the name printed on the card (may differ from account holder).
	CardHolder string `json:"card_holder,omitempty"  example:"John Doe"`
	// ExpiryMonth contains the two-digit expiration month (01-12).
	ExpiryMonth string `json:"expiry_month,omitempty" example:"12"`
	// ExpiryYear contains the two-digit expiration year (YY format).
	ExpiryYear string `json:"expiry_year,omitempty"  example:"2025"`
	// CVV contains the 3-4 digit card verification value (PCI DSS sensitive data).
	CVV string `json:"cvv,omitempty"          example:"123"`
	// Description contains optional user-provided notes about this card.
	Description string `json:"description,omitempty"  example:"Main credit card"`
	// ID contains the unique identifier for this bank card record.
	ID uuid.UUID `json:"id,omitempty"           example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ToApp converts this DTO to an application layer BankCard entity with the specified user ID.
func (b *BankCard) ToApp(userID uuid.UUID) *bankcard.BankCard {
	if b == nil {
		return nil
	}
	return &bankcard.BankCard{
		ID:          b.ID,
		UserID:      userID,
		CardNumber:  b.CardNumber,
		CardHolder:  b.CardHolder,
		ExpiryMonth: b.ExpiryMonth,
		ExpiryYear:  b.ExpiryYear,
		CVV:         b.CVV,
		Description: b.Description,
		UpdatedAt:   b.UpdatedAt,
	}
}

// BankCardsToApp converts a slice of DTOs to application layer BankCard entities with the specified user ID.
func BankCardsToApp(bcs []*BankCard, userID uuid.UUID) []*bankcard.BankCard {
	if bcs == nil {
		return nil
	}
	result := make([]*bankcard.BankCard, 0, len(bcs))
	for _, bc := range bcs {
		result = append(result, bc.ToApp(userID))
	}
	return result
}

// NewBankCardFromApp creates a DTO from an application layer BankCard entity.
func NewBankCardFromApp(bc *bankcard.BankCard) *BankCard {
	if bc == nil {
		return nil
	}
	return &BankCard{
		ID:          bc.ID,
		CardNumber:  bc.CardNumber,
		CardHolder:  bc.CardHolder,
		ExpiryMonth: bc.ExpiryMonth,
		ExpiryYear:  bc.ExpiryYear,
		CVV:         bc.CVV,
		Description: bc.Description,
		UpdatedAt:   bc.UpdatedAt,
	}
}

// NewBankCardsFromApp creates DTOs from a slice of application layer BankCard entities.
func NewBankCardsFromApp(bcs []*bankcard.BankCard) []*BankCard {
	if bcs == nil {
		return nil
	}
	result := make([]*BankCard, 0, len(bcs))
	for _, bc := range bcs {
		result = append(result, NewBankCardFromApp(bc))
	}
	return result
}

// PushRequest represents the data required to create or update a bank card.
type PushRequest struct {
	// CardNumber contains the 13-19 digit payment card number (required, PCI DSS sensitive).
	CardNumber string `json:"card_number"          binding:"required" example:"1234567812345678"`
	// CardHolder contains the name as printed on the card (required, max 255 chars).
	CardHolder string `json:"card_holder"          binding:"required" example:"John Doe"`
	// ExpiryMonth contains the two-digit expiration month 01-12 (required).
	ExpiryMonth string `json:"expiry_month"         binding:"required" example:"12"`
	// ExpiryYear contains the two-digit expiration year YY format (required).
	ExpiryYear string `json:"expiry_year"          binding:"required" example:"25"`
	// CVV contains the 3-4 digit card verification value (required, PCI DSS sensitive).
	CVV string `json:"cvv"                  binding:"required" example:"123"`
	// Description contains optional user notes about this card (max 500 chars).
	Description string `json:"description,omitzero"                    example:"Main credit card"`
}

// PullRequest represents the request to retrieve a specific bank card.
type PullRequest struct {
	// ID contains the UUID of the bank card to retrieve (required, must be valid UUID).
	ID string `uri:"id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// PushResponse represents the response after creating or updating a bank card.
type PushResponse struct {
	// ID contains the UUID of the created or updated bank card.
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// PullResponse represents the response containing a specific bank card.
type PullResponse struct {
	// BankCard contains the requested bank card data.
	BankCard *BankCard `json:"bankcard"`
}

// ListResponse represents the response containing all user's bank cards.
type ListResponse struct {
	// BankCards contains the list of all bank cards belonging to the user.
	BankCards []*BankCard `json:"bankcards"`
}
