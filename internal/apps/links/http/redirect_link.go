package http

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
	sqlx "github.com/kirillismad/go-url-shortener/pkg/sql"
)

type RedirectHandler struct {
	db *sql.DB
}

func NewRedirectHandler() *RedirectHandler {
	return new(RedirectHandler)
}

func (h *RedirectHandler) WithDB(db *sql.DB) *RedirectHandler {
	h.db = db
	return h
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	short_id := r.PathValue("short_id")

	pattern := regexp.MustCompile(`^[a-zA-Z0-9\-_]{11}$`)
	if !pattern.MatchString(short_id) {
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": "Invalid link format"})
		return
	}

	var href string
	err := sqlx.InTransaction(ctx, h.db, func(tx *sql.Tx) error {
		query := `
			SELECT "id", "href" FROM "links" WHERE "short_id" = $1
		`
		var id int
		txErr := tx.QueryRowContext(ctx, query, short_id).Scan(&id, &href)
		if txErr != nil {
			return txErr
		}

		query = `
			UPDATE "links" SET "usage_count" = "usage_count" + 1, "usage_at" = NOW()
			WHERE "id" = $1
		`
		_, txErr = h.db.ExecContext(ctx, query, id)
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

	w.Header().Set("location", href)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
