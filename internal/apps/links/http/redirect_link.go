package http

import (
	"errors"
	"net/http"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/repo_factory"
	"github.com/kirillismad/go-url-shortener/internal/pkg/validator"
	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
)

type RedirectHandler struct {
	repoFactory *repo_factory.RepoFactory[IRedirectHandlerRepo]
}

func NewRedirectHandler() *RedirectHandler {
	return new(RedirectHandler)
}

func (h *RedirectHandler) WithRepoFactory(repoFactory *repo_factory.RepoFactory[IRedirectHandlerRepo]) *RedirectHandler {
	h.repoFactory = repoFactory
	return h
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	short_id := r.PathValue("short_id")

	// usecase start
	if err := validator.VarCtx(ctx, short_id, "short_id"); err != nil {
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": "Invalid link format"})
		return
	}

	var link entity.Link
	err := h.repoFactory.InTransaction(ctx, func(r IRedirectHandlerRepo) error {
		var txErr error
		link, txErr = r.GetLinkByShortID(ctx, short_id)
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		httpx.WriteJson(ctx, w, http.StatusNotFound, httpx.J{"msg": "Page not found"})
		return
	}

	// usecase end

	w.Header().Set("location", link.Href)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
