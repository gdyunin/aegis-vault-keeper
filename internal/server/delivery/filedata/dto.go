package filedata

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/google/uuid"
)

// PullRequest represents the request to retrieve a specific file.
type PullRequest struct {
	// ID is the unique file identifier in UUID format.
	ID string `uri:"id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"` // File ID (required)
}

// ListResponse represents the response containing all user's files metadata.
type ListResponse struct {
	// Files contains metadata for all files belonging to the user.
	Files []*FileData `json:"files"` // List of files metadata
}

// PushRequest represents the data required to upload a file.
type PushRequest struct {
	// StorageKey is the custom filename or key for storing the file.
	StorageKey string `form:"storage_key" example:"document.pdf"` // Custom storage key (filename)
	// Description is optional user-provided description of the file content.
	Description string `form:"description" example:"Important PDF"` // File description
}

// PushResponse represents the response after uploading a file.
type PushResponse struct {
	// ID is the unique identifier assigned to the uploaded file.
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // Uploaded file ID
}

// FileData represents a file entity with metadata.
type FileData struct {
	// UpdatedAt indicates when the file was last modified.
	UpdatedAt time.Time `json:"updated_at"     example:"2023-12-01T10:00:00Z"`
	// StorageKey is the filename or key used for storing the file.
	StorageKey string `json:"storage_key"    example:"document.pdf"`
	// HashSum is the MD5 hash of the file content for integrity verification.
	HashSum string `json:"hash_sum"       example:"d41d8cd98f00b204e9800998ecf8427e"`
	// Description is the user-provided description of the file content.
	Description string `json:"description"    example:"Important PDF document"`
	// Data contains the file content bytes (omitted in list responses).
	Data []byte `json:"data,omitempty"`
	// ID is the unique file identifier.
	ID uuid.UUID `json:"id"             example:"123e4567-e89b-12d3-a456-426614174000"`
	// UserID identifies the file owner.
	UserID uuid.UUID `json:"user_id"        example:"987fcdeb-51a2-43d1-9f12-ba9876543210"`
}

// NewFileDataFromApp converts an application filedata entity to delivery DTO format.
func NewFileDataFromApp(fd *filedata.FileData) *FileData {
	return &FileData{
		ID:          fd.ID,
		UserID:      fd.UserID,
		StorageKey:  fd.StorageKey,
		HashSum:     fd.HashSum,
		Description: fd.Description,
		UpdatedAt:   fd.UpdatedAt,
		Data:        fd.Data,
	}
}

// NewFileDataListFromApp converts a slice of application filedata entities to delivery DTO format.
func NewFileDataListFromApp(files []*filedata.FileData) []*FileData {
	result := make([]*FileData, 0, len(files))
	for _, fd := range files {
		result = append(result, NewFileDataFromApp(fd))
	}
	return result
}

// ToApp converts the delivery FileData DTO to application layer entity format.
func (f *FileData) ToApp(userID uuid.UUID) *filedata.FileData {
	if f == nil {
		return nil
	}
	return &filedata.FileData{
		ID:          f.ID,
		UserID:      userID,
		StorageKey:  f.StorageKey,
		HashSum:     f.HashSum,
		Description: f.Description,
		UpdatedAt:   f.UpdatedAt,
		Data:        f.Data,
	}
}

// FilesToApp converts a slice of delivery file DTOs to application entities.
func FilesToApp(files []*FileData, userID uuid.UUID) []*filedata.FileData {
	if files == nil {
		return nil
	}
	result := make([]*filedata.FileData, 0, len(files))
	for _, f := range files {
		result = append(result, f.ToApp(userID))
	}
	return result
}

// withoutData creates a copy of FileData with the Data field excluded for list responses.
func (f *FileData) withoutData() *FileData {
	return &FileData{
		ID:          f.ID,
		UserID:      f.UserID,
		StorageKey:  f.StorageKey,
		HashSum:     f.HashSum,
		Description: f.Description,
		UpdatedAt:   f.UpdatedAt,
	}
}
