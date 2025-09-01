package credential

import (
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCredential_ToApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fixedTime := time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)
	credID := uuid.New()

	tests := []struct {
		cred     *Credential
		expected *credential.Credential
		name     string
		userID   uuid.UUID
	}{
		{
			name: "valid credential conversion",
			cred: &Credential{
				ID:          credID,
				Login:       "user@example.com",
				Password:    "securePassword123",
				Description: "Email account credentials",
				UpdatedAt:   fixedTime,
			},
			userID: userID,
			expected: &credential.Credential{
				ID:          credID,
				UserID:      userID,
				Login:       "user@example.com",
				Password:    "securePassword123",
				Description: "Email account credentials",
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:     "nil credential",
			cred:     nil,
			userID:   userID,
			expected: nil,
		},
		{
			name: "empty fields",
			cred: &Credential{
				ID:          uuid.Nil,
				Login:       "",
				Password:    "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
			userID: userID,
			expected: &credential.Credential{
				ID:          uuid.Nil,
				UserID:      userID,
				Login:       "",
				Password:    "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.cred.ToApp(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCredentialsToApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	credID1 := uuid.New()
	credID2 := uuid.New()

	tests := []struct {
		name     string
		creds    []*Credential
		expected []*credential.Credential
		userID   uuid.UUID
	}{
		{
			name: "multiple credentials",
			creds: []*Credential{
				{
					ID:       credID1,
					Login:    "user1@example.com",
					Password: "password1",
				},
				{
					ID:       credID2,
					Login:    "user2@example.com",
					Password: "password2",
				},
			},
			userID: userID,
			expected: []*credential.Credential{
				{
					ID:       credID1,
					UserID:   userID,
					Login:    "user1@example.com",
					Password: "password1",
				},
				{
					ID:       credID2,
					UserID:   userID,
					Login:    "user2@example.com",
					Password: "password2",
				},
			},
		},
		{
			name:     "nil slice",
			creds:    nil,
			userID:   userID,
			expected: nil,
		},
		{
			name:     "empty slice",
			creds:    []*Credential{},
			userID:   userID,
			expected: []*credential.Credential{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CredentialsToApp(tt.creds, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCredentialFromApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	credID := uuid.New()
	fixedTime := time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		appCred  *credential.Credential
		expected *Credential
		name     string
	}{
		{
			name: "valid app credential conversion",
			appCred: &credential.Credential{
				ID:          credID,
				UserID:      userID,
				Login:       "user@example.com",
				Password:    "securePassword123",
				Description: "Email account credentials",
				UpdatedAt:   fixedTime,
			},
			expected: &Credential{
				ID:          credID,
				Login:       "user@example.com",
				Password:    "securePassword123",
				Description: "Email account credentials",
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:     "nil app credential",
			appCred:  nil,
			expected: nil,
		},
		{
			name: "empty fields",
			appCred: &credential.Credential{
				ID:          uuid.Nil,
				UserID:      userID,
				Login:       "",
				Password:    "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
			expected: &Credential{
				ID:          uuid.Nil,
				Login:       "",
				Password:    "",
				Description: "",
				UpdatedAt:   time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewCredentialFromApp(tt.appCred)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCredentialsFromApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	credID1 := uuid.New()
	credID2 := uuid.New()

	tests := []struct {
		name     string
		appCreds []*credential.Credential
		expected []*Credential
	}{
		{
			name: "multiple app credentials",
			appCreds: []*credential.Credential{
				{
					ID:       credID1,
					UserID:   userID,
					Login:    "user1@example.com",
					Password: "password1",
				},
				{
					ID:       credID2,
					UserID:   userID,
					Login:    "user2@example.com",
					Password: "password2",
				},
			},
			expected: []*Credential{
				{
					ID:       credID1,
					Login:    "user1@example.com",
					Password: "password1",
				},
				{
					ID:       credID2,
					Login:    "user2@example.com",
					Password: "password2",
				},
			},
		},
		{
			name:     "nil slice",
			appCreds: nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			appCreds: []*credential.Credential{},
			expected: []*Credential{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewCredentialsFromApp(tt.appCreds)
			assert.Equal(t, tt.expected, result)
		})
	}
}
