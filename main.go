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

type PingServer struct {
	db *sql.DB
}

func (s PingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := s.db.PingContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}

	b, err := json.Marshal(map[string]string{"message": "pong"})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

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

	for i := 0; i < shortIDLen; i++ {
		idx := rand.Intn(len(Alphabet))
		b = append(b, Alphabet[idx])
	}
	return string(b)
}

func IsValidURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	return u.IsAbs() && (u.Scheme == "http" || u.Scheme == "https")
}

func CreateLinkHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("io.ReadAll(): %v", err)

			b, err2 := json.Marshal(map[string]string{"msg": msg})
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		var input LinkInput
		if err = json.Unmarshal(body, &input); err != nil {
			msg := fmt.Sprintf("json.Unmarshal(): %v", err)

			b, err2 := json.Marshal(map[string]string{"msg": msg})
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		if !IsValidURL(input.Href) {
			msg := fmt.Sprintf("Invalid link: %s", input.Href)

			b, err2 := json.Marshal(map[string]string{"msg": msg})
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			msg := fmt.Sprintf("db.Begin(): %v", err)

			b, err2 := json.Marshal(map[string]string{"msg": msg})
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}
		defer tx.Rollback()

		var shortID string
		for {
			shortID = generateShortID()
			var exists bool
			err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM "links" WHERE "short_id" = $1)`, shortID).Scan(&exists)
			if err != nil {
				msg := fmt.Sprintf("tx.QueryRow: %v", err)

				b, err2 := json.Marshal(map[string]string{"msg": msg})
				if err2 != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(b)
				return

			}
			if !exists {
				break
			}
		}

		query := `
			INSERT INTO "links" ("short_id", "href") 
			VALUES ($1, $2)
		`
		_, err = tx.ExecContext(ctx, query, shortID, input.Href)
		if err != nil {
			msg := fmt.Sprintf("tx.Exec(): %v", err)

			b, err2 := json.Marshal(map[string]string{"msg": msg})
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		if err = tx.Commit(); err != nil {
			msg := fmt.Sprintf("tx.Commit(): %v", err)

			b, err2 := json.Marshal(map[string]string{"msg": msg})
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		output := LinkOutput{ShortLink: "/s/" + shortID}
		content, err := json.Marshal(output)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")
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

	mux.Handle("GET /ping", PingServer{db: db})
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
