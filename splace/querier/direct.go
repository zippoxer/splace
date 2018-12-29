package querier

import (
	"context"
	"database/sql"
	"fmt"
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
	rows, err := d.db.QueryContext(ctx, query, args...)
	return &directRows{Rows: rows}, err
}

func (d *Direct) Engine() Engine {
	return d.engine
}

func (d *Direct) Database() string {
	return d.dbName
}

type directRows struct {
	*sql.Rows
	scanner *stringStringScan
}

func (r *directRows) ScanStrings() ([]string, error) {
	if r.scanner == nil {
		columns, err := r.Columns()
		if err != nil {
			return nil, err
		}
		r.scanner = newStringStringScan(columns)
	}
	err := r.scanner.update(r.Rows)
	if err != nil {
		return nil, err
	}
	return r.scanner.get(), nil
}

// stringStringScan scans an unknown amount of columns from sql.Rows
type stringStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      []string
	colCount int
	colNames []string
}

func newStringStringScan(columnNames []string) *stringStringScan {
	lenCN := len(columnNames)
	s := &stringStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make([]string, lenCN),
		colCount: lenCN,
		colNames: columnNames,
	}
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
	}
	return s
}

func (s *stringStringScan) update(rows *sql.Rows) error {
	if err := rows.Scan(s.cp...); err != nil {
		return err
	}
	j := 0
	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[j] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
		j++
	}
	return nil
}

func (s *stringStringScan) get() []string {
	return s.row
}
