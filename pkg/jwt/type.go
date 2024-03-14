package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	JWTUser
	jwt.RegisteredClaims
}

type JWTUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

func BuildJWTClaims(user JWTUser, expireDuration time.Duration) JWTClaims {
	return JWTClaims{
		JWTUser: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
		},
	}
}
