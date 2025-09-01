package credential

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	"github.com/google/uuid"
)

// Credential represents a credential data transfer object for application layer communication.
type Credential struct {
	// UpdatedAt indicates when the credential was last modified.
	UpdatedAt time.Time
	// Login contains the credential login/username.
	Login string
	// Password contains the credential password.
	Password string
	// Description contains additional information about the credential.
	Description string
	// ID uniquely identifies the credential.
	ID uuid.UUID
	// UserID identifies the credential owner.
	UserID uuid.UUID
}

// newCredentialFromDomain converts a domain credential entity to application DTO.
func newCredentialFromDomain(c *credential.Credential) *Credential {
	if c == nil {
		return nil
	}
	return &Credential{
		ID:          c.ID,
		UserID:      c.UserID,
		Login:       string(c.Login),
		Password:    string(c.Password),
		Description: string(c.Description),
		UpdatedAt:   c.UpdatedAt,
	}
}

// newCredentialsFromDomain converts a slice of domain credential entities to application DTOs.
func newCredentialsFromDomain(cs []*credential.Credential) []*Credential {
	result := make([]*Credential, 0, len(cs))
	for _, c := range cs {
		result = append(result, newCredentialFromDomain(c))
	}
	return result
}

// PullParams contains parameters for retrieving a specific credential.
type PullParams struct {
	// ID specifies the credential to retrieve.
	ID uuid.UUID
	// UserID specifies the credential owner.
	UserID uuid.UUID
}

// ListParams contains parameters for listing user credentials.
type ListParams struct {
	// UserID specifies the credential owner.
	UserID uuid.UUID
}

// PushParams contains parameters for creating or updating a credential.
type PushParams struct {
	// Login specifies the credential login/username.
	Login string
	// Password specifies the credential password.
	Password string
	// Description provides additional information about the credential.
	Description string
	// ID uniquely identifies the credential.
	ID uuid.UUID
	// UserID identifies the credential owner.
	UserID uuid.UUID
}
