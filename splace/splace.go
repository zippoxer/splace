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

func (s *Splace) Search(ctx context.Context, opt SearchOptions) *Searcher {
	sr := newSearcher(ctx, s.db, opt)
	go sr.start()
	return sr
}

func (s *Splace) Replace(ctx context.Context, opt ReplaceOptions) *Replacer {
	r := newReplacer(ctx, s.db, opt)
	go r.start()
	return r
}

func (s *Splace) Tables(ctx context.Context) (map[string][]string, error) {
	rows, err := s.db.Query(ctx, `SELECT TABLE_NAME, COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA = ?`, s.db.Database())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := make(map[string][]string)
	for rows.Next() {
		row, err := rows.ScanStrings()
		if err != nil {
			return nil, err
		}
		cols, _ := tables[row[0]]
		tables[row[0]] = append(cols, row[1])
	}
	return tables, rows.Err()
}
