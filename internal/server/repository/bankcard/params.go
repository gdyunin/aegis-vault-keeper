package bankcard

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/google/uuid"
)

// SaveParams contains the parameters for saving a bank card entity to the repository.
type SaveParams struct {
	// Entity contains the bank card data to be persisted.
	Entity *bankcard.BankCard
}

// LoadParams contains the parameters for loading bank card entities from the repository.
type LoadParams struct {
	// ID contains the specific bank card identifier for single record lookup (optional).
	ID uuid.UUID
	// UserID contains the user identifier for filtering bank cards by owner (required).
	UserID uuid.UUID
}
