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

// makePullBankCardsTask creates a task function that pulls bank cards for a user and stores them in the target slice.
func (s *Service) makePullBankCardsTask(
	ctx context.Context,
	userID uuid.UUID,
	target *[]*bankcard.BankCard,
) func() error {
	return func() error {
		result, err := s.aggr.PullBankCards(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to pull bank cards: %w", err)
		}
		*target = result
		return nil
	}
}

// makePullCredentialsTask creates a task function that pulls credentials for a user and stores them
// in the target slice.
func (s *Service) makePullCredentialsTask(
	ctx context.Context,
	userID uuid.UUID,
	target *[]*credential.Credential,
) func() error {
	return func() error {
		result, err := s.aggr.PullCredentials(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to pull credentials: %w", err)
		}
		*target = result
		return nil
	}
}

// makePullNotesTask creates a task function that pulls notes for a user and stores them in the target slice.
func (s *Service) makePullNotesTask(
	ctx context.Context,
	userID uuid.UUID,
	target *[]*note.Note,
) func() error {
	return func() error {
		result, err := s.aggr.PullNotes(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to pull notes: %w", err)
		}
		*target = result
		return nil
	}
}

// makePullFilesTask creates a task function that pulls file data for a user and stores them in the target slice.
func (s *Service) makePullFilesTask(
	ctx context.Context,
	userID uuid.UUID,
	target *[]*filedata.FileData,
) func() error {
	return func() error {
		result, err := s.aggr.PullFiles(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to pull files: %w", err)
		}
		*target = result
		return nil
	}
}
