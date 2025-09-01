package filedata

import (
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileData represents a file with metadata for secure storage.
type FileData struct {
	// UpdatedAt contains the timestamp when the file was last modified.
	UpdatedAt time.Time
	// Description contains the encrypted file description.
	Description []byte
	// StorageKey contains the encrypted storage path for the file.
	StorageKey []byte
	// HashSum contains the encrypted SHA256 hash of the file content.
	HashSum []byte
	// ID uniquely identifies this file.
	ID uuid.UUID
	// UserID identifies the user who owns this file.
	UserID uuid.UUID
}

// NewFileDataParams contains parameters for creating a new file data entry.
type NewFileDataParams struct {
	// Description contains an optional description for the file.
	Description string
	// StorageKey contains the storage path for the file (required, validated for security).
	StorageKey string
	// HashSum contains the SHA256 hash of the file content (required, validated as hex).
	HashSum string
	// UserID identifies the user who will own this file.
	UserID uuid.UUID
}

// NewFile creates a new file data entry with the provided parameters after validation.
func NewFile(p NewFileDataParams) (*FileData, error) {
	if err := p.Validate(); err != nil {
		return nil, errors.Join(ErrNewFileParamsValidation, err)
	}
	return &FileData{
		ID:          uuid.New(),
		UserID:      p.UserID,
		Description: []byte(p.Description),
		StorageKey:  []byte(normalizeSlash(p.StorageKey)),
		HashSum:     []byte(strings.ToLower(strings.TrimSpace(p.HashSum))),
		UpdatedAt:   time.Now(),
	}, nil
}

// Validate checks that the file creation parameters are valid and secure.
func (p *NewFileDataParams) Validate() error {
	validations := []func() error{
		p.validateStorageKey,
		p.validateHashSum,
	}

	// errs collects all validation errors encountered during file data validation.
	var errs []error
	for _, fn := range validations {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateStorageKey ensures the storage key is safe and valid.
func (p *NewFileDataParams) validateStorageKey() error {
	key := normalizeSlash(strings.TrimSpace(p.StorageKey))
	if !validStorageKey(key) {
		return ErrIncorrectStorageKey
	}
	return nil
}

// validateHashSum ensures the hash sum is a valid SHA256 hex string.
func (p *NewFileDataParams) validateHashSum() error {
	hs := strings.ToLower(strings.TrimSpace(p.HashSum))
	if !isHexSHA256(hs) {
		return ErrIncorrectHashSum
	}
	return nil
}

// validStorageKey validates that a storage key is safe and doesn't contain path traversal.
func validStorageKey(key string) bool {
	if key == "" {
		return false
	}
	if strings.HasPrefix(key, "/") || strings.HasPrefix(key, `\`) {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(key))
	if strings.HasPrefix(clean, "..") {
		return false
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}

const (
	// SHA256HexLength is the expected length of a SHA256 hash in hexadecimal format.
	SHA256HexLength = 64
)

// isHexSHA256 checks if a string is a valid SHA256 hash in hexadecimal format.
func isHexSHA256(s string) bool {
	if len(s) != SHA256HexLength {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// normalizeSlash normalizes path separators and removes unsafe prefixes.
func normalizeSlash(p string) string {
	p = strings.ReplaceAll(p, `\`, `/`)
	return strings.TrimPrefix(p, "./")
}
