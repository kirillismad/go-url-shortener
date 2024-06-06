package http

import (
	"errors"
)

var (
	ErrReadBody      = errors.New("read body error")
	ErrJsonUnmarshal = errors.New("json unmarshal error")
	ErrJsonMarshal   = errors.New("json marshal error")
)
