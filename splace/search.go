package splace

import (
	"context"
	"time"

	"github.com/zippoxer/splace/splace/querier"
)

type SearchOptions struct {
	Search string
	Mode   Mode

	// Tables is a map of selected table names and column names.
	Tables map[string][]string

	// Limit sets the maximum amount of rows returned by each query.
	// A lower limit would provide frequent progress updates,
	// while with a higher limit the operation would complete faster.
	// Set to 0 for no limit.
	Limit int
}

type SearchResult struct {
	Table   string
	Columns []string
	SQL     string

	// Rows transmits the number of updated rows as soon as
	// each query completes. Expect only one transmission if SearchOptions.Limit is set to zero.
	// Rows is closed when we're done searching this table.
	Rows <-chan []string

	Start time.Time
}

type Searcher struct {
	ctx context.Context
	db  querier.Querier
	opt SearchOptions

	results chan SearchResult
	done    chan error
}

func newSearcher(ctx context.Context, db querier.Querier, opt SearchOptions) *Searcher {
	return &Searcher{
		ctx:     ctx,
		db:      db,
		opt:     opt,
		results: make(chan SearchResult, 32),
		done:    make(chan error),
	}
}

func (s *Searcher) start() {
	defer close(s.results)
	defer close(s.done)
	s.done <- s.search()
}

func (s *Searcher) search() error {
	qb := &queryBuilder{}
	for table, columns := range s.opt.Tables {
		if err := s.searchTable(qb, table, columns); err != nil {
			return err
		}
	}
	return nil
}

func (s *Searcher) searchTable(qb *queryBuilder, table string, columns []string) error {
	iterations := make(chan []string, 128)
	defer close(iterations)

	offset := 0
	for {
		query := qb.build(queryOptions{
			table:   table,
			columns: columns,
			mode:    s.opt.Mode,
			search:  s.opt.Search,
			offset:  offset,
			limit:   s.opt.Limit,
		})

		rows, err := s.db.Query(s.ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		if offset == 0 {
			resultColumns, err := rows.Columns()
			if err != nil {
				return err
			}

			s.results <- SearchResult{
				Table:   table,
				Columns: resultColumns,
				SQL:     query,
				Rows:    iterations,
				Start:   time.Now(),
			}
		}

		n := 0
		for rows.Next() {
			row, err := rows.ScanStrings()
			if err != nil {
				return err
			}
			cpy := make([]string, len(row))
			copy(cpy, row)
			iterations <- cpy
			n++
		}
		if rows.Err() != nil {
			return err
		}
		if n == 0 {
			return nil
		}

		offset += s.opt.Limit
	}
}

func (s *Searcher) Results() <-chan SearchResult {
	return s.results
}

func (s *Searcher) Done() <-chan error {
	return s.done
}
