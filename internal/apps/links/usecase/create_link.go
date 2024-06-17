package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/go-playground/validator/v10"
	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/usecase"
)

type CreateLinkData struct {
	Href string `validate:"required,http_url"`
}

type CreateLinkResult struct {
	ShortID string
}

type ICreateLinkHandler interface {
	Handle(ctx context.Context, data CreateLinkData) (CreateLinkResult, error)
}

type CreateLinkHandler struct {
	repoFactory usecase.RepoFactory[LinkRepo]
	validator   *validator.Validate
	shortIDLen  int
	alphabet    []rune
}

type CreateLinkParams struct {
	RepoFactory usecase.RepoFactory[LinkRepo]
	Validator   *validator.Validate
	ShortIDLen  int
	Alphabet    []rune
}

func NewCreateLinkHandler(params CreateLinkParams) ICreateLinkHandler {
	return &CreateLinkHandler{
		repoFactory: params.RepoFactory,
		validator:   params.Validator,
		shortIDLen:  params.ShortIDLen,
		alphabet:    params.Alphabet,
	}
}

func (h *CreateLinkHandler) Handle(ctx context.Context, data CreateLinkData) (CreateLinkResult, error) {
	if err := h.validator.StructCtx(ctx, &data); err != nil {
		return CreateLinkResult{}, usecase.NewErrValidation("Invalid request", err)
	}

	var link entity.Link
	err := h.repoFactory.InTransaction(ctx, func(repo LinkRepo) error {
		var txErr error
		link, txErr = repo.GetLinkByHref(ctx, data.Href)
		if txErr == nil {
			return nil
		}
		if !errors.Is(txErr, usecase.ErrNoResult) {
			return fmt.Errorf("repo.GetLinkByHref: %w", txErr)
		}

		shortID, txErr := h.generateUniqueShortID(ctx, repo)
		if txErr != nil {
			return txErr
		}

		link, txErr = repo.CreateLink(ctx, CreateLinkArgs{
			ShortID: shortID,
			Href:    data.Href,
		})
		if txErr != nil {
			return fmt.Errorf("repo.CreateLink: %w", txErr)
		}
		return txErr
	})
	if err != nil {
		return CreateLinkResult{}, err
	}
	return CreateLinkResult{ShortID: link.ShortID}, nil
}

func (h *CreateLinkHandler) generateShortID() string {
	b := make([]rune, 0, h.shortIDLen)
	for i := 0; i < h.shortIDLen; i++ {
		idx := rand.Intn(len(h.alphabet))
		b = append(b, h.alphabet[idx])
	}
	return string(b)
}

func (h *CreateLinkHandler) generateUniqueShortID(ctx context.Context, repo LinkRepo) (string, error) {
	for {
		shortID := h.generateShortID()
		exists, err := repo.IsLinkExistByShortID(ctx, shortID)
		if err != nil {
			return "", fmt.Errorf("repo.IsLinkExistByShortID: %w", err)
		}
		if !exists {
			return shortID, nil
		}
	}
}
