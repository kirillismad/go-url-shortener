package sql

import (
	"context"
	"database/sql"
)

func InTransaction(ctx context.Context, db *sql.DB, work func(*sql.Tx) error, opts ...TxOption) error {
	var txOptions *sql.TxOptions
	if len(opts) != 0 {
		txOptions = &sql.TxOptions{}
		for _, o := range opts {
			o(txOptions)
		}
	}

	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := work(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
