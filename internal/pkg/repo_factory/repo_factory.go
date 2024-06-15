package repo_factory

import (
	"context"
	"database/sql"

	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type NewRepoFn[R any] func(q *sqlc.Queries) R

type RepoFactory[R any] struct {
	db        *sql.DB
	newRepoFn NewRepoFn[R]
}

func NewRepoFactory[R any](db *sql.DB, newRepoFn NewRepoFn[R]) *RepoFactory[R] {
	return &RepoFactory[R]{
		db:        db,
		newRepoFn: newRepoFn,
	}
}

func (r *RepoFactory[R]) GetRepo() R {
	return r.newRepoFn(sqlc.New(r.db))
}

func (r *RepoFactory[R]) InTransaction(ctx context.Context, txFn func(R) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := txFn(r.newRepoFn(sqlc.New(tx))); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
