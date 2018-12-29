package querier

import "context"

type Engine int

const (
	MySQL Engine = iota
	PostgreSQL
	SQLServer
	Oracle
)

type Querier interface {
	Engine() Engine

	// Database returns the current database name.
	Database() string

	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
}

type Result interface {
	RowsAffected() (int64, error)
}

type Rows interface {
	Columns() ([]string, error)
	Next() bool
	ScanStrings() ([]string, error)
	Err() error
	Close() error
}
