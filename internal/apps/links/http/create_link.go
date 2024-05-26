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
	"github.com/kirillismad/go-url-shortener/pkg/sqlc"
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
		fmt.Fprintln(w, err.Error())
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
		shortID, txErr = h.getShortID(ctx, sqlc.New(tx), input.Href)
		if txErr != nil {
			return fmt.Errorf("h.getShortID: %w", txErr)
		}
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
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

func (h *CreateLinkHandler) getShortID(ctx context.Context, q *sqlc.Queries, href string) (string, error) {
	link, err := q.GetLinkByHref(ctx, href)
	if err == nil {
		return link.ShortID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("qtx.GetLinkByHref: %w", err)
	}

	shortID, err := h.generateUniqueShortID(ctx, q)
	if err != nil {
		return "", fmt.Errorf("h.generateUniqueShortID: %w", err)
	}

	link, err = q.CreateLink(ctx, sqlc.CreateLinkParams{ShortID: shortID, Href: href})
	if err != nil {
		return "", fmt.Errorf("qtx.CreateLink: %w", err)
	}
	return link.ShortID, nil
}

func (h *CreateLinkHandler) generateUniqueShortID(ctx context.Context, q *sqlc.Queries) (string, error) {
	for {
		shortID := h.generateShortID()
		exists, err := q.IsLinkExistByShortID(ctx, shortID)
		if err != nil {
			return "", fmt.Errorf("qtx.IsLinkExistByShortID: %w", err)
		}
		if !exists {
			return shortID, nil
		}
	}
}
