package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
	"github.com/google/uuid"
)

// TokenGenerateValidator defines the interface for JWT token generation and validation operations.
type TokenGenerateValidator interface {
	// GenerateAccessToken creates a new JWT access token for the specified user ID.
	GenerateAccessToken(userID uuid.UUID) (token string, tokenType string, expiresAt time.Time, err error)

	// ValidateAccessToken validates a JWT token string and returns the associated user ID.
	ValidateAccessToken(tokenString string) (uuid.UUID, error)
}

// CryptoKeyGenerator is an alias for auth.CryptoKeyGenerator.
type CryptoKeyGenerator auth.CryptoKeyGenerator

// PasswordHasherVerificator combines password hashing and verification functionality.
type PasswordHasherVerificator interface {
	auth.PasswordHasher
	auth.PasswordVerificator
}

// Repository defines the interface for user data persistence operations.
type Repository interface {
	// Save persists user data using the provided parameters.
	Save(ctx context.Context, params repository.SaveParams) error

	// Load retrieves user data using the provided parameters.
	Load(ctx context.Context, params repository.LoadParams) (*auth.User, error)
}

// Service provides authentication business logic operations.
type Service struct {
	// r is the repository interface for user data persistence operations.
	r Repository
	// passwordHasherVerificator handles password hashing and verification operations.
	passwordHasherVerificator PasswordHasherVerificator
	// cryptoKeyGenerator generates cryptographic keys for user data encryption.
	cryptoKeyGenerator CryptoKeyGenerator
	// tokenGenerateValidator handles JWT token generation and validation operations.
	tokenGenerateValidator TokenGenerateValidator
}

// NewService creates a new authentication service instance with the provided dependencies.
func NewService(
	r Repository,
	passwordHasherVerificator PasswordHasherVerificator,
	cryptoKeyGenerator CryptoKeyGenerator,
	tokenGenerator TokenGenerateValidator,
) *Service {
	return &Service{
		r:                         r,
		passwordHasherVerificator: passwordHasherVerificator,
		cryptoKeyGenerator:        cryptoKeyGenerator,
		tokenGenerateValidator:    tokenGenerator,
	}
}

// Register creates a new user account with the provided registration parameters.
func (s *Service) Register(ctx context.Context, params RegisterParams) (uuid.UUID, error) {
	u, err := auth.NewUser(
		auth.NewUserParams{Login: params.Login, Password: params.Password},
		s.passwordHasherVerificator,
		s.cryptoKeyGenerator,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create new user: %w", mapError(err))
	}

	if err := s.r.Save(ctx, repository.SaveParams{Entity: u}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save user: %w", mapError(err))
	}

	return u.ID, nil
}

// Login authenticates a user with the provided credentials and returns an access token.
func (s *Service) Login(ctx context.Context, params LoginParams) (AccessToken, error) {
	u, err := s.r.Load(ctx, repository.LoadParams{Login: params.Login})
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to load user: %w", mapError(err))
	}

	ok, err := u.VerifyPassword(s.passwordHasherVerificator, params.Password)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to verify password: %w", mapError(err))
	}
	if !ok {
		return AccessToken{}, fmt.Errorf("authentication failed: %w", ErrAuthWrongLoginOrPassword)
	}

	token, tokType, expiresAt, err := s.tokenGenerateValidator.GenerateAccessToken(u.ID)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to generate access token: %w", mapError(err))
	}

	return AccessToken{AccessToken: token, TokenType: tokType, ExpiresAt: expiresAt}, nil
}

// ValidateToken validates an access token and returns the associated user ID.
func (s *Service) ValidateToken(tokenString string) (uuid.UUID, error) {
	userID, err := s.tokenGenerateValidator.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to validate access token: %w", ErrAuthInvalidAccessToken)
	}
	return userID, nil
}
