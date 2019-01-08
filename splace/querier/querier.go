package querier

import (
	"context"
	"errors"
	"io"
)

var (
	ErrUnsupportedEngine = errors.New("database engine is not supported")
)

type Engine string

const (
	MySQL      Engine = "mysql"
	PostgreSQL Engine = "postgres"
	SQLServer  Engine = "sqlserver"
	Oracle     Engine = "oracle"
)

type Querier interface {
	Config() Config

	// DiscoveredConfigs returns the list of available database connections
	// discovered by the querier.
	DiscoveredConfigs() []DiscoveredConfig

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

type DiscoveredConfig struct {
	// Who specifies the name of the CMS or framework.
	Who string

	// Where specifies the filename, environment variable name or
	// wherever else the config was discovered.
	Where string

	Config Config
}
