package security_test

import (
	"strings"
	"testing"

	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestArgon2HasherHash(t *testing.T) {
	t.Parallel()
	hasher := &security.Argon2Hasher{}
	password := "securepassword"

	hashed, err := hasher.Hash(password)
	assert.NoError(t, err, "Hashing should not return an error")
	assert.NotEmpty(t, hashed, "Hashed password should not be empty")
	assert.True(t, strings.HasPrefix(string(hashed), "$argon2id$"), "Hashed password should have the correct prefix")
}

func TestArgon2Hasher_Verify(t *testing.T) {
	t.Parallel()
	hasher := &security.Argon2Hasher{}
	password := "securepassword"

	hashed, err := hasher.Hash(password)
	if err != nil {
		t.Fatal(err)
	}

	isValid, err := hasher.Verify(password, hashed)
	assert.NoError(t, err)
	assert.True(t, isValid)
}
