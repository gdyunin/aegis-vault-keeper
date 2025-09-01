package filedata

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/google/uuid"
)

// FileData represents file metadata and content for application layer operations.
type FileData struct {
	// UpdatedAt contains the last modification timestamp.
	UpdatedAt time.Time
	// StorageKey contains the unique file identifier in storage (path-safe).
	StorageKey string
	// HashSum contains the SHA256 hash of file content for integrity verification.
	HashSum string
	// Description contains user-provided file description (max 255 chars).
	Description string
	// Data contains the actual file content bytes (may be empty for metadata-only operations).
	Data []byte
	// ID contains the unique file identifier.
	ID uuid.UUID
	// UserID contains the file owner identifier.
	UserID uuid.UUID
}

// newFileFromDomain converts a domain FileData entity to application layer DTO.
func newFileFromDomain(c *filedata.FileData) *FileData {
	if c == nil {
		return nil
	}
	return &FileData{
		ID:          c.ID,
		UserID:      c.UserID,
		StorageKey:  string(c.StorageKey),
		HashSum:     string(c.HashSum),
		Description: string(c.Description),
		UpdatedAt:   c.UpdatedAt,
	}
}

// newFilesFromDomain converts a slice of domain FileData entities to application layer DTOs.
func newFilesFromDomain(cs []*filedata.FileData) []*FileData {
	result := make([]*FileData, 0, len(cs))
	for _, c := range cs {
		result = append(result, newFileFromDomain(c))
	}
	return result
}

// PullParams contains parameters for retrieving a specific file by ID.
type PullParams struct {
	// ID specifies the file to retrieve.
	ID uuid.UUID
	// UserID specifies the file owner for access control.
	UserID uuid.UUID
}

// ListParams contains parameters for retrieving all files belonging to a user.
type ListParams struct {
	// UserID specifies the file owner for filtering.
	UserID uuid.UUID
}

// PushParams contains parameters for creating or updating file data.
type PushParams struct {
	// StorageKey specifies the file storage identifier (path-safe string).
	StorageKey string
	// Description contains user-provided file description (max 255 chars).
	Description string
	// Data contains the file content bytes (required for new files).
	Data []byte
	// ID specifies the file ID for updates (uuid.Nil for new files).
	ID uuid.UUID
	// UserID specifies the file owner.
	UserID uuid.UUID
}

// calculateDataHashSum computes the SHA256 hash of the file data for integrity verification.
func (p *PushParams) calculateDataHashSum() string {
	hash := sha256.Sum256(p.Data)
	return hex.EncodeToString(hash[:])
}
