package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
	sqlx "github.com/kirillismad/go-url-shortener/pkg/sql"
)

type CreateLinkInput struct {
	Href string `json:"href"`
}

type CreateLinkOutput struct {
	ShortLink string `json:"shortLink"`
}

type CreateLinkHandler struct {
	db *sql.DB
}

func NewCreateLinkHandler() *CreateLinkHandler {
	return new(CreateLinkHandler)
}

func (h *CreateLinkHandler) WithDB(db *sql.DB) *CreateLinkHandler {
	h.db = db
	return h
}

func (h *CreateLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var input CreateLinkInput
	if err := json.Unmarshal(body, &input); err != nil {
		msg := fmt.Sprintf("json.Unmarshal: %s", err)
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": msg})
		return
	}

	if !h.isValidURL(input.Href) {
		msg := fmt.Sprintf("Invalid link: %s", input.Href)
		httpx.WriteJson(ctx, w, http.StatusBadRequest, httpx.J{"msg": msg})
		return
	}

	var shortID string
	err = sqlx.InTransaction(ctx, h.db, func(tx *sql.Tx) error {
		var txErr error
		shortID, txErr = h.getShortID(ctx, tx, input.Href)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	output := CreateLinkOutput{ShortLink: "/s/" + shortID}
	httpx.WriteJson(ctx, w, http.StatusCreated, output)
}

func (*CreateLinkHandler) isValidURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	return u.IsAbs() && (u.Scheme == "http" || u.Scheme == "https")
}

func (h *CreateLinkHandler) generateShortID() string {
	const (
		shortIDLen = 11
		alphabet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789" + "-_"
	)

	Alphabet := []rune(alphabet)

	b := make([]rune, 0, shortIDLen)

	for i := 0; i < shortIDLen; i++ {
		idx := rand.Intn(len(Alphabet))
		b = append(b, Alphabet[idx])
	}
	return string(b)
}

func (h *CreateLinkHandler) getShortID(ctx context.Context, tx *sql.Tx, href string) (string, error) {
	var shortID string
	query := `
		SELECT "short_id" FROM "links" WHERE "href" = $1
	`
	err := tx.QueryRowContext(ctx, query, href).Scan(&shortID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		shortID, err2 := h.generateUniqueShortID(ctx, tx)
		if err2 != nil {
			return "", err2
		}

		query2 := `
			INSERT INTO "links" ("short_id", "href") 
			VALUES ($1, $2)
		`
		_, err2 = tx.ExecContext(ctx, query2, shortID, href)
		if err2 != nil {
			return "", err2
		}
	}
	return shortID, nil
}

func (h *CreateLinkHandler) generateUniqueShortID(ctx context.Context, tx *sql.Tx) (string, error) {
	for {
		shortID := h.generateShortID()

		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM "links" WHERE "short_id" = $1)`
		err := tx.QueryRowContext(ctx, query, shortID).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return shortID, nil
		}
	}
}
