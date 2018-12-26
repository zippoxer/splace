package splace

import (
	"context"
	"splace/splace/querier"
	"time"
)

type ReplaceOptions struct {
	Search  string
	Replace string
	Mode    Mode

	// Tables is a map of selected table names and column names.
	Tables map[string][]string

	// Limit sets the maximum amount of rows updated with each query.
	// A lower limit would provide frequent progress updates,
	// while with a higher limit the operation would complete faster.
	// Set to 0 for no limit.
	Limit int

	Dry bool
}

type ReplaceResult struct {
	Table string
	SQL   string

	// AffectedRows transmits the number of updated rows as soon as
	// each query completes. Expect only one transmission if ReplaceOptions.Limit is set to zero.
	// AffectedRows is closed when we're done replacing in this table.
	AffectedRows <-chan int

	Start time.Time
}

type Replacer struct {
	db  querier.Querier
	opt ReplaceOptions
	ctx context.Context

	results chan ReplaceResult
	err     chan error
}

func newReplacer(db querier.Querier, opt ReplaceOptions) *Replacer {
	return &Replacer{
		db:  db,
		opt: opt,
	}
}

func (r *Replacer) start() {
	r.results = make(chan ReplaceResult)
	r.err = make(chan error)

	err := r.replace()
	if err != nil {
		r.err <- err
	}

	close(r.results)
	close(r.err)
}

func (r *Replacer) replace() error {
	qb := &queryBuilder{}
	for table, columns := range r.opt.Tables {
		query := qb.build(queryOptions{
			table:   table,
			columns: columns,
			mode:    r.opt.Mode,
			search:  r.opt.Search,
			limit:   r.opt.Limit,
			update:  true,
			replace: r.opt.Replace,
		})

		iterations := make(chan int)

		r.results <- ReplaceResult{
			Table:        table,
			SQL:          query,
			AffectedRows: iterations,
			Start:        time.Now(),
		}

		for {
			result, err := r.db.Exec(r.ctx, query)
			if err != nil {
				return err
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				break
			}
			iterations <- int(rowsAffected)
		}
	}
	return nil
}

func (r *Replacer) Results() <-chan ReplaceResult {
	return r.results
}

func (r *Replacer) Err() <-chan error {
	return r.err
}
