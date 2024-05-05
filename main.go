package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK\n"))
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

	// docker run --name url_shortener_db --rm -p 5432:5432 -e POSTGRES_PASSWORD=dbpassword -e POSTGRES_USER=dbuser -e POSTGRES_DB=dbname postgres:16
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
