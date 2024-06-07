package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/apps/links/http"
)

func (r *Repo) GetLinkByHref(ctx context.Context, href string) (entity.Link, error) {
	l, err := r.q.GetLinkByHref(ctx, href)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Link{}, errors.Join(http.ErrNoResult, err)
		}
		return entity.Link{}, err
	}
	return entity.Link{
		ID:         l.ID,
		ShortID:    l.ShortID,
		Href:       l.Href,
		CreatedAt:  l.CreatedAt,
		UsageCount: l.UsageCount,
		UsageAt:    l.UsageAt,
	}, nil
}
