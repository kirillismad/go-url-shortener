package http

import (
	"context"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
)

type TxFn func(IRepository) error

type IRepoFactory interface {
	GetRepo() IRepository
	InTransaction(context.Context, TxFn) error
}

type CreateLinkArgs struct {
	ShortID string
	Href    string
}

type IRepository interface {
	GetLinkByHref(context.Context, string) (entity.Link, error)
	CreateLink(context.Context, CreateLinkArgs) (entity.Link, error)
	IsLinkExistByShortID(context.Context, string) (bool, error)
	GetLinkByShortID(context.Context, string) (entity.Link, error)
	UpdateLinkUsageInfo(context.Context, int64) error
}

type TxFnG[RepoType any] func(RepoType) error

type IRepoFactoryG[R any] interface {
	GetRepo() R
	InTransaction(context.Context, TxFnG[R]) error
}
