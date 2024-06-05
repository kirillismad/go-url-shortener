package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJson[T any](ctx context.Context, w http.ResponseWriter, status int, content T) {
	b, err := json.Marshal(content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "json.Marshal: %v\n", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}
