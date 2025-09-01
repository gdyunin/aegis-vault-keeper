package auth

import (
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		entity *auth.User
		name   string
	}{
		{
			name: "save params with valid user entity",
			entity: &auth.User{
				ID:           uuid.New(),
				Login:        "test@example.com",
				PasswordHash: "hashed-password",
				CryptoKey:    []byte("crypto-key"),
			},
		},
		{
			name:   "save params with nil entity",
			entity: nil,
		},
		{
			name:   "save params with empty user entity",
			entity: &auth.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := SaveParams{
				Entity: tt.entity,
			}

			assert.Equal(t, tt.entity, params.Entity)
		})
	}
}

func TestLoadParams(t *testing.T) {
	t.Parallel()

	testID := uuid.New()

	tests := []struct {
		name     string
		login    string
		id       uuid.UUID
		checkID  bool
		checkStr bool
	}{
		{
			name:     "load params with login",
			login:    "test@example.com",
			id:       uuid.Nil,
			checkStr: true,
		},
		{
			name:    "load params with ID",
			login:   "",
			id:      testID,
			checkID: true,
		},
		{
			name:     "load params with both login and ID",
			login:    "test@example.com",
			id:       testID,
			checkID:  true,
			checkStr: true,
		},
		{
			name:  "load params with neither login nor ID",
			login: "",
			id:    uuid.Nil,
		},
		{
			name:     "load params with empty string login",
			login:    "",
			id:       uuid.Nil,
			checkStr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := LoadParams{
				Login: tt.login,
				ID:    tt.id,
			}

			if tt.checkStr {
				assert.Equal(t, tt.login, params.Login)
				assert.NotEmpty(t, params.Login)
			}
			if tt.checkID {
				assert.Equal(t, tt.id, params.ID)
				assert.NotEqual(t, uuid.Nil, params.ID)
			}
			if !tt.checkStr && !tt.checkID {
				assert.Empty(t, params.Login)
				assert.Equal(t, uuid.Nil, params.ID)
			}
		})
	}
}

func TestLoadParamsValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		params      LoadParams
		hasLogin    bool
		hasID       bool
		hasEither   bool
	}{
		{
			name: "valid params with login only",
			params: LoadParams{
				Login: "test@example.com",
				ID:    uuid.Nil,
			},
			hasLogin:    true,
			hasID:       false,
			hasEither:   true,
			description: "should have login but not ID",
		},
		{
			name: "valid params with ID only",
			params: LoadParams{
				Login: "",
				ID:    uuid.New(),
			},
			hasLogin:    false,
			hasID:       true,
			hasEither:   true,
			description: "should have ID but not login",
		},
		{
			name: "valid params with both login and ID",
			params: LoadParams{
				Login: "test@example.com",
				ID:    uuid.New(),
			},
			hasLogin:    true,
			hasID:       true,
			hasEither:   true,
			description: "should have both login and ID",
		},
		{
			name: "invalid params with neither login nor ID",
			params: LoadParams{
				Login: "",
				ID:    uuid.Nil,
			},
			hasLogin:    false,
			hasID:       false,
			hasEither:   false,
			description: "should have neither login nor ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Check login presence
			if tt.hasLogin {
				assert.NotEmpty(t, tt.params.Login, "login should not be empty")
			} else {
				assert.Empty(t, tt.params.Login, "login should be empty")
			}

			// Check ID presence
			if tt.hasID {
				assert.NotEqual(t, uuid.Nil, tt.params.ID, "ID should not be nil")
			} else {
				assert.Equal(t, uuid.Nil, tt.params.ID, "ID should be nil")
			}

			// Check if at least one identifier is present
			hasAtLeastOne := tt.params.Login != "" || tt.params.ID != uuid.Nil
			assert.Equal(t, tt.hasEither, hasAtLeastOne, tt.description)
		})
	}
}
