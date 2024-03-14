package product

type ProductResponse struct {
	ProductID     string   `json:"productId"`
	Name          string   `json:"name"`
	Price         int      `json:"price"`
	ImageURL      string   `json:"imageUrl"`
	Stock         int      `json:"stock"`
	Condition     string   `json:"condition"`
	Tags          []string `json:"tags"`
	IsPurchasable bool     `json:"isPurchasable"`
	PurchaseCount int      `json:"purchaseCount"`
}

type ProductDetailResponse struct {
	Product ProductResponse             `json:"product"`
	Seller  ProductDetailSellerResponse `json:"seller"`
}

type ProductDetailSellerResponse struct {
	Name             string                `json:"name"`
	ProductSoldTotal int                   `json:"productSoldTotal"`
	BankAccounts     []BankAccountResponse `json:"bankAccounts"`
}

type BankAccountResponse struct {
	BankAccountID     string `json:"bankAccountId"`
	BankName          string `json:"bankName"`
	BankAccountName   string `json:"bankAccountName"`
	BankAccountNumber string `json:"bankAccountNumber"`
}
