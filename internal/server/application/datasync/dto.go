package datasync

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/google/uuid"
)

// SyncPayload represents a complete data synchronization payload containing all user data types.
type SyncPayload struct {
	// BankCards contains the user's bank card data for synchronization.
	BankCards []*bankcard.BankCard
	// Credentials contains the user's credential data for synchronization.
	Credentials []*credential.Credential
	// Notes contains the user's note data for synchronization.
	Notes []*note.Note
	// Files contains the user's file data for synchronization.
	Files []*filedata.FileData
	// UserID identifies the user owning this data payload.
	UserID uuid.UUID
}
