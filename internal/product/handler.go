package product

import (
	"fmt"

	"github.com/ahmadnaufal/openidea-shopifyx/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ProductRepoImpl *ProductRepo
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func RegisterRoute(r *fiber.App) {
	productGroup := r.Group("/v1/product")

	productGroup.Post("", CreateProduct)
	productGroup.Patch("/:product_id", UpdateProduct)
	productGroup.Delete("/:product_id", DeleteProduct)
	productGroup.Get("", ListProducts)
	productGroup.Get("/:product_id", GetProduct)
}

func CreateProduct(c *fiber.Ctx) error {
	var payload CreateProductRequest
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

	productID := uuid.NewString()
	product := Product{
		ID:            productID,
		Name:          payload.Name,
		Price:         payload.Price,
		ImageURL:      payload.ImageURL,
		Stock:         payload.Stock,
		Condition:     payload.Condition,
		IsPurchasable: payload.IsPurchasable,
	}
	err = ProductRepoImpl.CreateProduct(ctx, product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// create product tags
	productTags := []ProductTag{}
	for _, val := range payload.Tags {
		productTags = append(productTags, ProductTag{
			ProductID: productID,
			Tag:       val,
		})
	}
	err = ProductRepoImpl.CreateProductTags(ctx, productTags)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "Product created successfully",
		Data:    ProductResponse{},
	})
}

func UpdateProduct(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "Product updated successfully",
		Data:    ProductResponse{},
	})
}

func DeleteProduct(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "Product deleted successfully",
	})
}

func ListProducts(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data:    []ProductResponse{},
		Meta: &model.ResponseMeta{
			Limit:  0,
			Offset: 0,
			Total:  100,
		},
	})
}

func GetProduct(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data: ProductDetailResponse{
			Product: ProductResponse{},
			Seller:  ProductDetailSellerResponse{},
		},
	})
}
