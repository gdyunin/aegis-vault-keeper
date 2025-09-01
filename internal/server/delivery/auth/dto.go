package auth

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest represents the data required for user registration.
type RegisterRequest struct {
	// Login contains the user's email address or username (required, unique across system).
	Login string `json:"login"    binding:"required" example:"user@example.com"`
	// Password contains the user's plaintext password (required, min 8 chars, will be hashed).
	Password string `json:"password" binding:"required" example:"securePassword123"`
}

// LoginRequest represents the data required for user authentication.
type LoginRequest struct {
	// Login contains the user's email address or username (required, must exist in system).
	Login string `json:"login"    binding:"required" example:"user@example.com"`
	// Password contains the user's plaintext password (required, verified against stored hash).
	Password string `json:"password" binding:"required" example:"securePassword123"`
}

// RegisterResponse represents the response after successful user registration.
type RegisterResponse struct {
	// ID contains the newly created user's unique identifier.
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// AccessToken represents the authentication token and its metadata.
type AccessToken struct {
	// AccessToken contains the JWT token for authenticating subsequent requests.
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// ExpiresAt specifies when the token becomes invalid and must be refreshed.
	ExpiresAt time.Time `json:"expires_at"   example:"2023-12-31T23:59:59Z"`
	// TokenType specifies the token type, always "Bearer" for OAuth 2.0 compliance.
	TokenType string `json:"token_type"   example:"Bearer"`
}
