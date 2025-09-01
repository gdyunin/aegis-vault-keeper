package auth

import "errors"

var (
	// ErrUserNotFound indicates that the requested user was not found in the repository.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists indicates that a user with the given credentials already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
)
