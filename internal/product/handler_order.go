package product

import (
	"context"

	"github.com/ahmadnaufal/openidea-shopifyx/internal/model"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func BuyProduct(c *fiber.Ctx) error {
	var payload BuyProductRequest

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	productID := c.Params("product_id")
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "invalid_request_body",
		})
	}

	// validation for request body
	if err := validation.Validate(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "failed_request_body_validation",
		})
	}

	ctx := c.Context()
	order, err := validateAndCreateOrder(ctx, productID, claims.UserID, payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "success",
		Data: OrderResponse{
			ID:                   order.ID,
			ProductID:            order.ProductID,
			BankAccountID:        order.BankAccountID,
			PaymentProofImageURL: order.PaymentProofImageURL,
			Quantity:             order.Quantity,
		},
	})
}

func validateAndCreateOrder(ctx context.Context, productID, userID string, payload BuyProductRequest) (Order, error) {
	// check for bank account existence
	bankAccount, err := BankAccountRepoImpl.GetBankAccountByID(ctx, payload.BankAccountID)
	if err != nil {
		return Order{}, err
	}

	// check for product existence
	product, err := ProductRepoImpl.GetProductByID(ctx, productID)
	if err != nil {
		return Order{}, err
	}

	// return 400 for bank account & product incompatibility
	if bankAccount.UserID != product.UserID {
		return Order{}, errors.New("stock not available")
	}

	// return 400 if user tries to buy his/her own product
	if product.UserID == userID {
		return Order{}, errors.New("user cannot buy his/her own product")
	}

	// check for product stock existence
	if product.Stock < payload.Quantity {
		return Order{}, errors.New("stock cannot be reduced")
	}

	tx, err := TrxProvider.NewTransaction(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()

	// create the order
	orderID := uuid.NewString()
	order := Order{
		ID:                   orderID,
		UserID:               userID,
		ProductID:            product.ID,
		BankAccountID:        bankAccount.ID,
		PaymentProofImageURL: payload.PaymentProofImageURL,
		Quantity:             payload.Quantity,
	}
	err = ProductRepoImpl.CreateOrder(ctx, tx, order)
	if err != nil {
		return Order{}, err
	}

	// decrement order stock
	product.Stock -= payload.Quantity
	err = ProductRepoImpl.UpdateProductStock(ctx, tx, productID, product.Stock)
	if err != nil {
		return Order{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Order{}, err
	}

	return order, nil
}
