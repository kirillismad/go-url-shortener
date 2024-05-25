package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sethvargo/go-envconfig"

	common_http "github.com/kirillismad/go-url-shortener/internal/common/http"
	httpx "github.com/kirillismad/go-url-shortener/pkg/http"
	sqlx "github.com/kirillismad/go-url-shortener/pkg/sql"
)

type LinkInput struct {
	Href string `json:"href"`
}

type LinkOutput struct {
	ShortLink string `json:"shortLink"`
}

type CreateLinkHandler struct {
	db *sql.DB
}

func (h *CreateLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var input LinkInput
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

	output := LinkOutput{ShortLink: "/s/" + shortID}
	httpx.WriteJson(ctx, w, http.StatusCreated, output)
}

func (_ *CreateLinkHandler) isValidURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	return u.IsAbs() && (u.Scheme == "http" || u.Scheme == "https")
}

var Alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789" + "-_")

const shortIDLen = 11

func (h *CreateLinkHandler) generateShortID() string {
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

type RedirectHandler struct {
	db *sql.DB
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

type Config struct {
	Server struct {
		Host            string        `env:"HOST, required"`
		Port            uint          `env:"PORT, required"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT, default=10s"`
		ReadTimeout     time.Duration `env:"READ_TIMEOUT, default=0s"`
		WriteTimeout    time.Duration `env:"WRITE_TIMEOUT, default=0s"`
		IdleTimeout     time.Duration `env:"IDLE_TIMEOUT, default=0s"`
	} `env:", prefix=SERVER_"`
	DB struct {
		User     string `env:"USER, required"`
		Password string `env:"PASSWORD, required"`
		Host     string `env:"HOST, required"`
		Port     uint   `env:"PORT, required"`
		Name     string `env:"NAME, required"`
		SSLMode  string `env:"SSLMODE, default=disable"`
	} `env:", prefix=DB_"`
}

func main() {
	ctx := context.Background()

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("envconfig.Process: %v", err)
	}

	v := make(url.Values, 1)
	v.Set("sslmode", cfg.DB.SSLMode)
	connString := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:     cfg.DB.Host,
		Path:     cfg.DB.Name,
		RawQuery: v.Encode(),
	}

	db, err := sql.Open("pgx", connString.String())
	if err != nil {
		log.Fatal(err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("GET /ping", common_http.NewPingHandler().WithDB(db))
	mux.Handle("POST /new", &CreateLinkHandler{db: db})
	mux.Handle("GET /s/{short_id}", &RedirectHandler{db: db})

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Println("Server is starting")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}

		log.Println("Server stops serving new connections")
	}()

	// graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	ctx, release := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer release()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
}
