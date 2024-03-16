package product

import (
	"context"
	"database/sql"
	"fmt"

	bankaccount "github.com/ahmadnaufal/openidea-shopifyx/internal/bank_account"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/config"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/model"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/user"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ProductRepoImpl     *ProductRepo
	UserRepoImpl        *user.UserRepo
	BankAccountRepoImpl *bankaccount.BankAccountRepo
	TrxProvider         *config.TransactionProvider
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	productGroup := r.Group("/v1/product")

	authMiddleware := jwtProvider.Middleware()
	authPublicMiddleware := jwtProvider.MiddlewareWithPublic()

	productGroup.Post("", authMiddleware, CreateProduct)
	productGroup.Patch("/:product_id", authMiddleware, UpdateProduct)
	productGroup.Delete("/:product_id", authMiddleware, DeleteProduct)
	productGroup.Post("/:product_id/stock", authMiddleware, UpdateProductStock)
	productGroup.Post("/:product_id/buy", authMiddleware, BuyProduct)

	// endpoints that can be public
	productGroup.Get("", authPublicMiddleware, ListProducts)
	productGroup.Get("/:product_id", GetProduct)
}

func CreateProduct(c *fiber.Ctx) error {
	var payload CreateProductRequest

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
	userID := claims.UserID
	product, err := saveProductAndTags(ctx, userID, payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "Product created successfully",
		Data: ProductResponse{
			ProductID:     product.ID,
			Name:          product.Name,
			Price:         product.Price,
			ImageURL:      product.ImageURL,
			Stock:         product.Stock,
			Condition:     product.Condition,
			Tags:          payload.Tags,
			IsPurchasable: product.IsPurchasable,
			PurchaseCount: 0,
		},
	})
}

func saveProductAndTags(ctx context.Context, userID string, payload CreateProductRequest) (Product, error) {
	tx, err := TrxProvider.NewTransaction(ctx)
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback()

	productID := uuid.NewString()
	product := Product{
		ID:            productID,
		UserID:        userID,
		Name:          payload.Name,
		Price:         payload.Price,
		ImageURL:      payload.ImageURL,
		Stock:         payload.Stock,
		Condition:     payload.Condition,
		IsPurchasable: *payload.IsPurchasable,
	}
	err = ProductRepoImpl.CreateProduct(ctx, tx, product)
	if err != nil {
		return Product{}, err
	}

	// create product tags
	productTags := []ProductTag{}
	for _, val := range payload.Tags {
		productTags = append(productTags, ProductTag{
			ProductID: productID,
			Tag:       val,
		})
	}
	err = ProductRepoImpl.CreateProductTags(ctx, tx, productTags)
	if err != nil {
		return Product{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Product{}, err
	}

	return product, nil
}

func UpdateProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	var payload UpdateProductRequest
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

	// check existing product
	product, err := ProductRepoImpl.GetProductByID(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "product not found",
				Code:    "entity_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// check product ownership
	if product.UserID != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "cannot update a product that is owned by another user",
			Code:    "update_product_forbidden",
		})
	}

	product, err = updateProductAndTags(ctx, product, payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "Product updated successfully",
		Data: ProductResponse{
			ProductID:     product.ID,
			Name:          product.Name,
			Price:         product.Price,
			ImageURL:      product.ImageURL,
			Stock:         product.Stock,
			Condition:     product.Condition,
			Tags:          payload.Tags,
			IsPurchasable: product.IsPurchasable,
			PurchaseCount: 0,
		},
	})
}

func updateProductAndTags(ctx context.Context, product Product, payload UpdateProductRequest) (Product, error) {
	tx, err := TrxProvider.NewTransaction(ctx)
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback()

	product.Name = payload.Name
	product.Price = payload.Price
	product.ImageURL = payload.ImageURL
	product.Condition = payload.Condition
	product.IsPurchasable = *payload.IsPurchasable
	err = ProductRepoImpl.UpdateProduct(ctx, tx, product)
	if err != nil {
		return Product{}, err
	}

	// product tag updates: get existing tags for the product
	productToTagMap, err := ProductRepoImpl.BulkGetProductTags(ctx, []string{product.ID})
	if err != nil {
		return Product{}, err
	}
	existingTags := productToTagMap[product.ID]

	existingTagMap := map[string]ProductTag{}

	for _, val := range existingTags {
		existingTagMap[val.Tag] = val
	}

	newProductTags := []ProductTag{}
	for _, tag := range payload.Tags {
		if _, ok := existingTagMap[tag]; !ok {
			// tag is new: add to newProductTags
			newProductTags = append(newProductTags, ProductTag{
				ProductID: product.ID,
				Tag:       tag,
			})

			continue
		}

		// else: remove it from existingTagMap, so we can use the one that still exists for deletion
		delete(existingTagMap, tag)
	}

	// for the one that still exists in existingTagMap, remove them
	deletedProductTags := []ProductTag{}
	for _, existingTag := range existingTagMap {
		deletedProductTags = append(deletedProductTags, existingTag)
	}

	if len(newProductTags) > 0 {
		err = ProductRepoImpl.CreateProductTags(ctx, tx, newProductTags)
		if err != nil {
			return Product{}, err
		}
	}

	if len(deletedProductTags) > 0 {
		err = ProductRepoImpl.DeleteProductTags(ctx, tx, deletedProductTags)
		if err != nil {
			return Product{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return Product{}, err
	}

	return product, err
}

func DeleteProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	ctx := c.Context()

	// check existing product
	product, err := ProductRepoImpl.GetProductByID(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "product not found",
				Code:    "entity_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// check product ownership
	if product.UserID != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "cannot update a product that is owned by another user",
			Code:    "update_product_forbidden",
		})
	}

	err = deleteProductAndTags(ctx, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "Product deleted successfully",
	})
}

func deleteProductAndTags(ctx context.Context, productID string) error {
	tx, err := TrxProvider.NewTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = ProductRepoImpl.DeleteProduct(ctx, tx, productID)
	if err != nil {
		return err
	}

	// delete all product tags associated with this product
	err = ProductRepoImpl.DeleteProductTagsByProductID(ctx, tx, productID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func ListProducts(c *fiber.Ctx) error {
	var req ListProductsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "invalid_request_body",
		})
	}

	if req.UserOnly {
		claims, err := jwt.GetLoggedInUser(c)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
				Code:    "forbidden",
			})
		}

		req.UserID = claims.UserID
	}

	ctx := c.Context()
	products, count, err := ProductRepoImpl.ListProducts(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	productIDs := []string{}
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	// TODO get product tags and count purchase count for each
	productTagsMap := map[string][]ProductTag{}
	if len(productIDs) > 0 {
		productTagsMap, err = ProductRepoImpl.BulkGetProductTags(ctx, productIDs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "something wrong with the server. Please contact admin",
				Code:    "internal_server_error",
			})
		}
	}

	responses := []ProductResponse{}
	for _, product := range products {
		productTags := productTagsMap[product.ID]
		tags := []string{}
		for _, tag := range productTags {
			tags = append(tags, tag.Tag)
		}

		responses = append(responses, ProductResponse{
			ProductID:     product.ID,
			Name:          product.Name,
			Price:         product.Price,
			ImageURL:      product.ImageURL,
			Stock:         product.Stock,
			Condition:     product.Condition,
			Tags:          tags,
			IsPurchasable: product.IsPurchasable,
			PurchaseCount: 0,
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data:    responses,
		Meta: &model.ResponseMeta{
			Limit:  req.Limit,
			Offset: req.Offset,
			Total:  count,
		},
	})
}

func GetProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")

	ctx := c.Context()

	product, err := ProductRepoImpl.GetProductByID(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "product not found",
				Code:    "entity_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	tagsMap, err := ProductRepoImpl.BulkGetProductTags(ctx, []string{productID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}
	productTags := tagsMap[productID]

	strTags := []string{}
	for _, v := range productTags {
		strTags = append(strTags, v.Tag)
	}

	// populate user
	productUser, err := UserRepoImpl.GetUserByID(ctx, product.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// get user bank accounts
	bankAccounts, err := BankAccountRepoImpl.GetBankAccountsByUserID(ctx, product.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}
	bankAccountResponses := []BankAccountResponse{}
	for _, bankAccount := range bankAccounts {
		bankAccountResponses = append(bankAccountResponses, BankAccountResponse{
			BankAccountID:     bankAccount.ID,
			BankName:          bankAccount.BankName,
			BankAccountName:   bankAccount.BankAccountName,
			BankAccountNumber: bankAccount.BankAccountNumber,
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data: ProductDetailResponse{
			Product: ProductResponse{
				ProductID:     product.ID,
				Name:          product.Name,
				Price:         product.Price,
				ImageURL:      product.ImageURL,
				Stock:         product.Stock,
				Condition:     product.Condition,
				Tags:          strTags,
				IsPurchasable: product.IsPurchasable,
				PurchaseCount: 0,
			},
			Seller: ProductDetailSellerResponse{
				Name:             productUser.Name,
				ProductSoldTotal: 0,
				BankAccounts:     bankAccountResponses,
			},
		},
	})
}

func UpdateProductStock(c *fiber.Ctx) error {
	productID := c.Params("product_id")

	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	var payload UpdateProductStockRequest
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

	product, err := ProductRepoImpl.GetProductByID(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "product not found",
				Code:    "entity_not_found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	// check product ownership
	if product.UserID != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "cannot update a product that is owned by another user",
			Code:    "update_product_forbidden",
		})
	}

	// update product stock
	err = ProductRepoImpl.UpdateProductStock(ctx, nil, productID, payload.Stock)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data: ProductResponse{
			ProductID:     product.ID,
			Name:          product.Name,
			Price:         product.Price,
			ImageURL:      product.ImageURL,
			Stock:         product.Stock,
			Condition:     product.Condition,
			IsPurchasable: product.IsPurchasable,
			PurchaseCount: 0,
		},
	})
}
