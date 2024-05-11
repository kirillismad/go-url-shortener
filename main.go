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
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sethvargo/go-envconfig"
)

type J map[string]interface{}

func WriteJson(ctx context.Context, w http.ResponseWriter, status int, content any) {
	b, err := json.Marshal(content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func PingHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := db.PingContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		content := J{"msg": "pong"}
		WriteJson(r.Context(), w, http.StatusOK, content)
	}
}

type Link struct {
	ID         int64
	ShortID    string
	Href       string
	CreatedAt  time.Time
	UsageCount int64
	UsageAt    time.Time
}

type LinkInput struct {
	Href string `json:"href"`
}

type LinkOutput struct {
	ShortLink string `json:"shortLink"`
}

func isValidURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	return u.IsAbs() && (u.Scheme == "http" || u.Scheme == "https")
}

type TxOption func(*sql.TxOptions)

func WithIsolationLevel(level sql.IsolationLevel) TxOption {
	return func(opt *sql.TxOptions) {
		opt.Isolation = level
	}
}

func WithReadOnly(value bool) TxOption {
	return func(opt *sql.TxOptions) {
		opt.ReadOnly = value
	}
}

type CreateLinkHandler struct {
	db *sql.DB
}

func InTransaction(ctx context.Context, db *sql.DB, f func(*sql.Tx) error, opts ...TxOption) error {
	var txOptions *sql.TxOptions
	if len(opts) != 0 {
		txOptions = &sql.TxOptions{}
		for _, o := range opts {
			o(txOptions)
		}
	}

	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := f(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
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
		WriteJson(ctx, w, http.StatusBadRequest, J{"msg": msg})
		return
	}

	if !isValidURL(input.Href) {
		msg := fmt.Sprintf("Invalid link: %s", input.Href)
		WriteJson(ctx, w, http.StatusBadRequest, J{"msg": msg})
		return
	}

	var shortID string
	err = InTransaction(ctx, h.db, func(tx *sql.Tx) error {
		var err2 error
		shortID, err2 = h.getShortID(ctx, tx, input.Href)
		if err2 != nil {
			return err2
		}
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	output := LinkOutput{ShortLink: "/s/" + shortID}
	WriteJson(ctx, w, http.StatusOK, output)
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
		shortID, err = h.generateUniqueShortID(ctx, tx)
		if err != nil {
			return "", err
		}

		query2 := `
			INSERT INTO "links" ("short_id", "href") 
			VALUES ($1, $2)
		`
		_, err = tx.ExecContext(ctx, query2, shortID, href)
		if err != nil {
			return "", err
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

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatal(err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", PingHandler(db))
	mux.Handle("POST /new", &CreateLinkHandler{db: db})

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
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
