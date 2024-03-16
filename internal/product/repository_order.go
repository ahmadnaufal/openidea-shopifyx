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

type purchaseCountResult struct {
	ProductID     string `db:"product_id"`
	PurchaseCount int    `db:"purchase_count"`
}

func (r ProductRepo) GetPurchaseCountByProductIDs(ctx context.Context, productIDs []string) (map[string]int, error) {
	query := `
		SELECT
			product_id,
			SUM(quantity) AS purchase_count
		FROM
			orders
		WHERE
			product_id IN (?)
		GROUP BY
			product_id
	`

	updatedQuery, args, err := sqlx.In(query, productIDs)
	if err != nil {
		return nil, err
	}

	var purchaseCounts []purchaseCountResult
	err = r.db.SelectContext(ctx, &purchaseCounts, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return nil, err
	}

	mapRes := map[string]int{}
	for _, v := range purchaseCounts {
		mapRes[v.ProductID] = v.PurchaseCount
	}

	return mapRes, nil
}

func (r ProductRepo) GetProductPurchasedByUserID(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT
			SUM(o.quantity) AS purchase_count
		FROM
			orders o
			INNER JOIN products p
			ON p.id = o.product_id
		WHERE
			p.user_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return count, err
	}

	return count, nil
}
