package datasync

import (
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/note"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSyncPayload_ToApp(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		payload     *SyncPayload
		checkResult func(*testing.T, *SyncPayload, uuid.UUID)
		name        string
		userID      uuid.UUID
		expectedNil bool
	}{
		{
			name:        "nil payload returns nil",
			payload:     nil,
			userID:      testUserID,
			expectedNil: true,
		},
		{
			name: "empty payload converts correctly",
			payload: &SyncPayload{
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			userID: testUserID,
			checkResult: func(t *testing.T, payload *SyncPayload, userID uuid.UUID) {
				t.Helper()
				result := payload.ToApp(userID)
				assert.NotNil(t, result)
				assert.Equal(t, userID, result.UserID)
				assert.Empty(t, result.BankCards)
				assert.Empty(t, result.Credentials)
				assert.Empty(t, result.Notes)
				assert.Empty(t, result.Files)
			},
		},
		{
			name: "payload with data converts correctly",
			payload: &SyncPayload{
				BankCards: []*bankcard.BankCard{
					{
						ID:         uuid.New(),
						CardNumber: "4111111111111111",
						CardHolder: "John Doe",
					},
				},
				Credentials: []*credential.Credential{
					{
						ID:       uuid.New(),
						Login:    "test@example.com",
						Password: "password123",
					},
				},
				Notes: []*note.Note{
					{
						ID:   uuid.New(),
						Note: "Test note",
					},
				},
				Files: []*filedata.FileData{
					{
						ID:         uuid.New(),
						StorageKey: "test.txt",
						Data:       []byte("test data"),
					},
				},
			},
			userID: testUserID,
			checkResult: func(t *testing.T, payload *SyncPayload, userID uuid.UUID) {
				t.Helper()
				result := payload.ToApp(userID)
				assert.NotNil(t, result)
				assert.Equal(t, userID, result.UserID)
				assert.Len(t, result.BankCards, 1)
				assert.Len(t, result.Credentials, 1)
				assert.Len(t, result.Notes, 1)
				assert.Len(t, result.Files, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectedNil {
				result := tt.payload.ToApp(tt.userID)
				assert.Nil(t, result)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, tt.payload, tt.userID)
			}
		})
	}
}

func TestNewSyncPayloadFromApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		payload     interface{}
		name        string
		expectedNil bool
	}{
		{
			name:        "nil payload returns nil",
			payload:     nil,
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectedNil {
				result := NewSyncPayloadFromApp(nil)
				assert.Nil(t, result)
			}
		})
	}
}

func TestSyncPayload_isEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		payload  *SyncPayload
		name     string
		expected bool
	}{
		{
			name: "empty payload is empty",
			payload: &SyncPayload{
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			expected: true,
		},
		{
			name: "nil slices are empty",
			payload: &SyncPayload{
				BankCards:   nil,
				Credentials: nil,
				Notes:       nil,
				Files:       nil,
			},
			expected: true,
		},
		{
			name: "payload with bank cards is not empty",
			payload: &SyncPayload{
				BankCards: []*bankcard.BankCard{
					{ID: uuid.New(), CardNumber: "4111111111111111"},
				},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			expected: false,
		},
		{
			name: "payload with credentials is not empty",
			payload: &SyncPayload{
				BankCards: []*bankcard.BankCard{},
				Credentials: []*credential.Credential{
					{ID: uuid.New(), Login: "test@example.com"},
				},
				Notes: []*note.Note{},
				Files: []*filedata.FileData{},
			},
			expected: false,
		},
		{
			name: "payload with notes is not empty",
			payload: &SyncPayload{
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes: []*note.Note{
					{ID: uuid.New(), Note: "Test note"},
				},
				Files: []*filedata.FileData{},
			},
			expected: false,
		},
		{
			name: "payload with files is not empty",
			payload: &SyncPayload{
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files: []*filedata.FileData{
					{ID: uuid.New(), StorageKey: "test.txt"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.payload.isEmpty()

			assert.Equal(t, tt.expected, result)
		})
	}
}
