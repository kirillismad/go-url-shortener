package usecase

import (
	"context"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
)

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
