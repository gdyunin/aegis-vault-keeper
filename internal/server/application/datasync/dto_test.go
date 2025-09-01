package datasync

import (
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncPayload(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		name            string
		payload         SyncPayload
		wantUserID      uuid.UUID
		wantBankCards   int
		wantCredentials int
		wantNotes       int
		wantFiles       int
	}{
		{
			name: "success/populated_payload",
			payload: SyncPayload{
				UserID: testUserID,
				BankCards: []*bankcard.BankCard{
					{ID: uuid.New(), UserID: testUserID},
					{ID: uuid.New(), UserID: testUserID},
				},
				Credentials: []*credential.Credential{
					{ID: uuid.New(), UserID: testUserID},
				},
				Notes: []*note.Note{
					{ID: uuid.New(), UserID: testUserID},
					{ID: uuid.New(), UserID: testUserID},
					{ID: uuid.New(), UserID: testUserID},
				},
				Files: []*filedata.FileData{
					{ID: uuid.New(), UserID: testUserID},
				},
			},
			wantUserID:      testUserID,
			wantBankCards:   2,
			wantCredentials: 1,
			wantNotes:       3,
			wantFiles:       1,
		},
		{
			name: "success/empty_payload",
			payload: SyncPayload{
				UserID:      testUserID,
				BankCards:   []*bankcard.BankCard{},
				Credentials: []*credential.Credential{},
				Notes:       []*note.Note{},
				Files:       []*filedata.FileData{},
			},
			wantUserID:      testUserID,
			wantBankCards:   0,
			wantCredentials: 0,
			wantNotes:       0,
			wantFiles:       0,
		},
		{
			name: "success/nil_slices",
			payload: SyncPayload{
				UserID:      testUserID,
				BankCards:   nil,
				Credentials: nil,
				Notes:       nil,
				Files:       nil,
			},
			wantUserID:      testUserID,
			wantBankCards:   0,
			wantCredentials: 0,
			wantNotes:       0,
			wantFiles:       0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantUserID, tt.payload.UserID)
			assert.Len(t, tt.payload.BankCards, tt.wantBankCards)
			assert.Len(t, tt.payload.Credentials, tt.wantCredentials)
			assert.Len(t, tt.payload.Notes, tt.wantNotes)
			assert.Len(t, tt.payload.Files, tt.wantFiles)
		})
	}
}

func TestSyncPayload_StructFields(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testBankCardID := uuid.New()
	testCredentialID := uuid.New()
	testNoteID := uuid.New()
	testFileID := uuid.New()

	// Test that all fields can be set and accessed
	payload := SyncPayload{
		UserID: testUserID,
		BankCards: []*bankcard.BankCard{
			{ID: testBankCardID, UserID: testUserID},
		},
		Credentials: []*credential.Credential{
			{ID: testCredentialID, UserID: testUserID},
		},
		Notes: []*note.Note{
			{ID: testNoteID, UserID: testUserID},
		},
		Files: []*filedata.FileData{
			{ID: testFileID, UserID: testUserID},
		},
	}

	// Verify all fields are accessible
	require.Equal(t, testUserID, payload.UserID)
	require.Len(t, payload.BankCards, 1)
	require.Len(t, payload.Credentials, 1)
	require.Len(t, payload.Notes, 1)
	require.Len(t, payload.Files, 1)

	// Verify nested IDs
	assert.Equal(t, testBankCardID, payload.BankCards[0].ID)
	assert.Equal(t, testCredentialID, payload.Credentials[0].ID)
	assert.Equal(t, testNoteID, payload.Notes[0].ID)
	assert.Equal(t, testFileID, payload.Files[0].ID)
}
