package credential

import "errors"

// ErrNewCredentialParamsValidation indicates validation failure during credential creation.
var ErrNewCredentialParamsValidation = errors.New("invalid parameters for new credential")

// ErrIncorrectLogin indicates the login field is empty or invalid.
var ErrIncorrectLogin = errors.New("incorrect login")

// ErrIncorrectPassword indicates the password field is empty or invalid.
var ErrIncorrectPassword = errors.New("incorrect password")
