package bankaccount

type BankAccountRequest struct {
	BankName          string `json:"bankName" validate:"required,min=5,max=15"`
	BankAccountName   string `json:"bankAccountName" validate:"required,min=5,max=15"`
	BankAccountNumber string `json:"bankAccountNumber" validate:"required,min=5,max=15"`
}

type BankAccount struct {
	ID                string `db:"id"`
	UserID            string `db:"user_id"`
	BankName          string `db:"bank_name"`
	BankAccountName   string `db:"bank_account_name"`
	BankAccountNumber string `db:"bank_account_number"`
}
