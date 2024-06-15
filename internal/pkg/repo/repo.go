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

func NewLinkRepo(q *sqlc.Queries) http.LinkRepo {
	return newRepo(q)
}
