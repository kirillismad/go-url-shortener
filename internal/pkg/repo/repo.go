package repo

import (
	"github.com/kirillismad/go-url-shortener/internal/apps/links/usecase"
	"github.com/kirillismad/go-url-shortener/internal/pkg/sqlc"
)

type Repo struct {
	q *sqlc.Queries
}

func newRepo(q *sqlc.Queries) *Repo {
	return &Repo{q: q}
}

func NewLinkRepo(q *sqlc.Queries) usecase.LinkRepo {
	return newRepo(q)
}
