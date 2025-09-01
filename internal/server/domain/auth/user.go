package auth

import (
	"errors"

	"github.com/google/uuid"
)

const (
	// CryptoKeySize defines the size in bytes for user-specific encryption keys.
	cryptoKeySize = 32

	// LoginMinLen defines the minimum length for user login.
	loginMinLen = 5

	// LoginMaxLen defines the maximum length for user login.
	loginMaxLen = 50

	// PasswordMinLen defines the minimum length for user password.
	passwordMinLen = 8

	// PasswordMaxLen defines the maximum length for user password.
	passwordMaxLen = 64
)

type (
	// CryptoKeyGenerator defines the interface for generating user-specific encryption keys.
	CryptoKeyGenerator interface {
		// CryptoKeyGenerate generates a random encryption key of the specified size.
		CryptoKeyGenerate(size int) ([]byte, error)
	}

	// PasswordHasher defines the interface for password hashing operations.
	PasswordHasher interface {
		// PasswordHash generates a hash for the provided password.
		PasswordHash(password string) (string, error)
	}

	// PasswordVerificator defines the interface for password verification operations.
	PasswordVerificator interface {
		// PasswordVerify checks if the verifying data matches the hashed data.
		PasswordVerify(hashedData, verifyingData string) (bool, error)
	}
)

// User represents a user entity in the authentication domain.
type User struct {
	// Login contains the user's unique login identifier.
	Login string
	// PasswordHash contains the hashed password.
	PasswordHash string
	// CryptoKey contains the user-specific encryption key.
	CryptoKey []byte
	// ID is the unique identifier of the user.
	ID uuid.UUID
}

// NewUser creates a new user entity with the provided parameters and dependencies.
func NewUser(params NewUserParams, hasher PasswordHasher, cryptoKeyGen CryptoKeyGenerator) (*User, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.Join(ErrNewUserParamsValidation, err)
	}

	cryptoKey, err := cryptoKeyGen.CryptoKeyGenerate(cryptoKeySize)
	if err != nil {
		return nil, errors.Join(ErrCryptoKeyGenerate, err)
	}

	passwordHash, err := hasher.PasswordHash(params.Password)
	if err != nil {
		return nil, errors.Join(ErrPasswordHash, err)
	}

	u := User{
		ID:           uuid.New(),
		Login:        params.Login,
		PasswordHash: passwordHash,
		CryptoKey:    cryptoKey,
	}

	return &u, nil
}

// VerifyPassword verifies if the provided password matches the user's stored password.
func (u *User) VerifyPassword(verificator PasswordVerificator, password string) (bool, error) {
	verified, err := verificator.PasswordVerify(u.PasswordHash, password)
	if err != nil {
		return false, errors.Join(ErrPasswordVerificationFailed, err)
	}
	return verified, nil
}

// NewUserParams contains parameters for creating a new user.
type NewUserParams struct {
	// Login specifies the user's login identifier.
	Login string
	// Password specifies the user's password.
	Password string
}

// Validate validates the new user parameters and returns any validation errors.
func (up *NewUserParams) Validate() error {
	validations := []func() error{
		up.validateLogin,
		up.validatePassword,
	}

	// errs collects all validation errors encountered during user parameter validation.
	var errs []error
	for _, validation := range validations {
		if err := validation(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateLogin validates the login parameter length constraints.
func (up *NewUserParams) validateLogin() error {
	if len(up.Login) < loginMinLen || len(up.Login) > loginMaxLen {
		return ErrIncorrectLogin
	}
	return nil
}

// validatePassword validates the password parameter length constraints.
func (up *NewUserParams) validatePassword() error {
	if len(up.Password) < passwordMinLen || len(up.Password) > passwordMaxLen {
		return ErrIncorrectPassword
	}
	return nil
}
