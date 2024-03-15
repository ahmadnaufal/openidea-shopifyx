package bankaccount

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type BankAccountRepo struct {
	db *sqlx.DB
}

func NewBankAccountRepo(db *sqlx.DB) BankAccountRepo {
	return BankAccountRepo{db: db}
}

func (r BankAccountRepo) CreateBankAccount(ctx context.Context, bankAccount BankAccount) error {
	query := `
		INSERT INTO bank_accounts
			(id, user_id, bank_name, bank_account_name, bank_account_number)
		VALUES
			(:id, :user_id, :bank_name, :bank_account_name, :bank_account_number)
	`

	updatedQuery, args, err := sqlx.Named(query, bankAccount)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r BankAccountRepo) GetBankAccountsByUserID(ctx context.Context, userID string) ([]BankAccount, error) {
	var result []BankAccount

	query := `
		SELECT
			id,
			user_id,
			bank_name,
			bank_account_name,
			bank_account_number
		FROM
			bank_accounts
		WHERE
			user_id = $1
			AND deleted_at IS NULL
		ORDER BY
			created_at DESC
	`

	err := r.db.SelectContext(ctx, &result, query, userID)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r BankAccountRepo) GetBankAccountByID(ctx context.Context, bankAccountID string) (BankAccount, error) {
	var result BankAccount

	query := `
		SELECT
			id,
			user_id,
			bank_name,
			bank_account_name,
			bank_account_number
		FROM
			bank_accounts
		WHERE
			id = $1
			AND deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &result, query, bankAccountID)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r BankAccountRepo) UpdateBankAccount(ctx context.Context, bankAccount BankAccount) error {
	query := `
		UPDATE
			bank_accounts
		SET
			bank_name = :bank_name,
			bank_account_name = :bank_account_name,
			bank_account_number = :bank_account_number
		WHERE
			id = :id
			AND deleted_at IS NULL
	`

	updatedQuery, args, err := sqlx.Named(query, bankAccount)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r BankAccountRepo) DeleteBankAccount(ctx context.Context, bankAccountID string) error {
	query := `
		UPDATE
			bank_accounts
		SET
			deleted_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, bankAccountID)
	if err != nil {
		return err
	}

	return nil
}
