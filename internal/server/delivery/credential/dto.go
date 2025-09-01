package credential

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/google/uuid"
)

// Credential represents a login/password credential entity for API transfer.
type Credential struct {
	// UpdatedAt contains the timestamp when this credential was last modified.
	UpdatedAt time.Time `json:"updated_at,omitzero"  example:"2023-12-01T10:00:00Z"`
	// Login contains the username, email, or account identifier (sensitive data).
	Login string `json:"login,omitzero"       example:"user@example.com"`
	// Password contains the plaintext password (highly sensitive, transmitted encrypted).
	Password string `json:"password,omitzero"    example:"securePassword123"`
	// Description contains optional user notes about where this credential is used.
	Description string `json:"description,omitzero" example:"Email account credentials"`
	// ID contains the unique identifier for this credential record.
	ID uuid.UUID `json:"id,omitzero"          example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ToApp converts this DTO to an application layer Credential entity with the specified user ID.
func (c *Credential) ToApp(userID uuid.UUID) *credential.Credential {
	if c == nil {
		return nil
	}
	return &credential.Credential{
		ID:          c.ID,
		UserID:      userID,
		Login:       c.Login,
		Password:    c.Password,
		Description: c.Description,
		UpdatedAt:   c.UpdatedAt,
	}
}

// CredentialsToApp converts a slice of DTOs to application layer Credential entities with the specified user ID.
func CredentialsToApp(creds []*Credential, userID uuid.UUID) []*credential.Credential {
	if creds == nil {
		return nil
	}
	result := make([]*credential.Credential, 0, len(creds))
	for _, c := range creds {
		result = append(result, c.ToApp(userID))
	}
	return result
}

// NewCredentialFromApp creates a DTO from an application layer Credential entity.
func NewCredentialFromApp(c *credential.Credential) *Credential {
	if c == nil {
		return nil
	}
	return &Credential{
		ID:          c.ID,
		Login:       c.Login,
		Password:    c.Password,
		Description: c.Description,
		UpdatedAt:   c.UpdatedAt,
	}
}

// NewCredentialsFromApp converts a slice of application credential entities to delivery DTO format.
func NewCredentialsFromApp(creds []*credential.Credential) []*Credential {
	if creds == nil {
		return nil
	}
	result := make([]*Credential, 0, len(creds))
	for _, c := range creds {
		result = append(result, NewCredentialFromApp(c))
	}
	return result
}

// PushRequest represents the data required to create or update a credential.
type PushRequest struct {
	// Login username or email (required)
	Login string `json:"login"                binding:"required" example:"user@example.com"`
	// Password (required)
	Password string `json:"password"             binding:"required" example:"securePassword123"`
	// Optional description
	Description string `json:"description,omitzero"                    example:"Email account credentials"`
}

// PullRequest represents the request to retrieve a specific credential.
type PullRequest struct {
	// Credential ID (required)
	ID string `uri:"id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// PushResponse represents the response after creating or updating a credential.
type PushResponse struct {
	// Created or updated credential ID
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// PullResponse represents the response containing a specific credential.
type PullResponse struct {
	// Credential data
	Credential *Credential `json:"credential"`
}

// ListResponse represents the response containing all user's credentials.
type ListResponse struct {
	// List of credentials
	Credentials []*Credential `json:"credentials"`
}
