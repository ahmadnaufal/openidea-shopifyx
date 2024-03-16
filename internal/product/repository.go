package product

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ProductRepo struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) ProductRepo {
	return ProductRepo{db: db}
}

func (r ProductRepo) CreateProduct(ctx context.Context, tx *sql.Tx, product Product) error {
	query := `
		INSERT INTO products
			(
				id,
				user_id,
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
				:user_id,
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

func (r ProductRepo) CreateProductTags(ctx context.Context, tx *sql.Tx, tags []ProductTag) error {
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

func (r ProductRepo) UpdateProduct(ctx context.Context, tx *sql.Tx, product Product) error {
	query := `
		UPDATE products
		SET
			name = :name,
			price = :price,
			image_url = :image_url,
			condition = :condition,
			is_purchasable = :is_purchasable,
			updated_at = NOW()
		WHERE
			id = :id
			AND deleted_at IS NULL
	`

	updatedQuery, args, err := sqlx.Named(query, product)
	if err != nil {
		return err
	}

	var result sql.Result
	if tx != nil {
		result, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	} else {
		result, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	}
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows != 1 {
		return errors.New("error affected row count is not equal to 1")
	}

	return nil
}

func (r ProductRepo) GetProductByID(ctx context.Context, id string) (Product, error) {
	var result Product

	query := `
		SELECT
			id,
			user_id,
			name,
			price,
			image_url,
			stock,
			condition,
			is_purchasable
		FROM
			products
		WHERE
			id = $1
			AND deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &result, query, id)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r ProductRepo) BulkGetProductTags(ctx context.Context, productIDs []string) (map[string][]ProductTag, error) {
	var result []ProductTag

	query := `
		SELECT
			id,
			product_id,
			tag
		FROM
			product_tags
		WHERE
			product_id IN (?)
		ORDER BY
			product_id ASC, tag ASC	
	`

	updatedQuery, args, err := sqlx.In(query, productIDs)
	if err != nil {
		return nil, err
	}

	err = r.db.SelectContext(ctx, &result, r.db.Rebind(updatedQuery), args...)
	if err != nil {
		return nil, err
	}

	// group the fetched tags by each of product IDs
	productIDToTagMap := map[string][]ProductTag{}
	for _, tag := range result {
		productID := tag.ProductID
		productIDToTagMap[productID] = append(productIDToTagMap[productID], tag)
	}

	return productIDToTagMap, nil
}

func (r ProductRepo) DeleteProductTags(ctx context.Context, tx *sql.Tx, tags []ProductTag) error {
	productTagIDs := []int{}
	for _, v := range tags {
		productTagIDs = append(productTagIDs, v.ID)
	}

	query := `
		DELETE FROM
			product_tags
		WHERE
			id IN (?)
	`

	updatedQuery, args, err := sqlx.In(query, productTagIDs)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductRepo) DeleteProductTagsByProductID(ctx context.Context, tx *sql.Tx, productID string) error {
	query := `
		DELETE FROM
			product_tags
		WHERE
			product_id = $1
	`

	_, err := tx.ExecContext(ctx, query, productID)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductRepo) DeleteProduct(ctx context.Context, tx *sql.Tx, productID string) error {
	query := `
		UPDATE
			products
		SET
			updated_at = NOW(),
			deleted_at = NOW()
		WHERE
			id = $1
	`

	_, err := tx.ExecContext(ctx, query, productID)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductRepo) UpdateProductStock(ctx context.Context, tx *sql.Tx, productID string, updatedStock int) error {
	query := `
		UPDATE products
		SET
			stock = $1
		WHERE
			id = $2
			AND deleted_at IS NULL
	`

	var (
		result sql.Result
		err    error
	)

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, updatedStock, productID)
	} else {
		result, err = r.db.ExecContext(ctx, query, updatedStock, productID)
	}
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows != 1 {
		return errors.New("error affected row count is not equal to 1")
	}

	return nil
}

func (r ProductRepo) ListProducts(ctx context.Context, req ListProductsRequest) ([]Product, int, error) {
	var products []Product

	baseQuery := `
		SELECT DISTINCT
			p.id AS id,
			p.user_id,
			p.name,
			p.price,
			p.image_url,
			p.stock,
			p.condition,
			p.is_purchasable,
			p.created_at
		FROM
			products p
			INNER JOIN product_tags pt
			ON p.id = pt.product_id
		WHERE
			p.deleted_at IS NULL %s
	`

	args := []interface{}{}

	filterQuery, filterArgs := getFilter(req)

	args = append(args, filterArgs...)

	queryWithFilter := fmt.Sprintf(baseQuery, filterQuery)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS temp", queryWithFilter)

	var count int
	err := r.db.GetContext(ctx, &count, sqlx.Rebind(sqlx.DOLLAR, countQuery), args...)
	if err != nil {
		return products, count, err
	}

	orderQuery := getSortBy(req)
	limitQuery, limitArgs := getLimitAndOffset(req)
	args = append(args, limitArgs...)

	query := fmt.Sprintf("%s %s %s", queryWithFilter, orderQuery, limitQuery)

	err = r.db.SelectContext(ctx, &products, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	if err != nil {
		return products, count, err
	}

	return products, count, nil
}

func getFilter(req ListProductsRequest) (string, []interface{}) {
	args := []interface{}{}
	filter := ""

	if req.UserOnly {
		filter += " AND p.user_id = ?"
		args = append(args, req.UserID)
	}

	// tags, bit complex
	if len(req.Tags) > 0 {
		placeholders := []string{}
		for _, tag := range req.Tags {
			args = append(args, tag)
			placeholders = append(placeholders, "?")
		}
		filter += fmt.Sprintf(" AND pt.tag IN (%s)", strings.Join(placeholders, ", "))
	}

	if req.Condition != "" {
		filter += " AND p.condition = ?"
		args = append(args, req.Condition)
	}

	if !req.ShowEmptyStock {
		filter += " AND p.stock > 0"
	}

	if req.MaxPrice > 0 {
		filter += " AND p.price <= ?"
		args = append(args, req.MaxPrice)
	}

	if req.MinPrice > 0 {
		filter += " AND p.price >= ?"
		args = append(args, req.MinPrice)
	}

	if req.Search != "" {
		filter += " AND p.name ILIKE '%' || ? || '%'"
		args = append(args, req.Search)
	}

	return filter, args
}

func getSortBy(req ListProductsRequest) string {
	sortColumn := strings.ToLower(req.SortBy)
	if sortColumn != "price" && sortColumn != "created_at" {
		sortColumn = "created_at"
	}

	sortOrdering := strings.ToUpper(req.OrderBy)
	if sortOrdering != "ASC" && sortOrdering != "DESC" {
		sortOrdering = "DESC"
	}

	query := fmt.Sprintf(`
		ORDER BY
			p.%s %s 
	`, sortColumn, sortOrdering)

	return query
}

func getLimitAndOffset(req ListProductsRequest) (string, []interface{}) {
	// by default, set limit to 50
	query := "LIMIT ? OFFSET ?"

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	args := []interface{}{limit, offset}

	return query, args
}
