package http

import (
	"net/http"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/usecase"
	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
)

type RedirectHandler struct {
	usecase usecase.IGetLinkByShortIDHandler
}

func NewRedirectHandler(usecase usecase.IGetLinkByShortIDHandler) *RedirectHandler {
	h := new(RedirectHandler)
	h.usecase = usecase
	return h
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	short_id := r.PathValue("short_id")

	result, err := h.usecase.Handle(ctx, usecase.GetLinkByShortIDData{
		ShortID: short_id,
	})
	if err != nil {
		httpx.HandleError(ctx, w, err)
		return
	}

	w.Header().Set("location", result.Href)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
