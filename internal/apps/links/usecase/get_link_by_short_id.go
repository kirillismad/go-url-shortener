package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/usecase"
)

type GetLinkByShortIDData struct {
	ShortID string `validate:"required,short_id"`
}

type GetLinkByShortIDResult struct {
	Href string
}

type IGetLinkByShortIDHandler interface {
	Handle(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error)
}

type GetLinkByShortIDHandler struct {
	repoFactory usecase.RepoFactory[LinkRepo]
	validator   *validator.Validate
}

type GetLinkByShortIDParams struct {
	RepoFactory usecase.RepoFactory[LinkRepo]
	Validator   *validator.Validate
}

func NewGetLinkByShortIDHandler(params GetLinkByShortIDParams) IGetLinkByShortIDHandler {
	return &GetLinkByShortIDHandler{
		repoFactory: params.RepoFactory,
		validator:   params.Validator,
	}
}

func (h *GetLinkByShortIDHandler) Handle(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error) {
	if err := h.validator.StructCtx(ctx, data); err != nil {
		return GetLinkByShortIDResult{}, usecase.NewErrValidation("Invalid link format", err)
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
