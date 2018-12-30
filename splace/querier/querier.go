package querier

import (
	"context"
	"io"
)

type Engine int

const (
	MySQL Engine = iota
	PostgreSQL
	SQLServer
	Oracle
)

func (e Engine) String() string {
	switch e {
	case MySQL:
		return "Mysql"
	case PostgreSQL:
		return "PostgreSQL"
	case SQLServer:
		return "SQLServer"
	case Oracle:
		return "Oracle"
	}
	return ""
}

type Querier interface {
	Engine() Engine

	// Database returns the current database name.
	Database() string

	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)

	Dump(ctx context.Context, w io.Writer) error
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
