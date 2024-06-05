package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"

	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	common_http "github.com/kirillismad/go-url-shortener/internal/apps/common/http"
	links_http "github.com/kirillismad/go-url-shortener/internal/apps/links/http"
	"github.com/kirillismad/go-url-shortener/internal/pkg/repo"
	"github.com/kirillismad/go-url-shortener/internal/pkg/repo_factory"
	"github.com/kirillismad/go-url-shortener/pkg/config"
)

type Config struct {
	Server struct {
		Host            string        `env:"HOST" yaml:"host"`
		Port            uint          `env:"PORT" yaml:"port"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" yaml:"shutdown_timeout"`
		ReadTimeout     time.Duration `env:"READ_TIMEOUT" yaml:"read_timeout"`
		WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" yaml:"write_timeout"`
		IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" yaml:"idle_timeout"`
	} `env:", prefix=SERVER_" yaml:"server" validate:"required"`
	DB struct {
		User     string `env:"USER, required" yaml:"user"`
		Password string `env:"PASSWORD, required" yaml:"password"`
		Host     string `env:"HOST, required" yaml:"host"`
		Port     uint   `env:"PORT, required" yaml:"port"`
		Name     string `env:"NAME, required" yaml:"name"`
		SSLMode  string `env:"SSLMODE" yaml:"sslmode"`
	} `env:", prefix=DB_" yaml:"db" validate:"required"`
}

func main() {
	ctx := context.Background()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getwd: %v", err)
	}
	log.Printf("Working directory: %s\n", workDir)

	var configPath string
	defaultConfigPath := filepath.Join(workDir, "config", "local.yaml")
	flag.StringVar(&configPath, "config", defaultConfigPath, "config path")
	flag.StringVar(&configPath, "c", defaultConfigPath, "config path")
	flag.Parse()

	cfg, err := config.GetConfig[Config](ctx, configPath)
	if err != nil {
		log.Fatalf("config.GetConfig: %v", err)
	}
	log.Printf("Configuration file: %s\n", configPath)

	v := make(url.Values, 1)
	v.Set("sslmode", cfg.DB.SSLMode)
	connString := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.DB.Host, cfg.DB.Port),
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
	mux.Handle(
		"POST /new",
		links_http.NewCreateLinkHandler().WithRepoFactory(repo_factory.NewRepoFactory(db, repo.NewCreateLinkRepo)),
	)
	mux.Handle(
		"GET /s/{short_id}",
		links_http.NewRedirectHandler().WithRepoFactory(repo_factory.NewRepoFactory(db, repo.NewRedirectHandlerRepo)),
	)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("Server is starting\n")
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
