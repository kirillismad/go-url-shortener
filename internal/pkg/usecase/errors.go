package usecase

import "errors"

var ErrNoResult = errors.New("no result error")

type ErrValidation struct {
	message string
	err     error
}

func NewErrValidation(msg string, err error) ErrValidation {
	return ErrValidation{message: msg, err: err}
}

func (e ErrValidation) Error() string {
	return e.message
}

func (e ErrValidation) Unwrap() error {
	return e.err
}
