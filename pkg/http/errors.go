package http

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrReadBody      = errors.New("read body error")
	ErrJsonUnmarshal = errors.New("json unmarshal error")
	ErrJsonMarshal   = errors.New("json marshal error")
)

func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrReadBody):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, ErrJsonUnmarshal):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, ErrJsonMarshal):
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(err.Error()))
}
