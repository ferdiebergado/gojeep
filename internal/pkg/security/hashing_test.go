package security_test

import (
	"strings"
	"testing"

	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestArgon2HasherHash(t *testing.T) {
	hasher := &security.Argon2Hasher{}
	password := "securepassword"

	hashed, err := hasher.Hash(password)

	assert.NoError(t, err, "Hashing should not return an error")
	assert.NotEmpty(t, hashed, "Hashed password should not be empty")
	assert.True(t, strings.HasPrefix(hashed, "$argon2id$"), "Hashed password should have the correct prefix")
}
