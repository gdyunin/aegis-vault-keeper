package security

import (
	"crypto/rand"
	"fmt"
)

// CryptoKeyGenerator provides cryptographically secure key generation functionality.
type CryptoKeyGenerator struct{}

// NewCryptoKeyGenerator creates a new CryptoKeyGenerator instance.
func NewCryptoKeyGenerator() *CryptoKeyGenerator {
	return &CryptoKeyGenerator{}
}

// CryptoKeyGenerate generates a cryptographically secure random key of the specified size.
func (c *CryptoKeyGenerator) CryptoKeyGenerate(size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid key size: %d", size)
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate crypto key: %w", err)
	}

	return key, nil
}
