package repo

import (
	"context"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
)

type CreateLinkArgs struct {
	ShortID string
	Href    string
}

func (r *Repository) CreateLink(ctx context.Context, args CreateLinkArgs) (entity.Link, error) {
	p := sqlc.CreateLinkParams{
		ShortID: args.ShortID,
		Href:    args.Href,
	}
	l, err := r.q.CreateLink(ctx, p)
	if err != nil {
		return entity.Link{}, err
	}

	e := entity.Link{
		ID:         l.ID,
		ShortID:    l.ShortID,
		Href:       l.Href,
		CreatedAt:  l.CreatedAt,
		UsageCount: l.UsageCount,
		UsageAt:    l.UsageAt,
	}
	return e, nil
}
