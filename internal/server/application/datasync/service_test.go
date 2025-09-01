package datasync

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
)

// Mock implementations for testing.
type mockBankCardService struct {
	listError  error
	pushError  error
	listResult []*bankcard.BankCard
	pushResult uuid.UUID
}

func (m *mockBankCardService) List(
	ctx context.Context,
	params bankcard.ListParams,
) ([]*bankcard.BankCard, error) {
	return m.listResult, m.listError
}

func (m *mockBankCardService) Push(ctx context.Context, params *bankcard.PushParams) (uuid.UUID, error) {
	return m.pushResult, m.pushError
}

type mockCredentialService struct {
	listError  error
	pushError  error
	listResult []*credential.Credential
	pushResult uuid.UUID
}

func (m *mockCredentialService) List(
	ctx context.Context, params credential.ListParams,
) ([]*credential.Credential, error) {
	return m.listResult, m.listError
}

func (m *mockCredentialService) Push(ctx context.Context, params *credential.PushParams) (uuid.UUID, error) {
	return m.pushResult, m.pushError
}

type mockNoteService struct {
	listError  error
	pushError  error
	listResult []*note.Note
	pushResult uuid.UUID
}

func (m *mockNoteService) List(ctx context.Context, params note.ListParams) ([]*note.Note, error) {
	return m.listResult, m.listError
}

func (m *mockNoteService) Push(ctx context.Context, params *note.PushParams) (uuid.UUID, error) {
	return m.pushResult, m.pushError
}

type mockFileDataService struct {
	listError  error
	pushError  error
	listResult []*filedata.FileData
	pushResult uuid.UUID
}

func (m *mockFileDataService) List(
	ctx context.Context,
	params filedata.ListParams,
) ([]*filedata.FileData, error) {
	return m.listResult, m.listError
}

func (m *mockFileDataService) Push(ctx context.Context, params *filedata.PushParams) (uuid.UUID, error) {
	return m.pushResult, m.pushError
}

func TestNewService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		aggr *ServicesAggregator
		name string
	}{
		{
			name: "valid aggregator",
			aggr: &ServicesAggregator{},
		},
		{
			name: "nil aggregator",
			aggr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(tt.aggr)

			assert.NotNil(t, service)
			assert.Equal(t, tt.aggr, service.aggr)
		})
	}
}

func TestService_Pull(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		bankcardService   *mockBankCardService
		credentialService *mockCredentialService
		noteService       *mockNoteService
		fileDataService   *mockFileDataService
		name              string
		errContains       string
		wantErr           bool
	}{
		{
			name: "successful pull all data",
			bankcardService: &mockBankCardService{
				listResult: []*bankcard.BankCard{
					{ID: uuid.New(), UserID: userID, CardNumber: "1234567890123456"},
				},
			},
			credentialService: &mockCredentialService{
				listResult: []*credential.Credential{
					{ID: uuid.New(), UserID: userID, Login: "test@example.com"},
				},
			},
			noteService: &mockNoteService{
				listResult: []*note.Note{
					{ID: uuid.New(), UserID: userID, Note: "Test Note"},
				},
			},
			fileDataService: &mockFileDataService{
				listResult: []*filedata.FileData{
					{ID: uuid.New(), UserID: userID, StorageKey: "test.txt"},
				},
			},
			wantErr: false,
		},
		{
			name: "bankcard service error",
			bankcardService: &mockBankCardService{
				listError: errors.New("bankcard service error"),
			},
			credentialService: &mockCredentialService{
				listResult: []*credential.Credential{},
			},
			noteService: &mockNoteService{
				listResult: []*note.Note{},
			},
			fileDataService: &mockFileDataService{
				listResult: []*filedata.FileData{},
			},
			wantErr:     true,
			errContains: "failed to pull data",
		},
		{
			name: "credential service error",
			bankcardService: &mockBankCardService{
				listResult: []*bankcard.BankCard{},
			},
			credentialService: &mockCredentialService{
				listError: errors.New("credential service error"),
			},
			noteService: &mockNoteService{
				listResult: []*note.Note{},
			},
			fileDataService: &mockFileDataService{
				listResult: []*filedata.FileData{},
			},
			wantErr:     true,
			errContains: "failed to pull data",
		},
		{
			name: "note service error",
			bankcardService: &mockBankCardService{
				listResult: []*bankcard.BankCard{},
			},
			credentialService: &mockCredentialService{
				listResult: []*credential.Credential{},
			},
			noteService: &mockNoteService{
				listError: errors.New("note service error"),
			},
			fileDataService: &mockFileDataService{
				listResult: []*filedata.FileData{},
			},
			wantErr:     true,
			errContains: "failed to pull data",
		},
		{
			name: "file data service error",
			bankcardService: &mockBankCardService{
				listResult: []*bankcard.BankCard{},
			},
			credentialService: &mockCredentialService{
				listResult: []*credential.Credential{},
			},
			noteService: &mockNoteService{
				listResult: []*note.Note{},
			},
			fileDataService: &mockFileDataService{
				listError: errors.New("file data service error"),
			},
			wantErr:     true,
			errContains: "failed to pull data",
		},
		{
			name: "all services return empty results",
			bankcardService: &mockBankCardService{
				listResult: []*bankcard.BankCard{},
			},
			credentialService: &mockCredentialService{
				listResult: []*credential.Credential{},
			},
			noteService: &mockNoteService{
				listResult: []*note.Note{},
			},
			fileDataService: &mockFileDataService{
				listResult: []*filedata.FileData{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				tt.bankcardService,
				tt.credentialService,
				tt.noteService,
				tt.fileDataService,
			)
			service := NewService(aggr)

			result, err := service.Pull(context.Background(), userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, userID, result.UserID)
				assert.Equal(t, tt.bankcardService.listResult, result.BankCards)
				assert.Equal(t, tt.credentialService.listResult, result.Credentials)
				assert.Equal(t, tt.noteService.listResult, result.Notes)
				assert.Equal(t, tt.fileDataService.listResult, result.Files)
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		payload           *SyncPayload
		bankcardService   *mockBankCardService
		credentialService *mockCredentialService
		noteService       *mockNoteService
		fileDataService   *mockFileDataService
		name              string
		errContains       string
		wantErr           bool
	}{
		{
			name: "successful push all data",
			payload: &SyncPayload{
				UserID: userID,
				BankCards: []*bankcard.BankCard{
					{ID: uuid.New(), UserID: userID, CardNumber: "1234567890123456"},
				},
				Credentials: []*credential.Credential{
					{ID: uuid.New(), UserID: userID, Login: "test@example.com"},
				},
				Notes: []*note.Note{
					{ID: uuid.New(), UserID: userID, Note: "Test Note"},
				},
				Files: []*filedata.FileData{
					{ID: uuid.New(), UserID: userID, StorageKey: "test.txt"},
				},
			},
			bankcardService: &mockBankCardService{
				pushResult: uuid.New(),
			},
			credentialService: &mockCredentialService{
				pushResult: uuid.New(),
			},
			noteService: &mockNoteService{
				pushResult: uuid.New(),
			},
			fileDataService: &mockFileDataService{
				pushResult: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "bankcard service push error",
			payload: &SyncPayload{
				UserID:      userID,
				BankCards:   []*bankcard.BankCard{{ID: uuid.New()}},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			bankcardService: &mockBankCardService{
				pushError: errors.New("bankcard push error"),
			},
			credentialService: &mockCredentialService{},
			noteService:       &mockNoteService{},
			fileDataService:   &mockFileDataService{},
			wantErr:           true,
			errContains:       "failed to push data",
		},
		{
			name: "credential service push error",
			payload: &SyncPayload{
				UserID:      userID,
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{{ID: uuid.New()}},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			bankcardService: &mockBankCardService{},
			credentialService: &mockCredentialService{
				pushError: errors.New("credential push error"),
			},
			noteService:     &mockNoteService{},
			fileDataService: &mockFileDataService{},
			wantErr:         true,
			errContains:     "failed to push data",
		},
		{
			name: "note service push error",
			payload: &SyncPayload{
				UserID:      userID,
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{{ID: uuid.New()}},
				Files:       []*filedata.FileData{},
			},
			bankcardService:   &mockBankCardService{},
			credentialService: &mockCredentialService{},
			noteService: &mockNoteService{
				pushError: errors.New("note push error"),
			},
			fileDataService: &mockFileDataService{},
			wantErr:         true,
			errContains:     "failed to push data",
		},
		{
			name: "file data service push error",
			payload: &SyncPayload{
				UserID:      userID,
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{{ID: uuid.New()}},
			},
			bankcardService:   &mockBankCardService{},
			credentialService: &mockCredentialService{},
			noteService:       &mockNoteService{},
			fileDataService: &mockFileDataService{
				pushError: errors.New("file data push error"),
			},
			wantErr:     true,
			errContains: "failed to push data",
		},
		{
			name: "empty payload",
			payload: &SyncPayload{
				UserID:      userID,
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			bankcardService:   &mockBankCardService{},
			credentialService: &mockCredentialService{},
			noteService:       &mockNoteService{},
			fileDataService:   &mockFileDataService{},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				tt.bankcardService,
				tt.credentialService,
				tt.noteService,
				tt.fileDataService,
			)
			service := NewService(aggr)

			err := service.Push(context.Background(), tt.payload)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
