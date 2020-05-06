package splace

import (
	"context"
	"sync"
	"time"

	"github.com/zippoxer/splace/splace/querier"
)

type SearchOptions struct {
	Search string
	Mode   Mode

	// Tables is a map of selected table names and column names.
	Tables TableMap

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
	ctx       context.Context
	ctxCancel context.CancelFunc
	db        querier.Querier
	opt       SearchOptions

	results chan SearchResult
	done    chan error
}

type searchTask struct {
	table   string
	columns []string
}

func newSearcher(ctx context.Context, db querier.Querier, opt SearchOptions) *Searcher {
	c, cancel := context.WithCancel(ctx)
	return &Searcher{
		ctx:       c,
		ctxCancel: cancel,
		db:        db,
		opt:       opt,
		results:   make(chan SearchResult, 32),
		done:      make(chan error),
	}
}

func (s *Searcher) start() {
	defer close(s.results)
	defer close(s.done)

	var (
		tasks = make(chan searchTask)
		wg    sync.WaitGroup
		wgErr error
	)
	defer func() {
		s.done <- wgErr
	}()

	// Produce a search task for each table.
	go func() {
		defer close(tasks)
		for table, columns := range s.opt.Tables {
			var columnNames []string
			for _, col := range columns {
				columnNames = append(columnNames, col.Column)
			}
			if len(columnNames) == 0 {
				continue
			}
			tasks <- searchTask{
				table:   table,
				columns: columnNames,
			}
		}
	}()

	// Spawn a fixed amount of workers.
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				err := s.searchTable(task.table, task.columns)
				if err != nil && s.ctx.Err() != context.Canceled {
					// Cancel all tasks.
					s.ctxCancel()
					wgErr = err
				}
			}
		}()
	}

	wg.Wait()
}

func (s *Searcher) searchTable(table string, columns []string) error {
	qb := &queryBuilder{}
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
		if s.opt.Limit == 0 || n == 0 {
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
