package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SigningMethod jwt.SigningMethod
	SigningKey    string
	KeyLen        uint32
	Expiration    time.Duration
	Issuer        string
	Audience      []string
}

type JWT struct {
	cfg JWTConfig
}

func NewJWT(cfg JWTConfig) *JWT {
	return &JWT{
		cfg: cfg,
	}
}

func (j *JWT) Sign(subject string) (string, error) {
	id, err := GenerateRandomBytesEncoded(j.cfg.KeyLen)
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(j.cfg.Expiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    j.cfg.Issuer,
		Subject:   subject,
		ID:        id,
		Audience:  j.cfg.Audience,
	}
	token := jwt.NewWithClaims(j.cfg.SigningMethod, claims)
	return token.SignedString([]byte(j.cfg.SigningKey))
}

func (j *JWT) Verify(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(j.cfg.SigningKey), nil
	}, jwt.WithValidMethods([]string{j.cfg.SigningMethod.Alg()}))
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("token claims is not a RegisteredClaims: %T", token.Claims)
	}

	return claims.Subject, nil
}
