package repo

import (
	"context"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
)

func (r *Repo) GetLinkByShortID(ctx context.Context, shortID string) (entity.Link, error) {
	l, err := r.q.GetLinkByShortID(ctx, shortID)
	if err != nil {
		return entity.Link{}, err
	}
	return entity.Link{
		ID:         l.ID,
		ShortID:    l.ShortID,
		Href:       l.Href,
		CreatedAt:  l.CreatedAt,
		UsageCount: l.UsageCount,
		UsageAt:    l.UsageAt,
	}, err
}
