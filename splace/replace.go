package splace

import (
	"context"
	"time"

	"github.com/zippoxer/splace/splace/querier"
)

type ReplaceOptions struct {
	Search  string
	Replace string
	Mode    Mode

	// Tables is a map of selected table names and column names.
	Tables TableMap

	// Limit sets the maximum amount of rows updated with each query.
	// A lower limit would provide frequent progress updates,
	// while with a higher limit the operation would complete faster.
	// Set to 0 for no limit.
	Limit int
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
	ctx context.Context
	db  querier.Querier
	opt ReplaceOptions

	results chan ReplaceResult
	done    chan error
}

func newReplacer(ctx context.Context, db querier.Querier, opt ReplaceOptions) *Replacer {
	return &Replacer{
		ctx:     ctx,
		db:      db,
		opt:     opt,
		results: make(chan ReplaceResult, 128),
		done:    make(chan error),
	}
}

func (r *Replacer) start() {
	defer close(r.results)
	defer close(r.done)
	r.done <- r.replace()
}

func (r *Replacer) replace() error {
	qb := &queryBuilder{}
	for table, columns := range r.opt.Tables {
		var cols []string
		for _, col := range columns {
			if isColumnTypeReplacable(col.Type) {
				cols = append(cols, col.Column)
			}
		}
		if err := r.replaceTable(qb, table, cols); err != nil {
			return err
		}
	}
	return nil
}

func (r *Replacer) replaceTable(qb *queryBuilder, table string, columns []string) error {
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
	defer close(iterations)

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
			return nil
		}
		iterations <- int(rowsAffected)
		if r.opt.Limit == 0 {
			return nil
		}
	}
}

func (r *Replacer) Results() <-chan ReplaceResult {
	return r.results
}

func (r *Replacer) Done() <-chan error {
	return r.done
}
