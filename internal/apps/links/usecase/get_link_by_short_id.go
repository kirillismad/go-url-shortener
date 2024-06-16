package usecase

import (
	"context"

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

func NewGetLinkByShortIDHandler(repoFactory LinkRepoFactory) IGetLinkByShortIDHandler {
	h := new(GetLinkByShortIDHandler)
	h.repoFactory = repoFactory
	return h
}

func (h *GetLinkByShortIDHandler) Handle(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error) {
	if err := validator.VarCtx(ctx, data.ShortID, "short_id"); err != nil {
		return GetLinkByShortIDResult{}, NewErrValidation("Invalid link format", err)
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
		return GetLinkByShortIDResult{}, err
	}
	return GetLinkByShortIDResult{Href: link.Href}, nil
}
