//go:generate mockgen -destination=mock/signer_mock.go -package=mock . Signer
package security

import (
	"fmt"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type Signer interface {
	Sign(subject string, audience []string, duration string) (string, error)
	Verify(tokenString string) (string, error)
}

type signer struct {
	method jwt.SigningMethod
	cfg    config.JWTConfig
}

var _ Signer = (*signer)(nil)

func NewSigner(cfg config.JWTConfig) Signer {
	return &signer{
		method: jwt.SigningMethodHS256,
		cfg:    cfg,
	}
}

func (j *signer) Sign(subject string, audience []string, duration string) (string, error) {
	id, err := GenerateRandomBytesEncoded(j.cfg.KeyLen)
	if err != nil {
		return "", err
	}

	now := time.Now()
	ttl, err := time.ParseDuration(duration)
	if err != nil {
		return "", err
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    j.cfg.Issuer,
		Subject:   subject,
		ID:        id,
		Audience:  audience,
	}

	token := jwt.NewWithClaims(j.method, claims)
	return token.SignedString([]byte(j.cfg.SigningKey))
}

func (j *signer) Verify(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(j.cfg.SigningKey), nil
	}, jwt.WithValidMethods([]string{j.method.Alg()}))
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("token claims is not a RegisteredClaims: %T", token.Claims)
	}

	return claims.Subject, nil
}
