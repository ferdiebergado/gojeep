package security_test

import (
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTSignAndVerify(t *testing.T) {
	cfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	jwtHandler := security.NewJWT(cfg)

	subject := "testuser"
	tokenString, err := jwtHandler.Sign(subject)
	assert.NoError(t, err)

	verifiedSubject, err := jwtHandler.Verify(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, subject, verifiedSubject)
}

func TestJWTVerifyInvalidToken(t *testing.T) {
	cfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	jwtHandler := security.NewJWT(cfg)

	invalidToken := "invalid-token"
	_, err := jwtHandler.Verify(invalidToken)
	assert.Error(t, err)
}

func TestJWTVerifyModifiedToken(t *testing.T) {
	cfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	jwtHandler := security.NewJWT(cfg)

	subject := "testuser"
	tokenString, err := jwtHandler.Sign(subject)
	assert.NoError(t, err)

	modifiedToken := tokenString + "modified"
	_, err = jwtHandler.Verify(modifiedToken)
	assert.Error(t, err)
}

func TestJWTVerifyExpiredToken(t *testing.T) {
	originalExpiration := 24 * time.Hour
	cfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    -1 * time.Hour, // expired token
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	jwtHandler := security.NewJWT(cfg)

	subject := "testuser"
	tokenString, err := jwtHandler.Sign(subject)
	cfg.Expiration = originalExpiration //reset
	assert.NoError(t, err)

	_, err = jwtHandler.Verify(tokenString)
	assert.Error(t, err)
}

func TestJWTVerifyWrongSigningKey(t *testing.T) {
	cfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	jwtHandler := security.NewJWT(cfg)

	subject := "testuser"
	tokenString, err := jwtHandler.Sign(subject)
	assert.NoError(t, err)

	wrongCfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "WRONGKEY",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
	}
	wrongJwtHandler := security.NewJWT(wrongCfg)

	_, err = wrongJwtHandler.Verify(tokenString)
	assert.Error(t, err)
}

func TestJWTVerifyWrongSigningMethod(t *testing.T) {
	cfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodHS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	jwtHandler := security.NewJWT(cfg)

	subject := "testuser"
	tokenString, err := jwtHandler.Sign(subject)
	assert.NoError(t, err)

	wrongCfg := security.JWTConfig{
		SigningMethod: jwt.SigningMethodRS256,
		SigningKey:    "CHANGEME",
		KeyLen:        32,
		Expiration:    24 * time.Hour,
		Issuer:        "TEST",
		Audience:      []string{"any"},
	}
	wrongJwtHandler := security.NewJWT(wrongCfg)

	_, err = wrongJwtHandler.Verify(tokenString)
	assert.Error(t, err)
}
