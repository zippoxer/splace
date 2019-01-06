package splace

import (
	"context"
	"strings"

	"github.com/zippoxer/splace/splace/querier"
)

type Mode int

const (
	Equals Mode = iota
	Contains
	Like
	Regexp
)

type TableMap map[string][]ColumnInfo

type ColumnInfo struct {
	Column string
	Type   string
}

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

func (s *Splace) Tables(ctx context.Context) (TableMap, error) {
	query := `SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE FROM ` +
		`INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA = ?`
	rows, err := s.db.Query(ctx,
		query,
		s.db.Config().Database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := TableMap{}
	for rows.Next() {
		row, err := rows.ScanStrings()
		if err != nil {
			return nil, err
		}
		cols, _ := tables[row[0]]
		tables[row[0]] = append(cols, ColumnInfo{
			Column: row[1],
			Type:   row[2],
		})
	}
	return tables, rows.Err()
}

func isColumnTypeReplacable(columnType string) bool {
	switch strings.ToLower(columnType) {
	// Date & time fields are validated and may error if the replacement isn't correct.
	case "datetime", "datetime2", "smalldatetime", "date", "time", "datetimeoffset", "timestamp":
		return false
	}
	return true
}
