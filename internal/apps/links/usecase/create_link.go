package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/validator"
)

type CreateLinkData struct {
	Href string
}

type CreateLinkResult struct {
	ShortLink string
}

type ICreateLinkHandler interface {
	Handle(ctx context.Context, data CreateLinkData) (CreateLinkResult, error)
}

type CreateLinkHandler struct {
	repoFactory LinkRepoFactory
}

func (h *CreateLinkHandler) Handle(ctx context.Context, data CreateLinkData) (CreateLinkResult, error) {
	if err := validator.StructCtx(ctx, &data); err != nil {
		// msg := fmt.Sprintf("Invalid link: %s", input.Href)
		// httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": msg})
		return CreateLinkResult{}, err
	}

	var link entity.Link
	err := h.repoFactory.InTransaction(ctx, func(r LinkRepo) error {
		var txErr error
		link, txErr = r.GetLinkByHref(ctx, data.Href)
		if txErr == nil {
			return nil
		}
		if !errors.Is(txErr, ErrNoResult) {
			return fmt.Errorf("r.GetLinkByHref: %w", txErr)
		}

		shortID, txErr := h.generateUniqueShortID(ctx, r)
		if txErr != nil {
			return fmt.Errorf("h.generateUniqueShortID: %w", txErr)
		}

		link, txErr = r.CreateLink(ctx, CreateLinkArgs{
			ShortID: shortID,
			Href:    data.Href,
		})
		if txErr != nil {
			return fmt.Errorf("r.CreateLink: %w", txErr)
		}
		return txErr
	})
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		return CreateLinkResult{}, err
	}
	return CreateLinkResult{ShortLink: "/s/" + link.ShortID}, nil
}

func (h *CreateLinkHandler) generateShortID() string {
	const (
		shortIDLen = 11
		alphabet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789" + "-_"
	)

	alph := []rune(alphabet)

	b := make([]rune, 0, shortIDLen)
	for i := 0; i < shortIDLen; i++ {
		idx := rand.Intn(len(alph))
		b = append(b, alph[idx])
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
