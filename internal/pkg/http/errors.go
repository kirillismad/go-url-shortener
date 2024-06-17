package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/kirillismad/go-url-shortener/internal/pkg/usecase"
)

var (
	ErrReadBody      = errors.New("read body error")
	ErrJsonUnmarshal = errors.New("json unmarshal error")
)

func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	var errValidation *usecase.ErrValidation
	switch {
	case errors.As(err, &errValidation):
		WriteJson(ctx, w, http.StatusBadRequest, J{"msg": errValidation.Error()})
	case errors.Is(err, ErrReadBody):
		WriteJson(ctx, w, http.StatusBadRequest, J{"msg": err.Error()})
	case errors.Is(err, ErrJsonUnmarshal):
		WriteJson(ctx, w, http.StatusBadRequest, J{"msg": err.Error()})
	case errors.Is(err, usecase.ErrNoResult):
		WriteJson(ctx, w, http.StatusNotFound, J{"msg": "not found"})
	default:
		WriteJson(ctx, w, http.StatusInternalServerError, J{"msg": err.Error()})
	}
}
