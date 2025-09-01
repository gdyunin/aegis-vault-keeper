package auth

import "time"

// RegisterParams contains the parameters required for user registration.
type RegisterParams struct {
	// Login specifies the username for the new user account.
	Login string
	// Password specifies the password for the new user account.
	Password string
}

// LoginParams contains the parameters required for user authentication.
type LoginParams struct {
	// Login specifies the username for authentication.
	Login string
	// Password specifies the password for authentication.
	Password string
}

// AccessToken represents a JWT access token with its metadata.
type AccessToken struct {
	// AccessToken contains the JWT token string.
	AccessToken string
	// ExpiresAt specifies when the token expires.
	ExpiresAt time.Time
	// TokenType specifies the type of token (typically "Bearer").
	TokenType string
}
