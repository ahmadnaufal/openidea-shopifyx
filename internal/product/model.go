package product

type CreateProductRequest struct {
	Name          string   `json:"name" validate:"required,min=5,max=60"`
	Price         int      `json:"price" validate:"required,gte=0"`
	ImageURL      string   `json:"imageUrl" validate:"required,url"`
	Stock         int      `json:"stock" validate:"required,gte=0"`
	Condition     string   `json:"condition" validate:"oneof=new second"`
	Tags          []string `json:"tags" validate:"required,min=0"`
	IsPurchasable bool     `json:"isPurchasable" validate:"required"`
}

type UpdateProductRequest struct {
	Name          string   `json:"name" validate:"required,min=5,max=60"`
	Price         int      `json:"price" validate:"required,gte=0"`
	ImageURL      string   `json:"imageUrl" validate:"required,url"`
	Condition     string   `json:"condition" validate:"oneof=new second"`
	Tags          []string `json:"tags" validate:"required,min=0"`
	IsPurchasable bool     `json:"isPurchasable" validate:"required"`
}

type Product struct {
	ID            string `db:"id"`
	UserID        string `db:"user_id"`
	Name          string `db:"name"`
	Price         int    `db:"price"`
	ImageURL      string `db:"image_url"`
	Stock         int    `db:"stock"`
	Condition     string `db:"condition"`
	IsPurchasable bool   `db:"is_purchasable"`
}

type ProductTag struct {
	ID        int    `db:"id"`
	ProductID string `db:"product_id"`
	Tag       string `db:"tag"`
}
