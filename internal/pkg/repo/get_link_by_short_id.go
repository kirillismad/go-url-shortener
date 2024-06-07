package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/apps/links/http"
)

func (r *Repo) GetLinkByShortID(ctx context.Context, shortID string) (entity.Link, error) {
	l, err := r.q.GetLinkByShortID(ctx, shortID)
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
	}, err
}
