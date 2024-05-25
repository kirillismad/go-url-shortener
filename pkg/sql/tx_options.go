package sql

import "database/sql"

type TxOption func(*sql.TxOptions)

func WithIsolationLevel(level sql.IsolationLevel) TxOption {
	return func(opt *sql.TxOptions) {
		opt.Isolation = level
	}
}

func WithReadOnly(value bool) TxOption {
	return func(opt *sql.TxOptions) {
		opt.ReadOnly = value
	}
}
