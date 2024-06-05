package repo_factory

import (
	"context"
	"database/sql"

	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type RepoFactory[R any] struct {
	db       *sql.DB
	createFn func(q *sqlc.Queries) R
}

func NewRepoFactory[R any](db *sql.DB, createFn func(q *sqlc.Queries) R) *RepoFactory[R] {
	return &RepoFactory[R]{
		db:       db,
		createFn: createFn,
	}
}

func (r *RepoFactory[R]) GetRepo() R {
	return r.createFn(sqlc.New(r.db))
}

func (r *RepoFactory[R]) InTransaction(ctx context.Context, txFn func(R) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := txFn(r.createFn(sqlc.New(tx))); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
