package repo

import (
	"github.com/kirillismad/go-url-shortener/internal/apps/links/http"
	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type Repo struct {
	q *sqlc.Queries
}

func newRepo(q *sqlc.Queries) *Repo {
	return &Repo{q: q}
}

func NewCreateLinkRepo(q *sqlc.Queries) http.ICreateLinkRepo {
	return newRepo(q)
}

func NewRedirectHandlerRepo(q *sqlc.Queries) http.IRedirectHandlerRepo {
	return newRepo(q)
}
