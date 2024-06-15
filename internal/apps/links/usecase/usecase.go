package usecase

import (
	"context"
	"errors"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
)

var ErrNoResult = errors.New("no result error")

type LinkRepoFactory interface {
	GetRepo() LinkRepo
	InTransaction(ctx context.Context, txFn func(r LinkRepo) error) error
}

type CreateLinkArgs struct {
	ShortID string
	Href    string
}

type LinkRepo interface {
	CreateLink(context.Context, CreateLinkArgs) (entity.Link, error)
	GetLinkByHref(context.Context, string) (entity.Link, error)
	IsLinkExistByShortID(context.Context, string) (bool, error)
	GetLinkByShortID(context.Context, string) (entity.Link, error)
	UpdateLinkUsageInfo(context.Context, int64) error
}
