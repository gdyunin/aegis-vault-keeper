package credential

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredential(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	type args struct {
		params NewCredentialParams
	}
	tests := []struct {
		want    func(t *testing.T, c *Credential)
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid/complete_credential",
			args: args{
				params: NewCredentialParams{
					Login:       "john.doe@example.com",
					Password:    "SecurePassword123!",
					Description: "Personal email account",
					UserID:      userID,
				},
			},
			want: func(t *testing.T, c *Credential) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, c.ID)
				assert.Equal(t, userID, c.UserID)
				assert.Equal(t, []byte("john.doe@example.com"), c.Login)
				assert.Equal(t, []byte("SecurePassword123!"), c.Password)
				assert.Equal(t, []byte("Personal email account"), c.Description)
				assert.WithinDuration(t, time.Now(), c.UpdatedAt, time.Second)
			},
		},
		{
			name: "valid/minimal_credential",
			args: args{
				params: NewCredentialParams{
					Login:       "admin",
					Password:    "pass",
					Description: "",
					UserID:      userID,
				},
			},
			want: func(t *testing.T, c *Credential) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, c.ID)
				assert.Equal(t, userID, c.UserID)
				assert.Equal(t, []byte("admin"), c.Login)
				assert.Equal(t, []byte("pass"), c.Password)
				assert.Equal(t, []byte(""), c.Description)
				assert.WithinDuration(t, time.Now(), c.UpdatedAt, time.Second)
			},
		},
		{
			name: "valid/long_values",
			args: args{
				params: NewCredentialParams{
					Login:       "very.long.email.address@very.long.domain.example.com",
					Password:    "VeryLongPasswordWith123Numbers!@#AndSpecialCharacters",
					Description: "This is a very long description that contains multiple words and should still be valid",
					UserID:      userID,
				},
			},
			want: func(t *testing.T, c *Credential) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, c.ID)
				assert.Equal(t, userID, c.UserID)
				assert.Equal(t, []byte("very.long.email.address@very.long.domain.example.com"), c.Login)
				assert.Equal(t, []byte("VeryLongPasswordWith123Numbers!@#AndSpecialCharacters"), c.Password)
				assert.Equal(t, []byte("This is a very long description that contains multiple words "+
					"and should still be valid"), c.Description)
			},
		},
		{
			name: "invalid/empty_login",
			args: args{
				params: NewCredentialParams{
					Login:       "",
					Password:    "password123",
					Description: "Test account",
					UserID:      userID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid/empty_password",
			args: args{
				params: NewCredentialParams{
					Login:       "user@example.com",
					Password:    "",
					Description: "Test account",
					UserID:      userID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid/both_login_and_password_empty",
			args: args{
				params: NewCredentialParams{
					Login:       "",
					Password:    "",
					Description: "Test account",
					UserID:      userID,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewCredential(tt.args.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestNewCredentialParams_Validate(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		errType error
		name    string
		params  NewCredentialParams
		wantErr bool
	}{
		{
			name: "valid/standard_credentials",
			params: NewCredentialParams{
				Login:       "user@example.com",
				Password:    "password123",
				Description: "Test account",
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "valid/minimal_credentials",
			params: NewCredentialParams{
				Login:       "a",
				Password:    "p",
				Description: "",
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "valid/special_characters",
			params: NewCredentialParams{
				Login:       "user+tag@example.co.uk",
				Password:    "P@ssw0rd!#$%",
				Description: "Account with special chars",
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "valid/unicode_characters",
			params: NewCredentialParams{
				Login:       "пользователь@пример.рф",
				Password:    "пароль123",
				Description: "Описание на русском",
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "invalid/empty_login",
			params: NewCredentialParams{
				Login:       "",
				Password:    "password123",
				Description: "Test",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectLogin,
		},
		{
			name: "invalid/empty_password",
			params: NewCredentialParams{
				Login:       "user@example.com",
				Password:    "",
				Description: "Test",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrIncorrectPassword,
		},
		{
			name: "invalid/both_empty",
			params: NewCredentialParams{
				Login:       "",
				Password:    "",
				Description: "Test",
				UserID:      userID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
