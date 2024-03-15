package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUser struct {
	UserID   string `json:"userId"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

func BuildJWTClaims(user JWTUser, expireDuration time.Duration) jwt.MapClaims {
	return jwt.MapClaims{
		"userId":   user.UserID,
		"name":     user.Name,
		"username": user.Username,
		"exp":      jwt.NewNumericDate(time.Now().Add(3 * time.Minute)),
	}
}
