package filedata

import "errors"

var (
	// ErrIncorrectStorageKey indicates that the provided storage key is invalid or empty.
	ErrIncorrectStorageKey = errors.New("invalid storage key")
	// ErrIncorrectHashSum indicates that the provided hash sum is invalid or doesn't match.
	ErrIncorrectHashSum = errors.New("invalid hash sum")

	// ErrNewFileParamsValidation indicates that file creation parameters failed validation.
	ErrNewFileParamsValidation = errors.New("new file parameters validation failed")
)
