//go:generate mockgen -destination=mock/hasher_mock.go -package=mock . Hasher
package security

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Hasher interface {
	Hash(plain string) ([]byte, error)
	Verify(plain string, hashed []byte) error
}

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
func (h *Argon2Hasher) Hash(plain string) ([]byte, error) {
	// Generate a random salt
	salt, err := GenerateRandomBytes(SaltLength)
	if err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	// Hash the password
	hash := argon2.IDKey([]byte(plain), salt, Iterations, Memory, Parallelism, KeyLength)

	// Encode the salt and hash for storage
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	// Return the formatted password hash
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		Memory, Iterations, Parallelism, saltBase64, hashBase64)

	return []byte(encoded), nil
}

// Verify implements Hasher.
func (h *Argon2Hasher) Verify(plain string, hashed []byte) error {
	parts := strings.Split(string(hashed), "$")
	if len(parts) != 6 {
		return errors.New("invalid hash format")
	}

	memory, time, threads, err := parseParams(parts[3])
	if err != nil {
		return fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("base64 decode salt: %w", err)
	}

	actualHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("base64 decode hash: %w", err)
	}

	hashLen := len(actualHash)
	if hashLen > int(^uint32(0)) {
		return errors.New("hash length exceeds uint32")
	}

	computedHash := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(hashLen))
	if subtle.ConstantTimeCompare(computedHash, actualHash) == 0 {
		return ErrHashMismatch
	}
	return nil
}

// parseParams parses the argon2 param string like "m=65536,t=3,p=4"
func parseParams(paramStr string) (memory uint32, time uint32, threads uint8, err error) {
	params := strings.Split(paramStr, ",")
	if len(params) != 3 {
		return 0, 0, 0, errors.New("invalid argon2 parameter count")
	}

	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			return 0, 0, 0, fmt.Errorf("invalid param format: %q", param)
		}
		key, val := kv[0], kv[1]

		switch key {
		case "m":
			u, err := strconv.ParseUint(val, 10, 32)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("invalid memory: %w", err)
			}
			memory = uint32(u)
		case "t":
			u, err := strconv.ParseUint(val, 10, 32)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("invalid time: %w", err)
			}
			time = uint32(u)
		case "p":
			u, err := strconv.ParseUint(val, 10, 8)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("invalid threads: %w", err)
			}
			threads = uint8(u)
		default:
			return 0, 0, 0, fmt.Errorf("unexpected param: %s", key)
		}
	}

	return memory, time, threads, nil
}
