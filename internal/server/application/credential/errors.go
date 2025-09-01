package credential

import (
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/errutil"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
)

// Credential error definitions.
var (
	// ErrCredentialAppError indicates a general credential application error.
	ErrCredentialAppError = errors.New("credential application error")

	// ErrCredentialTechError indicates a technical error in the credential system.
	ErrCredentialTechError = errors.New("credential technical error")

	// ErrCredentialIncorrectLogin indicates an incorrect login was provided.
	ErrCredentialIncorrectLogin = errors.New("incorrect login")

	// ErrCredentialIncorrectPassword indicates an incorrect password was provided.
	ErrCredentialIncorrectPassword = errors.New("incorrect password")

	// ErrCredentialNotFound indicates the requested credential was not found.
	ErrCredentialNotFound = errors.New("credential not found")

	// ErrCredentialAccessDenied indicates access to the credential is not permitted.
	ErrCredentialAccessDenied = errors.New("access to this credential is denied")
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
	case errors.Is(err, credential.ErrNewCredentialParamsValidation):
		return ErrCredentialAppError
	case errors.Is(err, credential.ErrIncorrectLogin):
		return ErrCredentialIncorrectLogin
	case errors.Is(err, credential.ErrIncorrectPassword):
		return ErrCredentialIncorrectPassword
	default:
		return errors.Join(ErrCredentialTechError, err)
	}
}
