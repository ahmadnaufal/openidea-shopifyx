package jwt

import (
	"encoding/base64"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

var defaultSigningMethod *jwt.SigningMethodHMAC = jwt.SigningMethodHS256

type JWTProvider struct {
	privateKey []byte
}

func NewJWTProvider(privateKey string) JWTProvider {
	privateKeyDecoded, _ := base64.StdEncoding.DecodeString(privateKey)

	return JWTProvider{
		privateKey: privateKeyDecoded,
	}
}

func (p JWTProvider) GenerateToken(payload JWTClaims) (string, error) {
	token, err := jwt.NewWithClaims(defaultSigningMethod, payload).SignedString(p.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "error returning signed string")
	}

	return token, nil
}
