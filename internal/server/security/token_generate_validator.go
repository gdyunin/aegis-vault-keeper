package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// TokenTypeBearer defines the Bearer token type for JWT authentication.
	TokenTypeBearer = "Bearer"
)

// Claims represents the JWT token claims including user identification.
type Claims struct {
	jwt.RegisteredClaims
	// UserID contains the unique identifier of the authenticated user.
	UserID uuid.UUID `json:"user_id"`
}

// TokenGenerateValidator provides JWT token generation and validation functionality.
type TokenGenerateValidator struct {
	// secretKey contains the HMAC secret for signing and validating tokens.
	secretKey []byte
	// accessTokenExpireDuration defines how long access tokens remain valid.
	accessTokenExpireDuration time.Duration
}

const (
	// MinSecretKeyLength is the minimum required length for JWT secret key.
	MinSecretKeyLength = 32
)

// NewTokenGenerateValidator creates a new JWT token generator/validator with security validation.
func NewTokenGenerateValidator(
	secretKey []byte,
	accessTokenExpireDuration time.Duration,
) (*TokenGenerateValidator, error) {
	if len(secretKey) < MinSecretKeyLength {
		return nil, fmt.Errorf(
			"JWT error: secret key is too short, minimum %d bytes required",
			MinSecretKeyLength,
		)
	}
	return &TokenGenerateValidator{
		secretKey:                 secretKey,
		accessTokenExpireDuration: accessTokenExpireDuration,
	}, nil
}

// GenerateAccessToken creates a new JWT access token for the specified user.
func (t *TokenGenerateValidator) GenerateAccessToken(userID uuid.UUID) (string, string, time.Time, error) {
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(t.accessTokenExpireDuration)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "aegis_vault_keeper",
		},
	}

	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := rawToken.SignedString(t.secretKey)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("JWT error: failed to sign token: %w", err)
	}

	return tokenString, TokenTypeBearer, expiresAt, nil
}

// ValidateAccessToken validates a JWT token and returns the associated user ID.
func (t *TokenGenerateValidator) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("JWT error: unexpected signing method: %v", token.Header["alg"])
		}
		return t.secretKey, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("JWT error: invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims == nil {
		return uuid.Nil, errors.New("JWT error: token is not valid or has expired")
	}

	return claims.UserID, nil
}
