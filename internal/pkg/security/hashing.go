//go:generate mockgen -destination=mock/hasher_mock.go -package=mock . Hasher
package security

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Hasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hashed string) (bool, error)
}

type Argon2Hasher struct{}

var _ Hasher = (*Argon2Hasher)(nil)

// Parameters for the Argon2ID algorithm
const (
	Memory      = 64 * 1024 // 64 MB
	Iterations  = 3
	Parallelism = 2
	SaltLength  = 16 // 16 bytes
	KeyLength   = 32 // 32 bytes
)

// Hash implements Hasher.
func (h *Argon2Hasher) Hash(plain string) (string, error) {
	// Generate a random salt
	salt, err := GenerateRandomBytes(SaltLength)
	if err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	// Hash the password
	hash := argon2.IDKey([]byte(plain), salt, Iterations, Memory, Parallelism, KeyLength)

	// Encode the salt and hash for storage
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	// Return the formatted password hash
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		Memory, Iterations, Parallelism, saltBase64, hashBase64), nil
}

// Verify implements Hasher.
func (h *Argon2Hasher) Verify(plain string, hashed string) (bool, error) {
	parts := strings.Split(hashed, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	// Extract parameters and the salt/hash values
	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Check if len(expectedHash) can safely fit in a uint32
	hashLen := len(expectedHash)

	if hashLen > int(^uint32(0)) { // ^uint32(0) gives the max value of uint32
		return false, errors.New("expected hash length exceeds uint32 limits")
	}

	// Compute the hash with the same parameters
	computedHash := argon2.IDKey([]byte(plain), salt, iterations, memory, parallelism, uint32(hashLen))

	// Constant time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(computedHash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}
