package querier

import (
	"context"
	"database/sql"
)

type Direct struct {
	db     *sql.DB
	dbName string
	engine Engine
}

func NewDirect(dbName string, engine Engine, db *sql.DB) *Direct {
	return &Direct{
		db:     db,
		dbName: dbName,
		engine: engine,
	}
}

func (d *Direct) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *Direct) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *Direct) Engine() Engine {
	return d.engine
}

func (d *Direct) Database() string {
	return d.dbName
}
