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
	"regexp"

	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	validator10 "github.com/go-playground/validator/v10"
	common_http "github.com/kirillismad/go-url-shortener/internal/apps/common/http"
	links_http "github.com/kirillismad/go-url-shortener/internal/apps/links/http"
	"github.com/kirillismad/go-url-shortener/internal/apps/links/usecase"
	"github.com/kirillismad/go-url-shortener/internal/pkg/repo"
	"github.com/kirillismad/go-url-shortener/pkg/config"
)

type Config struct {
	Server struct {
		Host            string        `env:"HOST" yaml:"host" validate:"required"`
		Port            uint          `env:"PORT" yaml:"port" validate:"required"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" yaml:"shutdown_timeout" validate:"min=0s"`
		ReadTimeout     time.Duration `env:"READ_TIMEOUT" yaml:"read_timeout" validate:"min=0s"`
		WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" yaml:"write_timeout" validate:"min=0s"`
		IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" yaml:"idle_timeout" validate:"min=0s"`
	} `env:", prefix=SERVER_" yaml:"server" validate:"required"`
	DB struct {
		User     string `env:"USER, required" yaml:"user" validate:"required"`
		Password string `env:"PASSWORD, required" yaml:"password" validate:"required"`
		Host     string `env:"HOST, required" yaml:"host" validate:"required"`
		Port     uint   `env:"PORT, required" yaml:"port" validate:"required"`
		Name     string `env:"NAME, required" yaml:"name" validate:"required"`
		SSLMode  string `env:"SSLMODE" yaml:"sslmode" validate:"required"`
	} `env:", prefix=DB_" yaml:"db" validate:"required"`
	ShortID struct {
		Len      int    `env:"LEN, required" yaml:"len" validate:"min=8"`
		Alphabet string `env:"ALPHABET, required" yaml:"alphabet" validate:"required"`
	} `env:", prefix=SHORT_ID_" yaml:"short_id" validate:"required"`
}

func main() {
	ctx := context.Background()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getwd: %v", err)
	}
	log.Printf("Working directory: %s\n", workDir)

	// setup config
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

	// setup validator
	validator := validator10.New(validator10.WithRequiredStructEnabled())
	pattern := regexp.MustCompile(fmt.Sprintf(`^[%s]{%d}$`, cfg.ShortID.Alphabet, cfg.ShortID.Len))
	validator.RegisterValidation("short_id", func(fl validator10.FieldLevel) bool {
		return pattern.MatchString(fl.Field().String())
	})

	// setup db
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

	// setup repos
	linkRepoFactory := repo.NewRepoFactory(db, repo.NewLinkRepo)

	// setup mux
	mux := http.NewServeMux()
	mux.Handle("GET /ping", common_http.NewPingHandler().WithDB(db))
	mux.Handle(
		"POST /new",
		links_http.NewCreateLinkHandler(usecase.NewCreateLinkHandler(usecase.CreateLinkParams{
			RepoFactory: linkRepoFactory,
			Validator:   validator,
			ShortIDLen:  cfg.ShortID.Len,
			Alphabet:    []rune(cfg.ShortID.Alphabet),
		})),
	)
	mux.Handle(
		"GET /s/{short_id}",
		links_http.NewRedirectHandler(usecase.NewGetLinkByShortIDHandler(usecase.GetLinkByShortIDParams{
			RepoFactory: linkRepoFactory,
			Validator:   validator,
		})),
	)

	// setup server
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
