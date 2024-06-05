package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/repo_factory"
	"github.com/kirillismad/go-url-shortener/internal/pkg/validator"
	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
)

type CreateLinkArgs struct {
	ShortID string
	Href    string
}

type ICreateLinkRepo interface {
	GetLinkByHref(context.Context, string) (entity.Link, error)
	CreateLink(context.Context, CreateLinkArgs) (entity.Link, error)
	IsLinkExistByShortID(context.Context, string) (bool, error)
}

type CreateLinkInput struct {
	Href string `json:"href" validate:"http_url"`
}

type CreateLinkOutput struct {
	ShortLink string `json:"shortLink"`
}

type CreateLinkHandler struct {
	repoFactory *repo_factory.RepoFactory[ICreateLinkRepo]
}

func NewCreateLinkHandler() *CreateLinkHandler {
	return new(CreateLinkHandler)
}

func (h *CreateLinkHandler) WithRepoFactory(repoFactory *repo_factory.RepoFactory[ICreateLinkRepo]) *CreateLinkHandler {
	h.repoFactory = repoFactory
	return h
}

func (h *CreateLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "io.ReadAll: %v\n", err)
		return
	}

	var input CreateLinkInput
	if err := json.Unmarshal(body, &input); err != nil {
		msg := fmt.Sprintf("json.Unmarshal: %s", err)
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": msg})
		return
	}

	if err := validator.StructCtx(ctx, &input); err != nil {
		msg := fmt.Sprintf("Invalid link: %s", input.Href)
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": msg})
		return
	}

	var link entity.Link
	err = h.repoFactory.InTransaction(ctx, func(r ICreateLinkRepo) error {
		var txErr error
		link, txErr = r.GetLinkByHref(ctx, input.Href)
		if txErr == nil {
			return nil
		}
		if !errors.Is(txErr, sql.ErrNoRows) {
			return fmt.Errorf("r.GetLinkByHref: %w", err)
		}

		shortID, txErr := h.generateUniqueShortID(ctx, r)
		if txErr != nil {
			return fmt.Errorf("h.generateUniqueShortID: %w", err)
		}

		link, txErr = r.CreateLink(ctx, CreateLinkArgs{
			ShortID: shortID,
			Href:    input.Href,
		})
		if txErr != nil {
			return fmt.Errorf("r.CreateLink: %w", err)
		}
		return txErr
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "h.repoFactory.InTransaction: %v\n", err)
		return
	}

	output := CreateLinkOutput{ShortLink: "/s/" + link.ShortID}
	httpx.WriteJson(ctx, w, http.StatusCreated, output)
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

func (h *CreateLinkHandler) generateUniqueShortID(ctx context.Context, repo ICreateLinkRepo) (string, error) {
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
