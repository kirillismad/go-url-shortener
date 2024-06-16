package usecase

import (
	"context"
	"errors"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
)

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

var ErrNoResult = errors.New("no result error")

type ErrValidation struct {
	message string
	err     error
}

func NewErrValidation(msg string, err error) ErrValidation {
	return ErrValidation{message: msg, err: err}
}

func (e ErrValidation) Error() string {
	return e.message
}

func (e ErrValidation) Unwrap() error {
	return e.err
}
