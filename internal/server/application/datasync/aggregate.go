package datasync

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/google/uuid"
)

// BankCardService defines operations for synchronizing bank card data.
type BankCardService interface {
	List(ctx context.Context, params bankcard.ListParams) ([]*bankcard.BankCard, error)

	Push(ctx context.Context, params *bankcard.PushParams) (uuid.UUID, error)
}

// CredentialService defines operations for synchronizing credential data.
type CredentialService interface {
	List(ctx context.Context, params credential.ListParams) ([]*credential.Credential, error)

	Push(ctx context.Context, params *credential.PushParams) (uuid.UUID, error)
}

// NoteService defines operations for synchronizing note data.
type NoteService interface {
	List(ctx context.Context, params note.ListParams) ([]*note.Note, error)

	Push(ctx context.Context, params *note.PushParams) (uuid.UUID, error)
}

// FileDataService defines operations for synchronizing file data.
type FileDataService interface {
	List(ctx context.Context, params filedata.ListParams) ([]*filedata.FileData, error)

	Push(ctx context.Context, params *filedata.PushParams) (uuid.UUID, error)
}

// ServicesAggregator coordinates data synchronization operations across all data types.
type ServicesAggregator struct {
	// bankcardService handles bank card data operations.
	bankcardService BankCardService
	// credentialService handles credential data operations.
	credentialService CredentialService
	// noteService handles note data operations.
	noteService NoteService
	// fileDataService handles file data operations.
	fileDataService FileDataService
}

// NewServicesAggregator creates a new ServicesAggregator with the provided service dependencies.
func NewServicesAggregator(
	bankcardService BankCardService,
	credentialService CredentialService,
	noteService NoteService,
	fileDataService FileDataService,
) *ServicesAggregator {
	return &ServicesAggregator{
		bankcardService:   bankcardService,
		credentialService: credentialService,
		noteService:       noteService,
		fileDataService:   fileDataService,
	}
}

// PullBankCards retrieves all bank cards for the specified user.
func (a *ServicesAggregator) PullBankCards(
	ctx context.Context,
	userID uuid.UUID,
) ([]*bankcard.BankCard, error) {
	bankCards, err := a.bankcardService.List(ctx, bankcard.ListParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to pull bank cards: %w", err)
	}
	return bankCards, nil
}

// PullCredentials retrieves all credentials for the specified user.
func (a *ServicesAggregator) PullCredentials(
	ctx context.Context,
	userID uuid.UUID,
) ([]*credential.Credential, error) {
	credentials, err := a.credentialService.List(ctx, credential.ListParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to pull credentials: %w", err)
	}
	return credentials, nil
}

// PullNotes retrieves all notes for the specified user.
func (a *ServicesAggregator) PullNotes(ctx context.Context, userID uuid.UUID) ([]*note.Note, error) {
	notes, err := a.noteService.List(ctx, note.ListParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to pull notes: %w", err)
	}
	return notes, nil
}

// PushBankCards synchronizes bank card data to the server for the specified user.
func (a *ServicesAggregator) PushBankCards(
	ctx context.Context,
	userID uuid.UUID,
	cards []*bankcard.BankCard,
) error {
	for _, card := range cards {
		_, err := a.bankcardService.Push(ctx, &bankcard.PushParams{
			ID:          card.ID,
			UserID:      userID,
			CardNumber:  card.CardNumber,
			CardHolder:  card.CardHolder,
			ExpiryMonth: card.ExpiryMonth,
			ExpiryYear:  card.ExpiryYear,
			CVV:         card.CVV,
			Description: card.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to push bank card with ID %s: %w", card.ID, err)
		}
	}
	return nil
}

// PushCredentials synchronizes credential data to the server for the specified user.
func (a *ServicesAggregator) PushCredentials(
	ctx context.Context,
	userID uuid.UUID,
	credentials []*credential.Credential,
) error {
	for _, cred := range credentials {
		_, err := a.credentialService.Push(ctx, &credential.PushParams{
			ID:          cred.ID,
			UserID:      userID,
			Login:       cred.Login,
			Password:    cred.Password,
			Description: cred.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to push credential with ID %s: %w", cred.ID, err)
		}
	}
	return nil
}

// PushNotes synchronizes note data to the server for the specified user.
func (a *ServicesAggregator) PushNotes(ctx context.Context, userID uuid.UUID, notes []*note.Note) error {
	for _, n := range notes {
		_, err := a.noteService.Push(ctx, &note.PushParams{
			ID:          n.ID,
			UserID:      userID,
			Note:        n.Note,
			Description: n.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to push note with ID %s: %w", n.ID, err)
		}
	}
	return nil
}

// PullFiles retrieves all file data for the specified user.
func (a *ServicesAggregator) PullFiles(
	ctx context.Context,
	userID uuid.UUID,
) ([]*filedata.FileData, error) {
	files, err := a.fileDataService.List(ctx, filedata.ListParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to pull files: %w", err)
	}
	return files, nil
}

// PushFiles synchronizes file data to the server for the specified user.
func (a *ServicesAggregator) PushFiles(
	ctx context.Context,
	userID uuid.UUID,
	files []*filedata.FileData,
) error {
	for _, f := range files {
		_, err := a.fileDataService.Push(ctx, &filedata.PushParams{
			ID:          f.ID,
			UserID:      userID,
			StorageKey:  f.StorageKey,
			Description: f.Description,
			Data:        f.Data,
		})
		if err != nil {
			return fmt.Errorf("failed to push file with ID %s: %w", f.ID, err)
		}
	}
	return nil
}
