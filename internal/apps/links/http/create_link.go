package http

import (
	"net/http"

	"github.com/kirillismad/go-url-shortener/internal/apps/links/usecase"
	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
)

type CreateLinkInput struct {
	Href string `json:"href" validate:"http_url"`
}

type CreateLinkOutput struct {
	ShortLink string `json:"shortLink"`
}

type CreateLinkHandler struct {
	usecase usecase.ICreateLinkHandler
}

func NewCreateLinkHandler(usecase usecase.ICreateLinkHandler) *CreateLinkHandler {
	h := new(CreateLinkHandler)
	h.usecase = usecase
	return h
}

func (h *CreateLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httpx.ReadJson[CreateLinkInput](ctx, r)
	if err != nil {
		httpx.HandleError(ctx, w, err)
		return
	}

	result, err := h.usecase.Handle(ctx, usecase.CreateLinkData{
		Href: input.Href,
	})
	if err != nil {
		httpx.HandleError(ctx, w, err)
		return
	}

	output := CreateLinkOutput{ShortLink: "/s/" + result.ShortID}
	httpx.WriteJson(ctx, w, http.StatusCreated, output)
}
