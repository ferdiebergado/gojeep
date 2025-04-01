package security_test

import (
	"testing"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/stretchr/testify/assert"
)

var audience = []string{"localhost/verify"}

func TestJWTSignAndVerify(t *testing.T) {
	cfg := config.JWTConfig{
		SigningKey: "CHANGEME",
		KeyLen:     32,
		Issuer:     "TEST",
	}
	jwtHandler := security.NewSigner(cfg)

	subject := "testuser"
	ttl := "24h"
	tokenString, err := jwtHandler.Sign(subject, audience, ttl)
	assert.NoError(t, err)

	verifiedSubject, err := jwtHandler.Verify(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, subject, verifiedSubject)
}

func TestJWTVerifyInvalidToken(t *testing.T) {
	cfg := config.JWTConfig{
		SigningKey: "CHANGEME",
		KeyLen:     32,
		Issuer:     "TEST",
	}
	jwtHandler := security.NewSigner(cfg)

	invalidToken := "invalid-token"
	_, err := jwtHandler.Verify(invalidToken)
	assert.Error(t, err)
}

func TestJWTVerifyModifiedToken(t *testing.T) {
	cfg := config.JWTConfig{
		SigningKey: "CHANGEME",
		KeyLen:     32,
		Issuer:     "TEST",
	}
	jwtHandler := security.NewSigner(cfg)

	subject := "testuser"
	audience := []string{"localhost/verify"}
	ttl := "24h"

	tokenString, err := jwtHandler.Sign(subject, audience, ttl)
	assert.NoError(t, err)

	modifiedToken := tokenString + "modified"
	_, err = jwtHandler.Verify(modifiedToken)
	assert.Error(t, err)
}

func TestJWTVerifyExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		SigningKey: "CHANGEME",
		KeyLen:     32,
		Issuer:     "TEST",
	}
	jwtHandler := security.NewSigner(cfg)

	subject := "testuser"
	audience := []string{"localhost/verify"}
	ttl := "-1h"

	tokenString, err := jwtHandler.Sign(subject, audience, ttl)
	assert.NoError(t, err)

	_, err = jwtHandler.Verify(tokenString)
	assert.Error(t, err)
}

func TestJWTVerifyWrongSigningKey(t *testing.T) {
	cfg := config.JWTConfig{
		SigningKey: "CHANGEME",
		KeyLen:     32,
		Issuer:     "TEST",
	}
	jwtHandler := security.NewSigner(cfg)

	subject := "testuser"
	audience := []string{"localhost/verify"}
	ttl := "24h"

	tokenString, err := jwtHandler.Sign(subject, audience, ttl)
	assert.NoError(t, err)

	wrongCfg := config.JWTConfig{
		SigningKey: "WRONGKEY",
		KeyLen:     32,
	}
	wrongJwtHandler := security.NewSigner(wrongCfg)

	_, err = wrongJwtHandler.Verify(tokenString)
	assert.Error(t, err)
}
