package security

import "golang.org/x/crypto/bcrypt"

// BcryptHasher implements Hasher using bcrypt.
type BcryptHasher struct {
	cost int
}

var _ Hasher = (*BcryptHasher)(nil)

// NewBcryptHasher returns a BcryptHasher with the given cost.
// Use bcrypt.DefaultCost (10) unless you have strong reason otherwise.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash generates a bcrypt hash of the password.
func (h *BcryptHasher) Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), h.cost)
}

// Verify compares a bcrypt hashed password with its possible plaintext equivalent.
// Returns nil if they match, or bcrypt.ErrMismatchedHashAndPassword if not.
func (h *BcryptHasher) Verify(password string, hashed []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, []byte(password))
}
