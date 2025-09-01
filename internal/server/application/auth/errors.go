package auth

import (
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/errutil"
	domain "github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
)

// Authentication error definitions.
var (
	// ErrAuthAppError indicates a general authentication application error.
	ErrAuthAppError = errors.New("authentication application error")

	// ErrAuthTechError indicates a technical error in the authentication system.
	ErrAuthTechError = errors.New("authentication technical error")

	// ErrAuthIncorrectLogin indicates an incorrect login was provided.
	ErrAuthIncorrectLogin = errors.New("incorrect login")

	// ErrAuthIncorrectPassword indicates an incorrect password was provided.
	ErrAuthIncorrectPassword = errors.New("incorrect password")

	// ErrAuthWrongLoginOrPassword indicates invalid login credentials.
	ErrAuthWrongLoginOrPassword = errors.New("wrong login or password")

	// ErrAuthInvalidAccessToken indicates an invalid or expired access token.
	ErrAuthInvalidAccessToken = errors.New("invalid access token")

	// ErrAuthUserAlreadyExists indicates a user already exists with the given login.
	ErrAuthUserAlreadyExists = errors.New("user already exists")
)

// mapError maps domain and repository errors to application-level errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	mapped := errutil.MapError(mapFn, err)
	if mapped != nil {
		return fmt.Errorf("error after mapping: %w", mapped)
	}
	return nil
}

// mapFn provides the actual error mapping logic for different error types.
func mapFn(err error) error {
	switch {
	case errors.Is(err, domain.ErrNewUserParamsValidation):

		return ErrAuthAppError

	case errors.Is(err, domain.ErrIncorrectLogin):
		return ErrAuthIncorrectLogin

	case errors.Is(err, domain.ErrIncorrectPassword):
		return ErrAuthIncorrectPassword

	case errors.Is(err, domain.ErrPasswordVerificationFailed):
		return ErrAuthWrongLoginOrPassword

	case errors.Is(err, repository.ErrUserNotFound):
		return ErrAuthWrongLoginOrPassword

	case errors.Is(err, repository.ErrUserAlreadyExists):
		return ErrAuthUserAlreadyExists

	case errors.Is(err, ErrAuthInvalidAccessToken):
		return ErrAuthInvalidAccessToken

	default:
		return errors.Join(ErrAuthTechError, err)
	}
}
