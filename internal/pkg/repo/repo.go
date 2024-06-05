package repo

import (
	"github.com/kirillismad/go-url-shortener/internal/apps/links/http"
	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type Repository struct {
	q *sqlc.Queries
}

func newRepo(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func NewCreateLinkRepository(q *sqlc.Queries) http.ICreateLinkRepo {
	return newRepo(q)
}

func NewRedirectHandlerRepo(q *sqlc.Queries) http.IRedirectHandlerRepo {
	return newRepo(q)
}
