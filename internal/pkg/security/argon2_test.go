package security_test

import (
	"strings"
	"testing"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestArgon2HasherHash(t *testing.T) {
	t.Parallel()
	cfg := &config.Argon2Options{
		Memory:     65536,
		Iterations: 3,
		Threads:    2,
		SaltLength: 16,
		KeyLength:  32,
	}

	pepper := "pepper"

	hasher := security.NewArgon2Hasher(cfg, pepper)
	password := "securepassword"

	hashed, err := hasher.Hash(password)
	assert.NoError(t, err, "Hashing should not return an error")
	assert.NotEmpty(t, hashed, "Hashed password should not be empty")
	assert.True(t, strings.HasPrefix(hashed, "$argon2id$"), "Hashed password should have the correct prefix")
}

func TestArgon2Hasher_Verify(t *testing.T) {
	t.Parallel()
	cfg := &config.Argon2Options{
		Memory:     65536,
		Iterations: 3,
		Threads:    2,
		SaltLength: 16,
		KeyLength:  32,
	}
	pepper := "pepper"
	hasher := security.NewArgon2Hasher(cfg, pepper)
	password := "securepassword"

	hashed, err := hasher.Hash(password)
	if err != nil {
		t.Fatal(err)
	}

	isValid, err := hasher.Verify(password, hashed)
	assert.NoError(t, err)
	assert.True(t, isValid)
}
