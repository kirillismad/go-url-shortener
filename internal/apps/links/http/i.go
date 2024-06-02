package http

import (
	"context"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
)

type TxFn func(Repository) error

type RepoFactory interface {
	GetRepo() Repository
	InTransaction(context.Context, TxFn) error
}

type CreateLinkArgs struct {
	ShortID string
	Href    string
}

type Repository interface {
	GetLinkByHref(context.Context, string) (entity.Link, error)
	CreateLink(context.Context, CreateLinkArgs) (entity.Link, error)
	IsLinkExistByShortID(context.Context, string) (bool, error)
	GetLinkByShortID(context.Context, string) (entity.Link, error)
	UpdateLinkUsageInfo(context.Context, int64) error
}
