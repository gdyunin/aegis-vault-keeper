package filestorage

import "github.com/google/uuid"

// SaveParams contains parameters for saving file data to storage.
type SaveParams struct {
	// StorageKey identifies the file in storage.
	StorageKey string
	// Data contains the file content to be saved.
	Data []byte
	// UserID identifies the user who owns the file.
	UserID uuid.UUID
}

// LoadParams contains parameters for loading file data from storage.
type LoadParams struct {
	// StorageKey identifies the file in storage.
	StorageKey string
	// UserID identifies the user who owns the file.
	UserID uuid.UUID
}

// DeleteParams contains parameters for deleting file data from storage.
type DeleteParams struct {
	// StorageKey identifies the file in storage.
	StorageKey string
	// UserID identifies the user who owns the file.
	UserID uuid.UUID
}
