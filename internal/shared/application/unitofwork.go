package application

import (
	"context"
	"database/sql"
)

type TxFunc func(tx *sql.Tx) error

type UnitOfWork interface {
	Within(ctx context.Context, fn TxFunc) error
}

type SQLUnitOfWork struct {
	db *sql.DB
}

func NewSQLUnitOfWork(db *sql.DB) *SQLUnitOfWork {
	return &SQLUnitOfWork{db: db}
}

func (u *SQLUnitOfWork) Within(ctx context.Context, fn TxFunc) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
