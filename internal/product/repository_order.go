package product

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func (r ProductRepo) CreateOrder(ctx context.Context, tx *sql.Tx, order Order) error {
	query := `
		INSERT INTO orders
			(
				id,
				user_id,
				product_id,
				bank_account_id,
				payment_proof_image_url,
				quantity
			)
		VALUES
			(
				:id,
				:user_id,
				:product_id,
				:bank_account_id,
				:payment_proof_image_url,
				:quantity
			)
	`

	updatedQuery, args, err := sqlx.Named(query, order)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	} else {
		_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	}

	if err != nil {
		return err
	}

	return nil
}
