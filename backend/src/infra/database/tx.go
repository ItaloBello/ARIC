package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// EmTransacao roda fn dentro de uma transação: commit se fn devolver nil,
// rollback em erro ou panic.
func EmTransacao(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao abrir transação: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
