package auth

import "errors"

// User domain error definitions.
var (
	// ErrCryptoKeyGenerate indicates failure to generate user cryptographic key.
	ErrCryptoKeyGenerate = errors.New("failed to generate user cryptographic key")

	// ErrPasswordHash indicates failure to hash password.
	ErrPasswordHash = errors.New("failed to hash password")

	// ErrNewUserParamsValidation indicates new user parameters validation failed.
	ErrNewUserParamsValidation = errors.New("new user parameters validation failed")

	// ErrIncorrectLogin indicates the login format or length is incorrect.
	ErrIncorrectLogin = errors.New("incorrect login")

	// ErrIncorrectPassword indicates the password format or length is incorrect.
	ErrIncorrectPassword = errors.New("incorrect password")

	// ErrPasswordVerificationFailed indicates password verification failed.
	ErrPasswordVerificationFailed = errors.New("password verification failed")
)
