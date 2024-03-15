package bankaccount

import (
	"database/sql"
	"fmt"

	"github.com/ahmadnaufal/openidea-shopifyx/internal/model"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	BankAccountRepoImpl *BankAccountRepo
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	bankAccountGroup := r.Group("/v1/bank/account")
	bankAccountGroup.Use(jwtProvider.Middleware())

	bankAccountGroup.Post("", CreateBankAccount)
	bankAccountGroup.Get("", ListBankAccounts)
	bankAccountGroup.Patch("/:bank_account_id", UpdateBankAccount)
	bankAccountGroup.Delete("/:bank_account_id", DeleteBankAccount)
}

func CreateBankAccount(c *fiber.Ctx) error {
	var payload BankAccountRequest

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "invalid_request_body",
		})
	}

	// validation for request body
	err = validate.Struct(payload)
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
	bankAccountID := uuid.NewString()
	bankAccount := BankAccount{
		ID:                bankAccountID,
		UserID:            claims.UserID,
		BankName:          payload.BankName,
		BankAccountName:   payload.BankAccountName,
		BankAccountNumber: payload.BankAccountNumber,
	}
	// save data to db
	err = BankAccountRepoImpl.CreateBankAccount(ctx, bankAccount)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "success",
		Data:    bankAccountEntityToResponse(bankAccount),
	})
}

func ListBankAccounts(c *fiber.Ctx) error {
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}
	// fetch the logged in user's bank accounts
	ctx := c.Context()
	bankAccounts, err := BankAccountRepoImpl.GetBankAccountsByUserID(ctx, claims.UserID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	bankAccountResponses := make([]BankAccountResponse, len(bankAccounts))
	for i, bankAccount := range bankAccounts {
		bankAccountResponses[i] = bankAccountEntityToResponse(bankAccount)
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "success",
		Data:    bankAccountResponses,
	})
}

func UpdateBankAccount(c *fiber.Ctx) error {
	var payload BankAccountRequest
	bankAccountID := c.Params("bank_account_id")

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "invalid_request_body",
		})
	}

	// validation for request body
	err = validate.Struct(payload)
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
	// check if the mentioned bank account exists
	bankAccount, err := BankAccountRepoImpl.GetBankAccountByID(ctx, bankAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "bank account not found",
				Code:    "entity_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// check bank account ownership
	if bankAccount.UserID != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "cannot update a bank account that is owned by another user",
			Code:    "update_bank_account_forbidden",
		})
	}

	// save data to db
	bankAccount.BankName = payload.BankName
	bankAccount.BankAccountName = payload.BankAccountName
	bankAccount.BankAccountNumber = payload.BankAccountNumber
	err = BankAccountRepoImpl.UpdateBankAccount(ctx, bankAccount)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "success",
		Data:    bankAccountEntityToResponse(bankAccount),
	})
}

func DeleteBankAccount(c *fiber.Ctx) error {
	bankAccountID := c.Params("bank_account_id")

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	ctx := c.Context()
	// check if the mentioned bank account exists
	bankAccount, err := BankAccountRepoImpl.GetBankAccountByID(ctx, bankAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "bank account not found",
				Code:    "entity_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// check bank account ownership
	if bankAccount.UserID != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "cannot delete a bank account that is owned by another user",
			Code:    "update_bank_account_forbidden",
		})
	}

	// delete bank account
	err = BankAccountRepoImpl.DeleteBankAccount(ctx, bankAccount.ID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "success",
	})
}

func bankAccountEntityToResponse(bankAccount BankAccount) BankAccountResponse {
	return BankAccountResponse{
		BankAccountID:     bankAccount.ID,
		BankName:          bankAccount.BankName,
		BankAccountName:   bankAccount.BankAccountName,
		BankAccountNumber: bankAccount.BankAccountNumber,
	}
}
