package auth

import (
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputErr error
		wantErr  error
		name     string
	}{
		{
			name:     "nil_error",
			inputErr: nil,
			wantErr:  nil,
		},
		{
			name:     "domain_validation_error",
			inputErr: auth.ErrNewUserParamsValidation,
			wantErr:  ErrAuthAppError,
		},
		{
			name:     "domain_incorrect_login",
			inputErr: auth.ErrIncorrectLogin,
			wantErr:  ErrAuthIncorrectLogin,
		},
		{
			name:     "domain_incorrect_password",
			inputErr: auth.ErrIncorrectPassword,
			wantErr:  ErrAuthIncorrectPassword,
		},
		{
			name:     "domain_password_verification_failed",
			inputErr: auth.ErrPasswordVerificationFailed,
			wantErr:  ErrAuthWrongLoginOrPassword,
		},
		{
			name:     "repository_user_not_found",
			inputErr: repository.ErrUserNotFound,
			wantErr:  ErrAuthWrongLoginOrPassword,
		},
		{
			name:     "repository_user_already_exists",
			inputErr: repository.ErrUserAlreadyExists,
			wantErr:  ErrAuthUserAlreadyExists,
		},
		{
			name:     "invalid_access_token",
			inputErr: ErrAuthInvalidAccessToken,
			wantErr:  ErrAuthInvalidAccessToken,
		},
		{
			name:     "unknown_error",
			inputErr: errors.New("unknown error"),
			wantErr:  ErrAuthTechError, // mapError wraps this with "error after mapping"
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapError(tt.inputErr)

			if tt.wantErr == nil {
				assert.Nil(t, result)
			} else {
				require.Error(t, result)
				assert.Contains(t, result.Error(), "error after mapping")
				assert.ErrorIs(t, result, tt.wantErr)
			}
		})
	}
}

func TestMapFn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputErr error
		wantErr  error
		name     string
	}{
		{
			name:     "domain_validation_error",
			inputErr: auth.ErrNewUserParamsValidation,
			wantErr:  ErrAuthAppError,
		},
		{
			name:     "domain_incorrect_login",
			inputErr: auth.ErrIncorrectLogin,
			wantErr:  ErrAuthIncorrectLogin,
		},
		{
			name:     "domain_incorrect_password",
			inputErr: auth.ErrIncorrectPassword,
			wantErr:  ErrAuthIncorrectPassword,
		},
		{
			name:     "domain_password_verification_failed",
			inputErr: auth.ErrPasswordVerificationFailed,
			wantErr:  ErrAuthWrongLoginOrPassword,
		},
		{
			name:     "repository_user_not_found",
			inputErr: repository.ErrUserNotFound,
			wantErr:  ErrAuthWrongLoginOrPassword,
		},
		{
			name:     "repository_user_already_exists",
			inputErr: repository.ErrUserAlreadyExists,
			wantErr:  ErrAuthUserAlreadyExists,
		},
		{
			name:     "invalid_access_token",
			inputErr: ErrAuthInvalidAccessToken,
			wantErr:  ErrAuthInvalidAccessToken,
		},
		{
			name:     "wrapped_domain_error",
			inputErr: errors.Join(errors.New("wrapper"), auth.ErrIncorrectLogin),
			wantErr:  ErrAuthIncorrectLogin,
		},
		{
			name:     "unknown_error",
			inputErr: errors.New("unknown error"),
			wantErr:  ErrAuthTechError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapFn(tt.inputErr)

			require.NotNil(t, result)
			assert.ErrorIs(t, result, tt.wantErr)

			// For unknown errors, check that it's joined with ErrAuthTechError
			if tt.name == "unknown_error" {
				assert.ErrorIs(t, result, tt.inputErr)
			}
		})
	}
}
