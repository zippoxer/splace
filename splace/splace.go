package splace

import (
	"context"
	"splace/splace/querier"
)

type Mode int

const (
	Equals Mode = iota
	Like
	Regexp
)

type Splace struct {
	db querier.Querier
}

func New(db querier.Querier) *Splace {
	return &Splace{db: db}
}

func (s *Splace) Replace(ctx context.Context, opt ReplaceOptions) *Replacer {
	r := &Replacer{ctx: ctx, db: s.db, opt: opt}
	r.start()
	return r
}

func (s *Splace) Tables(ctx context.Context) (map[string][]string, error) {
	rows, err := s.db.Query(ctx, `SELECT TABLE_NAME, COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA = ?`, s.db.Database())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := make(map[string][]string)
	var tableName, columnName string
	for rows.Next() {
		if err := rows.Scan(&tableName, &columnName); err != nil {
			return nil, err
		}
		cols, _ := tables[tableName]
		tables[tableName] = append(cols, columnName)
	}
	return tables, rows.Err()
}
