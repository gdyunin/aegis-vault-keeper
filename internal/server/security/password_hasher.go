package security

// hashFunc defines the function signature for password hashing operations.
type hashFunc func(password string) (string, error)

// veriFunc defines the function signature for password verification operations.
type veriFunc func(hashedData, verifyingData string) (bool, error)

// PasswordHasherVerificator combines password hashing and verification functionality.
type PasswordHasherVerificator struct {
	// hasher is the function used to hash passwords.
	hasher hashFunc
	// verificator is the function used to verify passwords against hashes.
	verificator veriFunc
}

// NewPasswordHasherVerificator creates a new PasswordHasherVerificator with the provided functions.
func NewPasswordHasherVerificator(hasher hashFunc, verificator veriFunc) *PasswordHasherVerificator {
	return &PasswordHasherVerificator{
		hasher:      hasher,
		verificator: verificator,
	}
}

// PasswordHash hashes a plain text password using the configured hashing function.
func (p *PasswordHasherVerificator) PasswordHash(password string) (string, error) {
	return p.hasher(password)
}

// PasswordVerify verifies a plain text password against a hash using the configured verification function.
func (p *PasswordHasherVerificator) PasswordVerify(hashedData, verifyingData string) (bool, error) {
	return p.verificator(hashedData, verifyingData)
}
