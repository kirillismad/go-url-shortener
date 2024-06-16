package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func ReadJson[T any](ctx context.Context, r *http.Request) (T, error) {
	var zero T
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return zero, errors.Join(ErrReadBody, err)
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return zero, errors.Join(ErrJsonUnmarshal, err)
	}
	return result, nil
}

func WriteJson[T any](ctx context.Context, w http.ResponseWriter, status int, content T) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(content)
}
