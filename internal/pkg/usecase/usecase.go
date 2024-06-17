package usecase

import "context"

type RepoFactory[T any] interface {
	GetRepo() T
	InTransaction(ctx context.Context, txFn func(r T) error) error
}
