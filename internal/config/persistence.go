package config

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TransactionProvider struct {
	db *sqlx.DB
}

func NewTransactionProvider(db *sqlx.DB) TransactionProvider {
	return TransactionProvider{
		db: db,
	}
}

func (p *TransactionProvider) NewTransaction(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}
