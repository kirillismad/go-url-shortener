package repo

import (
	"context"
	"database/sql"

	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type LinkRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *LinkRepository {
	r := new(LinkRepository)
	r.db = db
	r.q = sqlc.New(db)
	return r
}

func (r *LinkRepository) InTransaction(ctx context.Context, work func(*LinkRepository) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txRepo := &LinkRepository{
		db: r.db,
		q:  r.q.WithTx(tx),
	}

	if err := work(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
