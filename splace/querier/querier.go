package querier

import (
	"context"
	"errors"
	"io"
)

var (
	ErrUnsupportedEngine = errors.New("database engine is not supported")
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
		return "mysql"
	case PostgreSQL:
		return "postgres"
	case SQLServer:
		return "sqlserver"
	case Oracle:
		return "oracle"
	}
	return ""
}

type Querier interface {
	Config() Config

	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)

	Dump(ctx context.Context, w io.Writer) error

	Close() error
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
