package jwt

import (
	"encoding/base64"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

func (p *JWTProvider) GenerateToken(payload jwt.MapClaims) (string, error) {
	token, err := jwt.NewWithClaims(defaultSigningMethod, payload).SignedString(p.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "error returning signed string")
	}

	return token, nil
}

func (p *JWTProvider) MiddlewareWithPublic() fiber.Handler {
	return jwtware.New(jwtware.Config{
		ContextKey: "user",
		Claims:     jwt.MapClaims{},
		SigningKey: jwtware.SigningKey{
			JWTAlg: jwtware.HS256,
			Key:    p.privateKey,
		},
		Filter: func(c *fiber.Ctx) bool {
			// only filter if there's userOnly
			return !c.QueryBool("userOnly", false)
		},
	})
}

func (p *JWTProvider) Middleware() fiber.Handler {
	return jwtware.New(jwtware.Config{
		ContextKey: "user",
		Claims:     jwt.MapClaims{},
		SigningKey: jwtware.SigningKey{
			JWTAlg: jwtware.HS256,
			Key:    p.privateKey,
		},
	})
}

func GetLoggedInUser(c *fiber.Ctx) (JWTUser, error) {
	jwtUser := JWTUser{}

	user, exist := c.Locals("user").(*jwt.Token)
	if !exist {
		return jwtUser, errors.New("unable to get logged in user data")
	}

	claims, exist := user.Claims.(jwt.MapClaims)
	if !exist {
		return jwtUser, errors.New("unable to get logged in user data")
	}

	jwtUser.UserID = claims["userId"].(string)
	jwtUser.Name = claims["name"].(string)
	jwtUser.Username = claims["username"].(string)

	return jwtUser, nil
}
