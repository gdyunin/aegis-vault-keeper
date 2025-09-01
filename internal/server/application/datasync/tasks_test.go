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

func TestService_makePullBankCardsTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedCards := []*bankcard.BankCard{
		{ID: uuid.New(), UserID: userID, CardNumber: "1234567890123456"},
	}

	tests := []struct {
		bankcardService *mockBankCardService
		name            string
		errContains     string
		want            []*bankcard.BankCard
		userID          uuid.UUID
		wantErr         bool
	}{
		{
			name: "successful task execution",
			bankcardService: &mockBankCardService{
				listResult: expectedCards,
			},
			userID:  userID,
			want:    expectedCards,
			wantErr: false,
		},
		{
			name: "service error",
			bankcardService: &mockBankCardService{
				listError: errors.New("service error"),
			},
			userID:      userID,
			want:        nil,
			wantErr:     true,
			errContains: "failed to pull bank cards",
		},
		{
			name: "empty result",
			bankcardService: &mockBankCardService{
				listResult: []*bankcard.BankCard{},
			},
			userID:  userID,
			want:    []*bankcard.BankCard{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				tt.bankcardService,
				&mockCredentialService{},
				&mockNoteService{},
				&mockFileDataService{},
			)
			service := NewService(aggr)

			var target []*bankcard.BankCard
			task := service.makePullBankCardsTask(context.Background(), tt.userID, &target)

			err := task()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, target)
			}
		})
	}
}

func TestService_makePullCredentialsTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedCredentials := []*credential.Credential{
		{ID: uuid.New(), UserID: userID, Login: "test@example.com"},
	}

	tests := []struct {
		credentialService *mockCredentialService
		name              string
		errContains       string
		want              []*credential.Credential
		userID            uuid.UUID
		wantErr           bool
	}{
		{
			name: "successful task execution",
			credentialService: &mockCredentialService{
				listResult: expectedCredentials,
			},
			userID:  userID,
			want:    expectedCredentials,
			wantErr: false,
		},
		{
			name: "service error",
			credentialService: &mockCredentialService{
				listError: errors.New("service error"),
			},
			userID:      userID,
			want:        nil,
			wantErr:     true,
			errContains: "failed to pull credentials",
		},
		{
			name: "empty result",
			credentialService: &mockCredentialService{
				listResult: []*credential.Credential{},
			},
			userID:  userID,
			want:    []*credential.Credential{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				&mockBankCardService{},
				tt.credentialService,
				&mockNoteService{},
				&mockFileDataService{},
			)
			service := NewService(aggr)

			var target []*credential.Credential
			task := service.makePullCredentialsTask(context.Background(), tt.userID, &target)

			err := task()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, target)
			}
		})
	}
}

func TestService_makePullNotesTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedNotes := []*note.Note{
		{ID: uuid.New(), UserID: userID, Note: "Test Note"},
	}

	tests := []struct {
		noteService *mockNoteService
		name        string
		errContains string
		want        []*note.Note
		userID      uuid.UUID
		wantErr     bool
	}{
		{
			name: "successful task execution",
			noteService: &mockNoteService{
				listResult: expectedNotes,
			},
			userID:  userID,
			want:    expectedNotes,
			wantErr: false,
		},
		{
			name: "service error",
			noteService: &mockNoteService{
				listError: errors.New("service error"),
			},
			userID:      userID,
			want:        nil,
			wantErr:     true,
			errContains: "failed to pull notes",
		},
		{
			name: "empty result",
			noteService: &mockNoteService{
				listResult: []*note.Note{},
			},
			userID:  userID,
			want:    []*note.Note{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				&mockBankCardService{},
				&mockCredentialService{},
				tt.noteService,
				&mockFileDataService{},
			)
			service := NewService(aggr)

			var target []*note.Note
			task := service.makePullNotesTask(context.Background(), tt.userID, &target)

			err := task()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, target)
			}
		})
	}
}

func TestService_makePullFilesTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedFiles := []*filedata.FileData{
		{ID: uuid.New(), UserID: userID, StorageKey: "test.txt"},
	}

	tests := []struct {
		fileDataService *mockFileDataService
		name            string
		errContains     string
		want            []*filedata.FileData
		userID          uuid.UUID
		wantErr         bool
	}{
		{
			name: "successful task execution",
			fileDataService: &mockFileDataService{
				listResult: expectedFiles,
			},
			userID:  userID,
			want:    expectedFiles,
			wantErr: false,
		},
		{
			name: "service error",
			fileDataService: &mockFileDataService{
				listError: errors.New("service error"),
			},
			userID:      userID,
			want:        nil,
			wantErr:     true,
			errContains: "failed to pull files",
		},
		{
			name: "empty result",
			fileDataService: &mockFileDataService{
				listResult: []*filedata.FileData{},
			},
			userID:  userID,
			want:    []*filedata.FileData{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				&mockBankCardService{},
				&mockCredentialService{},
				&mockNoteService{},
				tt.fileDataService,
			)
			service := NewService(aggr)

			var target []*filedata.FileData
			task := service.makePullFilesTask(context.Background(), tt.userID, &target)

			err := task()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, target)
			}
		})
	}
}

func TestService_makePushBankCardsTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	cards := []*bankcard.BankCard{
		{ID: uuid.New(), UserID: userID, CardNumber: "1234567890123456"},
	}

	tests := []struct {
		bankcardService *mockBankCardService
		name            string
		errContains     string
		cards           []*bankcard.BankCard
		userID          uuid.UUID
		wantErr         bool
	}{
		{
			name: "successful task execution",
			bankcardService: &mockBankCardService{
				pushResult: uuid.New(),
			},
			userID:  userID,
			cards:   cards,
			wantErr: false,
		},
		{
			name: "service error",
			bankcardService: &mockBankCardService{
				pushError: errors.New("service error"),
			},
			userID:      userID,
			cards:       cards,
			wantErr:     true,
			errContains: "failed to push bank card",
		},
		{
			name: "empty cards",
			bankcardService: &mockBankCardService{
				pushResult: uuid.New(),
			},
			userID:  userID,
			cards:   []*bankcard.BankCard{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				tt.bankcardService,
				&mockCredentialService{},
				&mockNoteService{},
				&mockFileDataService{},
			)
			service := NewService(aggr)

			task := service.makePushBankCardsTask(context.Background(), tt.userID, tt.cards)

			err := task()

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

func TestService_makePushCredentialsTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	credentials := []*credential.Credential{
		{ID: uuid.New(), UserID: userID, Login: "test@example.com"},
	}

	tests := []struct {
		credentialService *mockCredentialService
		name              string
		errContains       string
		credentials       []*credential.Credential
		userID            uuid.UUID
		wantErr           bool
	}{
		{
			name: "successful task execution",
			credentialService: &mockCredentialService{
				pushResult: uuid.New(),
			},
			userID:      userID,
			credentials: credentials,
			wantErr:     false,
		},
		{
			name: "service error",
			credentialService: &mockCredentialService{
				pushError: errors.New("service error"),
			},
			userID:      userID,
			credentials: credentials,
			wantErr:     true,
			errContains: "failed to push credential",
		},
		{
			name: "empty credentials",
			credentialService: &mockCredentialService{
				pushResult: uuid.New(),
			},
			userID:      userID,
			credentials: []*credential.Credential{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				&mockBankCardService{},
				tt.credentialService,
				&mockNoteService{},
				&mockFileDataService{},
			)
			service := NewService(aggr)

			task := service.makePushCredentialsTask(context.Background(), tt.userID, tt.credentials)

			err := task()

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

func TestService_makePushNotesTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notes := []*note.Note{
		{ID: uuid.New(), UserID: userID, Note: "Test Note"},
	}

	tests := []struct {
		noteService *mockNoteService
		name        string
		errContains string
		notes       []*note.Note
		userID      uuid.UUID
		wantErr     bool
	}{
		{
			name: "successful task execution",
			noteService: &mockNoteService{
				pushResult: uuid.New(),
			},
			userID:  userID,
			notes:   notes,
			wantErr: false,
		},
		{
			name: "service error",
			noteService: &mockNoteService{
				pushError: errors.New("service error"),
			},
			userID:      userID,
			notes:       notes,
			wantErr:     true,
			errContains: "failed to push note",
		},
		{
			name: "empty notes",
			noteService: &mockNoteService{
				pushResult: uuid.New(),
			},
			userID:  userID,
			notes:   []*note.Note{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				&mockBankCardService{},
				&mockCredentialService{},
				tt.noteService,
				&mockFileDataService{},
			)
			service := NewService(aggr)

			task := service.makePushNotesTask(context.Background(), tt.userID, tt.notes)

			err := task()

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

func TestService_makePushFilesTask(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	files := []*filedata.FileData{
		{ID: uuid.New(), UserID: userID, StorageKey: "test.txt"},
	}

	tests := []struct {
		fileDataService *mockFileDataService
		name            string
		errContains     string
		files           []*filedata.FileData
		userID          uuid.UUID
		wantErr         bool
	}{
		{
			name: "successful task execution",
			fileDataService: &mockFileDataService{
				pushResult: uuid.New(),
			},
			userID:  userID,
			files:   files,
			wantErr: false,
		},
		{
			name: "service error",
			fileDataService: &mockFileDataService{
				pushError: errors.New("service error"),
			},
			userID:      userID,
			files:       files,
			wantErr:     true,
			errContains: "failed to push file",
		},
		{
			name: "empty files",
			fileDataService: &mockFileDataService{
				pushResult: uuid.New(),
			},
			userID:  userID,
			files:   []*filedata.FileData{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggr := NewServicesAggregator(
				&mockBankCardService{},
				&mockCredentialService{},
				&mockNoteService{},
				tt.fileDataService,
			)
			service := NewService(aggr)

			task := service.makePushFilesTask(context.Background(), tt.userID, tt.files)

			err := task()

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
