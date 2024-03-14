package product

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type ProductRepo struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) ProductRepo {
	return ProductRepo{db: db}
}

func (r ProductRepo) CreateProduct(ctx context.Context, product Product) error {
	query := `
		INSERT INTO products
			(
				id,
				user_id
				name,
				price,
				image_url,
				stock,
				condition,
				is_purchasable
			)
		VALUES
			(
				:id,
				:user_id
				:name,
				:price,
				:image_url,
				:stock,
				:condition,
				:is_purchasable
			)
	`

	updatedQuery, args, err := sqlx.Named(query, product)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductRepo) CreateProductTags(ctx context.Context, tags []ProductTag) error {
	query := `
		INSERT INTO product_tags
			(
				product_id,
				tag
			)
		VALUES
			(
				:product_id,
				:tag
			)
	`

	updatedQuery, args, err := sqlx.Named(query, tags)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductRepo) UpdateProduct(ctx context.Context, id string, product Product) error {
	query := `
		INSERT INTO users
			(id, username, name, password)
		VALUES
			(:id, :username, :name, :password)
	`

	updatedQuery, args, err := sqlx.Named(query, user)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductRepo) GetProductByID(ctx context.Context, username string) (Product, error) {
	var result Product

	query := `
		SELECT
			id,
			username,
			name,
			password
		FROM
			users
		WHERE
			username = $1
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &result, query, username)
	if err != nil {
		return result, err
	}

	return result, nil
}
