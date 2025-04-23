package security

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2Hasher struct{}

var _ Hasher = (*Argon2Hasher)(nil)
var ErrHashMismatch = errors.New("hash mismatch")

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
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		Memory, Iterations, Parallelism, saltBase64, hashBase64)

	return encoded, nil
}

// Verify implements Hasher.
func (h *Argon2Hasher) Verify(plain string, hashed string) (bool, error) {
	parts := strings.Split(hashed, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid hash format")
	}

	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("base64 decode salt: %w", err)
	}

	actualHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("base64 decode hash: %w", err)
	}

	hashLen := len(actualHash)
	if hashLen > int(^uint32(0)) {
		return false, errors.New("hash length exceeds uint32")
	}

	computedHash := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(hashLen))
	if subtle.ConstantTimeCompare(computedHash, actualHash) == 1 {
		return true, nil
	}
	return false, nil
}
