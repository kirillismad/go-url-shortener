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
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sethvargo/go-envconfig"
)

func PingHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := db.PingContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Internal Server Error")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "200 OK")
	}
}

// CREATE TABLE IF NOT EXISTS "links" (
// 	"id" bigint GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE,
// 	"short_id" text NOT NULL UNIQUE,
// 	"href" text NOT NULL UNIQUE,
// 	"created_at" timestamp with time zone NOT NULL,
// 	"usage_count" timestamp with time zone NOT NULL,
// 	"usage_at" timestamp with time zone NOT NULL,
// 	PRIMARY KEY ("id")
// );

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

var Alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789" + "-_")

const shortIDLen = 11

func generateShortID() string {
	b := make([]rune, 0, shortIDLen)

	alphabetLen := len(Alphabet)
	for i := 0; i < shortIDLen; i++ {
		b = append(b, Alphabet[rand.Intn(alphabetLen)])
	}
	return string(b)
}

func CreateLinkHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	const query = `
		INSERT INTO "links" ("short_id", "href") 
		VALUES ($1, $2)
	`
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "io.ReadAll(): %v", err)
			return
		}

		var input LinkInput
		if err = json.Unmarshal(body, &input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "json.Unmarshal(): %v", err)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		shortID := generateShortID()
		_, err = db.Exec(query, shortID, input.Href)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "db.Exec(): %v", err)
			return
		}

		if err = tx.Commit(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "db.Begin(): %v", err)
			return
		}

		content, err := json.Marshal(LinkOutput{ShortLink: "/s/" + shortID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(content)
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
	mux.HandleFunc("POST /new", CreateLinkHandler(db))

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
