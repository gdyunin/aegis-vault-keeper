package datasync

import (
	"context"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Service coordinates data synchronization operations across all data types using concurrent tasks.
type Service struct {
	// aggr provides aggregated access to all application layer services for data synchronization.
	aggr *ServicesAggregator
}

// NewService creates a new Service with the provided services aggregator.
func NewService(aggr *ServicesAggregator) *Service {
	return &Service{aggr: aggr}
}

// Pull retrieves all user data concurrently and returns it as a SyncPayload.
func (s *Service) Pull(ctx context.Context, userID uuid.UUID) (*SyncPayload, error) {
	var (
		cards []*bankcard.BankCard
		creds []*credential.Credential
		notes []*note.Note
		files []*filedata.FileData
	)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(s.makePullBankCardsTask(ctx, userID, &cards))
	g.Go(s.makePullCredentialsTask(ctx, userID, &creds))
	g.Go(s.makePullNotesTask(ctx, userID, &notes))
	g.Go(s.makePullFilesTask(ctx, userID, &files))

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to pull data: %w", err)
	}

	return &SyncPayload{
		UserID:      userID,
		BankCards:   cards,
		Credentials: creds,
		Notes:       notes,
		Files:       files,
	}, nil
}

// Push synchronizes all data in the payload to the server concurrently.
func (s *Service) Push(ctx context.Context, payload *SyncPayload) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(s.makePushBankCardsTask(ctx, payload.UserID, payload.BankCards))
	g.Go(s.makePushCredentialsTask(ctx, payload.UserID, payload.Credentials))
	g.Go(s.makePushNotesTask(ctx, payload.UserID, payload.Notes))
	g.Go(s.makePushFilesTask(ctx, payload.UserID, payload.Files))

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to push data: %w", err)
	}
	return nil
}
