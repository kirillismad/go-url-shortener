package config

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func GetConfig[T any](path string) (T, error) {
	var zero, config T

	b, err := os.ReadFile(path)
	if err != nil {
		return zero, fmt.Errorf("os.ReadFile: %w", err)
	}

	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return zero, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	if err := envconfig.Process(context.Background(), &config); err != nil {
		return zero, fmt.Errorf("envconfig.Process: %w", err)
	}

	err = validate.Struct(&config)
	if err != nil {
		return zero, fmt.Errorf("validate.StructCtx: %w", err)
	}

	return config, nil
}
