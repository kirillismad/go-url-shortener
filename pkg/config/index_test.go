package config

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		r := require.New(t)

		type server struct {
			Host            string        `env:"HOST" yaml:"host" validate:"required"`
			Port            uint          `env:"PORT" yaml:"port" validate:"required"`
			ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" yaml:"shutdown_timeout" validate:"min=0s"`
			ReadTimeout     time.Duration `env:"READ_TIMEOUT" yaml:"read_timeout" validate:"min=0s"`
			WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" yaml:"write_timeout" validate:"min=0s"`
			IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" yaml:"idle_timeout" validate:"min=0s"`
		}

		type db struct {
			User     string `env:"USER, required" yaml:"user" validate:"required"`
			Password string `env:"PASSWORD, required" yaml:"password" validate:"required"`
			Host     string `env:"HOST, required" yaml:"host" validate:"required"`
			Port     uint   `env:"PORT, required" yaml:"port" validate:"required"`
			Name     string `env:"NAME, required" yaml:"name" validate:"required"`
			SSLMode  string `env:"SSLMODE" yaml:"sslmode" validate:"required"`
		}

		type config struct {
			Server server `env:", prefix=SERVER_" yaml:"server" validate:"required"`
			DB     db     `env:", prefix=DB_" yaml:"db" validate:"required"`
		}

		t.Setenv("DB_USER", "dbuser")
		t.Setenv("DB_PASSWORD", "dbpassword")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_NAME", "dbname")

		ctx := context.Background()
		cfg, err := GetConfig[config](ctx, "./testdata/test_config.yaml")

		r.NoError(err)

		exp := config{
			Server: server{
				Host:            "localhost",
				Port:            8000,
				ShutdownTimeout: 10 * time.Second,
				ReadTimeout:     0,
				WriteTimeout:    0,
				IdleTimeout:     0,
			},
			DB: db{
				User:     "dbuser",
				Password: "dbpassword",
				Host:     "localhost",
				Port:     5432,
				Name:     "dbname",
				SSLMode:  "disable",
			},
		}

		r.Equal(exp, cfg)
	})
}
