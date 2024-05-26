package http

import (
	"database/sql"
	"net/http"

	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
)

type PingHandler struct {
	db *sql.DB
}

func NewPingHandler() *PingHandler {
	return new(PingHandler)
}

func (h *PingHandler) WithDB(db *sql.DB) *PingHandler {
	h.db = db
	return h
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.db.PingContext(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	content := httpx.J{"msg": "pong"}
	httpx.WriteJson(ctx, w, http.StatusOK, content)
}
