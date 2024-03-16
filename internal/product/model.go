package product

import "time"

type CreateProductRequest struct {
	Name          string   `json:"name" validate:"required,min=5,max=60"`
	Price         int      `json:"price" validate:"required,gte=0"`
	ImageURL      string   `json:"imageUrl" validate:"required,url"`
	Stock         int      `json:"stock" validate:"required,gte=0"`
	Condition     string   `json:"condition" validate:"oneof=new second"`
	Tags          []string `json:"tags" validate:"required,min=0"`
	IsPurchasable *bool    `json:"isPurchasable" validate:"required"`
}

type UpdateProductRequest struct {
	Name          string   `json:"name" validate:"required,min=5,max=60"`
	Price         int      `json:"price" validate:"required,gte=0"`
	ImageURL      string   `json:"imageUrl" validate:"required,url"`
	Condition     string   `json:"condition" validate:"oneof=new second"`
	Tags          []string `json:"tags" validate:"required,min=0"`
	IsPurchasable *bool    `json:"isPurchasable" validate:"required"`
}

type UpdateProductStockRequest struct {
	Stock int `json:"stock" validate:"required,gte=0"`
}

type BuyProductRequest struct {
	BankAccountID        string `json:"bankAccountId" validate:"required"`
	PaymentProofImageURL string `json:"paymentProofImageUrl" validate:"required,url"`
	Quantity             int    `json:"quantity" validate:"required,gte=1"`
}

type ListProductsRequest struct {
	UserOnly       bool     `query:"userOnly"`
	Limit          int      `query:"limit"`
	Offset         int      `query:"offset"`
	Tags           []string `query:"tags"`
	Condition      string   `query:"condition"`
	ShowEmptyStock bool     `query:"showEmptyStock"`
	MaxPrice       int      `query:"maxPrice"`
	MinPrice       int      `query:"minPrice"`
	SortBy         string   `query:"sortBy"`
	OrderBy        string   `query:"orderBy"`
	Search         string   `query:"search"`

	// UserID to store userID when userOnly flag is enabled
	UserID string
}

type Product struct {
	ID            string    `db:"id"`
	UserID        string    `db:"user_id"`
	Name          string    `db:"name"`
	Price         int       `db:"price"`
	ImageURL      string    `db:"image_url"`
	Stock         int       `db:"stock"`
	Condition     string    `db:"condition"`
	IsPurchasable bool      `db:"is_purchasable"`
	CreatedAt     time.Time `db:"created_at"`
}

type ProductTag struct {
	ID        int    `db:"id"`
	ProductID string `db:"product_id"`
	Tag       string `db:"tag"`
}

type Order struct {
	ID                   string `db:"id"`
	UserID               string `db:"user_id"`
	ProductID            string `db:"product_id"`
	BankAccountID        string `db:"bank_account_id"`
	PaymentProofImageURL string `db:"payment_proof_image_url"`
	Quantity             int    `db:"quantity"`
}
