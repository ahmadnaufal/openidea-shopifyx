package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ahmadnaufal/openidea-shopifyx/internal/model"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	UserRepoImpl *UserRepo
	JwtProvider  *jwt.JWTProvider
	SaltCost     int
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func RegisterRoute(r *fiber.App) {
	r.Post("/v1/user/register", RegisterUser)
	r.Post("/v1/user/login", Authenticate)
}

func RegisterUser(c *fiber.Ctx) error {
	var payload RegisterUserRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "invalid_request_body",
		})
	}

	// validation for request body
	err := validate.Struct(payload)
	if err != nil {
		strError := ""

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, vErr := range validationErrors {
				strError += fmt.Sprintf("%s;", vErr.Error())
			}
		} else {
			strError = err.Error()
		}

		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: strError,
			Code:    "failed_request_body_validation",
		})
	}

	ctx := c.Context()

	// get by username first to check if its already registered. return conflict
	existingUser, err := UserRepoImpl.GetUserByUsername(ctx, payload.Username)
	if err != nil && err != sql.ErrNoRows {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}
	if existingUser.ID != "" {
		return c.Status(fiber.StatusConflict).JSON(model.ErrorResponse{
			Message: "username already used",
			Code:    "username_already_exists",
		})
	}

	// no problem. save the data by generating the password
	user, err := createUser(ctx, payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// generate JWT
	accessToken, err := generateAccessTokenFromUser(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data: UserAuthResponse{
			Username:    user.Username,
			Name:        user.Name,
			AccessToken: accessToken,
		},
	})
}

func createUser(ctx context.Context, payload RegisterUserRequest) (User, error) {
	// hash the password first using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), SaltCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:       uuid.NewString(),
		Username: payload.Username,
		Name:     payload.Name,
		Password: string(hashedPassword),
	}
	err = UserRepoImpl.CreateUser(ctx, user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func Authenticate(c *fiber.Ctx) error {
	var payload AuthenticateRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "invalid_request_body",
		})
	}

	// validation for request body
	err := validate.Struct(payload)
	if err != nil {
		strError := ""

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, vErr := range validationErrors {
				strError += fmt.Sprintf("%s;", vErr.Error())
			}
		} else {
			strError = err.Error()
		}

		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: strError,
			Code:    "failed_request_body_validation",
		})
	}

	ctx := c.Context()

	user, err := UserRepoImpl.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "username does not exists",
				Code:    "username_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// verify login
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Wrong password entered for the username",
			Code:    "wrong_password",
		})
	}

	// generate JWT
	accessToken, err := generateAccessTokenFromUser(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data: UserAuthResponse{
			Username:    user.Username,
			Name:        user.Name,
			AccessToken: accessToken,
		},
	})
}

func generateAccessTokenFromUser(user User) (string, error) {
	claims := jwt.BuildJWTClaims(jwt.JWTUser{
		UserID:   user.ID,
		Username: user.Username,
		Name:     user.Name,
	}, 3*time.Minute)

	accessToken, err := JwtProvider.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return accessToken, err
}
