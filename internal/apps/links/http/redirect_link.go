package http

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/entity"
	"github.com/kirillismad/go-url-shortener/internal/pkg/validator"
	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
)

type RedirectHandler struct {
	repoFactory RepoFactory
}

func NewRedirectHandler() *RedirectHandler {
	return new(RedirectHandler)
}

func (h *RedirectHandler) WithRepoFactory(repoFactory RepoFactory) *RedirectHandler {
	h.repoFactory = repoFactory
	return h
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	short_id := r.PathValue("short_id")

	if err := validator.VarCtx(ctx, short_id, "short_id"); err != nil {
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": "Invalid link format"})
		return
	}

	var link entity.Link
	err := h.repoFactory.InTransaction(ctx, func(r Repository) error {
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
		if !errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		httpx.WriteJson(ctx, w, http.StatusNotFound, httpx.J{"msg": "Page not found"})
		return
	}

	w.Header().Set("location", link.Href)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
