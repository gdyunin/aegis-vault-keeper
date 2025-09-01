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

// makePushBankCardsTask creates a task function that pushes bank cards for a user to the server.
func (s *Service) makePushBankCardsTask(
	ctx context.Context,
	userID uuid.UUID,
	cards []*bankcard.BankCard,
) func() error {
	return func() error {
		if err := s.aggr.PushBankCards(ctx, userID, cards); err != nil {
			return fmt.Errorf("failed to push bank cards: %w", err)
		}
		return nil
	}
}

// makePushCredentialsTask creates a task function that pushes credentials for a user to the server.
func (s *Service) makePushCredentialsTask(
	ctx context.Context,
	userID uuid.UUID,
	creds []*credential.Credential,
) func() error {
	return func() error {
		if err := s.aggr.PushCredentials(ctx, userID, creds); err != nil {
			return fmt.Errorf("failed to push credentials: %w", err)
		}
		return nil
	}
}

// makePushNotesTask creates a task function that pushes notes for a user to the server.
func (s *Service) makePushNotesTask(ctx context.Context, userID uuid.UUID, notes []*note.Note) func() error {
	return func() error {
		if err := s.aggr.PushNotes(ctx, userID, notes); err != nil {
			return fmt.Errorf("failed to push notes: %w", err)
		}
		return nil
	}
}

// makePushFilesTask creates a task function that pushes file data for a user to the server.
func (s *Service) makePushFilesTask(
	ctx context.Context,
	userID uuid.UUID,
	files []*filedata.FileData,
) func() error {
	return func() error {
		if err := s.aggr.PushFiles(ctx, userID, files); err != nil {
			return fmt.Errorf("failed to push files: %w", err)
		}
		return nil
	}
}
