package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MaxBcryptInputLength is the maximum length of input data for bcrypt hashing.
	MaxBcryptInputLength = 72
)

// HashBcrypt creates a bcrypt hash of the input data using the default cost.
// Returns an error if the input exceeds the bcrypt maximum length limit of 72 bytes.
func HashBcrypt(data string) (string, error) {
	if len(data) > MaxBcryptInputLength {
		return "", fmt.Errorf(
			"bcrypt error: input exceeds maximum length of %d bytes (length: %d)",
			MaxBcryptInputLength,
			len(data),
		)
	}

	hashedData, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt error: failed to hash input data: %w", err)
	}
	return string(hashedData), nil
}

// VerifyBcrypt compares a bcrypt hash against plain text data for verification.
// Returns true if the data matches the hash, false if not, and an error only for verification failures.
func VerifyBcrypt(hashedData string, verifyingData string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedData), []byte(verifyingData))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("bcrypt error: failed to verify hash: %w", err)
	}
	return true, nil
}
