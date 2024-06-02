package repo

import (
	"context"
	"database/sql"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/http"
	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type RepoFactory struct {
	db *sql.DB
}

func NewRepoFactory(db *sql.DB) http.RepoFactory {
	return &RepoFactory{db: db}
}

func (r *RepoFactory) GetRepo() http.Repository {
	return &Repository{q: sqlc.New(r.db)}
}

func (r *RepoFactory) InTransaction(ctx context.Context, txFn http.TxFn) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := txFn(&Repository{q: sqlc.New(tx)}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

type Repository struct {
	q *sqlc.Queries
}
