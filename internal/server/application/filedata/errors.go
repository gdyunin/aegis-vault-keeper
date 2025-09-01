package filedata

import (
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/errutil"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
)

var (
	// ErrFileAppError indicates a general file application error.
	ErrFileAppError = errors.New("file application error")

	// ErrFileTechError indicates a technical error in file operations.
	ErrFileTechError = errors.New("file technical error")

	// ErrFileIncorrectStorageKey indicates an invalid or malformed storage key.
	ErrFileIncorrectStorageKey = errors.New("incorrect storage key")

	// ErrFileIncorrectHashSum indicates a hash mismatch during file integrity verification.
	ErrFileIncorrectHashSum = errors.New("incorrect hash sum")

	// ErrFileDataRequired indicates that file data is required but not provided.
	ErrFileDataRequired = errors.New("file data is required")

	// ErrRollBackFileSaveFailed indicates that cleanup after a failed file save operation failed.
	ErrRollBackFileSaveFailed = errors.New("rollback of file save failed")

	// ErrFileNotFound indicates that the requested file does not exist.
	ErrFileNotFound = errors.New("file not found")

	// ErrFileAccessDenied indicates that the user lacks permission to access the file.
	ErrFileAccessDenied = errors.New("access to this file is denied")
)

// mapError maps domain layer errors to application layer errors for consistent error handling.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	mapped := errutil.MapError(mapFn, err)
	if mapped != nil {
		return fmt.Errorf("error after mapping: %w", mapped)
	}
	return nil
}

// mapFn provides the specific error mapping logic for file data domain errors.
func mapFn(err error) error {
	switch {
	case errors.Is(err, filedata.ErrNewFileParamsValidation):
		return ErrFileAppError
	case errors.Is(err, filedata.ErrIncorrectStorageKey):
		return ErrFileIncorrectStorageKey
	case errors.Is(err, filedata.ErrIncorrectHashSum):
		return ErrFileIncorrectHashSum
	default:
		return errors.Join(ErrFileTechError, err)
	}
}
