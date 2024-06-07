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

func WriteJson[T any](ctx context.Context, w http.ResponseWriter, status int, content T) error {
	b, err := json.Marshal(content)
	if err != nil {
		return errors.Join(ErrJsonMarshal, err)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
	return nil
}
