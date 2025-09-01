package datasync

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/datasync"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/note"
	"github.com/google/uuid"
)

// SyncPayload represents a complete set of user data for synchronization.
type SyncPayload struct {
	// BankCards contains the user's bank card data for synchronization.
	BankCards []*bankcard.BankCard `json:"bankcards,omitzero"` // User's bank cards
	// Credentials contains the user's credential data for synchronization.
	Credentials []*credential.Credential `json:"credentials,omitzero"` // User's credentials
	// Notes contains the user's note data for synchronization.
	Notes []*note.Note `json:"notes,omitzero"` // User's notes
	// Files contains the user's file data for synchronization.
	Files []*filedata.FileData `json:"files,omitzero"` // User's files
}

// ToApp converts the delivery layer SyncPayload to application layer format.
func (p *SyncPayload) ToApp(userID uuid.UUID) *datasync.SyncPayload {
	if p == nil {
		return nil
	}
	return &datasync.SyncPayload{
		UserID:      userID,
		BankCards:   bankcard.BankCardsToApp(p.BankCards, userID),
		Credentials: credential.CredentialsToApp(p.Credentials, userID),
		Notes:       note.NotesToApp(p.Notes, userID),
		Files:       filedata.FilesToApp(p.Files, userID),
	}
}

// NewSyncPayloadFromApp creates a delivery layer SyncPayload from application layer data.
func NewSyncPayloadFromApp(sp *datasync.SyncPayload) *SyncPayload {
	if sp == nil {
		return nil
	}
	return &SyncPayload{
		BankCards:   bankcard.NewBankCardsFromApp(sp.BankCards),
		Credentials: credential.NewCredentialsFromApp(sp.Credentials),
		Notes:       note.NewNotesFromApp(sp.Notes),
		Files:       filedata.NewFileDataListFromApp(sp.Files),
	}
}

// isEmpty checks if the sync payload contains no data.
func (p *SyncPayload) isEmpty() bool {
	return len(p.BankCards) == 0 && len(p.Credentials) == 0 && len(p.Notes) == 0 && len(p.Files) == 0
}
