package validator

import (
	"context"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func init() {
	v = validator.New(validator.WithRequiredStructEnabled())

	v.RegisterValidation("short_id", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[a-zA-Z0-9\-_]{11}$`).MatchString(fl.Field().String())
	})
}

func StructCtx(ctx context.Context, s interface{}) error {
	return v.StructCtx(ctx, s)
}

func VarCtx(ctx context.Context, field interface{}, tag string) error {
	return v.VarCtx(ctx, field, tag)
}
