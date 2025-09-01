package credential

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Credential represents a user credential entity with encrypted storage for sensitive data.
type Credential struct {
	// UpdatedAt contains the last modification timestamp.
	UpdatedAt time.Time
	// Login contains the encrypted username/login (required, 1-255 chars).
	Login []byte
	// Password contains the encrypted password (required, 1-255 chars).
	Password []byte
	// Description contains the encrypted user-provided description (optional, max 255 chars).
	Description []byte
	// ID contains the unique credential identifier.
	ID uuid.UUID
	// UserID contains the credential owner identifier.
	UserID uuid.UUID
}

// NewCredential creates a new credential entity with validation and encryption of sensitive data.
func NewCredential(params NewCredentialParams) (*Credential, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.Join(ErrNewCredentialParamsValidation, err)
	}

	c := Credential{
		ID:          uuid.New(),
		UserID:      params.UserID,
		Login:       []byte(params.Login),
		Password:    []byte(params.Password),
		Description: []byte(params.Description),
		UpdatedAt:   time.Now(),
	}

	return &c, nil
}

// NewCredentialParams contains the parameters for creating a new credential entity.
type NewCredentialParams struct {
	// Login contains the username/login (required, 1-255 chars).
	Login string
	// Password contains the password (required, 1-255 chars).
	Password string
	// Description contains optional user-provided description (max 255 chars).
	Description string
	// UserID identifies the user creating this credential.
	UserID uuid.UUID
}

// Validate performs comprehensive validation of all credential parameters.
func (cp *NewCredentialParams) Validate() error {
	validations := []func() error{
		cp.validateLogin,
		cp.validatePassword,
	}

	// errs collects all validation errors encountered during credential validation.
	var errs []error
	for _, fn := range validations {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateLogin validates that the login field is not empty.
func (cp *NewCredentialParams) validateLogin() error {
	if cp.Login == "" {
		return ErrIncorrectLogin
	}
	return nil
}

// validatePassword validates that the password field is not empty.
func (cp *NewCredentialParams) validatePassword() error {
	if cp.Password == "" {
		return ErrIncorrectPassword
	}
	return nil
}
