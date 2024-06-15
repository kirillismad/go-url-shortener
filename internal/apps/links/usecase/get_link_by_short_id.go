package usecase

import (
	"context"
	"errors"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/validator"
)

type GetLinkByShortIDData struct {
	ShortID string
}

type GetLinkByShortIDResult struct {
	Href string
}

type IGetLinkByShortIDHandler interface {
	Handle(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error)
}

type GetLinkByShortIDHandler struct {
	repoFactory LinkRepoFactory
}

func (h *GetLinkByShortIDHandler) Handle(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error) {
	if err := validator.VarCtx(ctx, data.ShortID, "short_id"); err != nil {
		// httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": "Invalid link format"})
		return GetLinkByShortIDResult{}, err
	}

	var link entity.Link
	err := h.repoFactory.InTransaction(ctx, func(r LinkRepo) error {
		var txErr error
		link, txErr = r.GetLinkByShortID(ctx, data.ShortID)
		if txErr != nil {
			return txErr
		}

		txErr = r.UpdateLinkUsageInfo(ctx, link.ID)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		if !errors.Is(err, ErrNoResult) {
			// w.WriteHeader(http.StatusInternalServerError)
			return GetLinkByShortIDResult{}, err
		}
		return GetLinkByShortIDResult{}, err
		// httpx.WriteJson(ctx, w, http.StatusNotFound, httpx.J{"msg": "Page not found"})
	}
	return GetLinkByShortIDResult{Href: link.Href}, nil
}
