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

func TestNewServicesAggregator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bankcardService   BankCardService
		credentialService CredentialService
		noteService       NoteService
		fileDataService   FileDataService
		name              string
	}{
		{
			name:              "valid services",
			bankcardService:   &mockBankCardService{},
			credentialService: &mockCredentialService{},
			noteService:       &mockNoteService{},
			fileDataService:   &mockFileDataService{},
		},
		{
			name:              "nil services",
			bankcardService:   nil,
			credentialService: nil,
			noteService:       nil,
			fileDataService:   nil,
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

			assert.NotNil(t, aggr)
			assert.Equal(t, tt.bankcardService, aggr.bankcardService)
			assert.Equal(t, tt.credentialService, aggr.credentialService)
			assert.Equal(t, tt.noteService, aggr.noteService)
			assert.Equal(t, tt.fileDataService, aggr.fileDataService)
		})
	}
}

func TestServicesAggregator_PullBankCards(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedCards := []*bankcard.BankCard{
		{ID: uuid.New(), UserID: userID, CardNumber: "1234567890123456"},
		{ID: uuid.New(), UserID: userID, CardNumber: "9876543210987654"},
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
			name: "successful pull",
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

			result, err := aggr.PullBankCards(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestServicesAggregator_PullCredentials(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedCredentials := []*credential.Credential{
		{ID: uuid.New(), UserID: userID, Login: "user1@example.com"},
		{ID: uuid.New(), UserID: userID, Login: "user2@example.com"},
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
			name: "successful pull",
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

			result, err := aggr.PullCredentials(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestServicesAggregator_PullNotes(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedNotes := []*note.Note{
		{ID: uuid.New(), UserID: userID, Note: "Note 1"},
		{ID: uuid.New(), UserID: userID, Note: "Note 2"},
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
			name: "successful pull",
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

			result, err := aggr.PullNotes(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestServicesAggregator_PullFiles(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedFiles := []*filedata.FileData{
		{ID: uuid.New(), UserID: userID, StorageKey: "file1.txt"},
		{ID: uuid.New(), UserID: userID, StorageKey: "file2.txt"},
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
			name: "successful pull",
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

			result, err := aggr.PullFiles(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestServicesAggregator_PushBankCards(t *testing.T) {
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
			name: "successful push",
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

			err := aggr.PushBankCards(context.Background(), tt.userID, tt.cards)

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

func TestServicesAggregator_PushCredentials(t *testing.T) {
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
			name: "successful push",
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

			err := aggr.PushCredentials(context.Background(), tt.userID, tt.credentials)

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

func TestServicesAggregator_PushNotes(t *testing.T) {
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
			name: "successful push",
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

			err := aggr.PushNotes(context.Background(), tt.userID, tt.notes)

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

func TestServicesAggregator_PushFiles(t *testing.T) {
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
			name: "successful push",
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

			err := aggr.PushFiles(context.Background(), tt.userID, tt.files)

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
