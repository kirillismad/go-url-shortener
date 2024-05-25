package http

import (
	"context"
	"encoding/json"
	"net/http"
)

func WriteJson[T any](ctx context.Context, w http.ResponseWriter, status int, content T) {
	b, err := json.Marshal(content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}
