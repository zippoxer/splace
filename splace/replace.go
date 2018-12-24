package splace

import "time"

type ReplaceOptions struct {
	Search  string
	Replace string
	Mode    Mode

	// Tables is a map of selected table names and column names.
	Tables map[string][]string

	// Limit sets the maximum amount of rows updated with each query.
	// A lower limit would provide progress updates more frequently, while a higher limit would perform better.
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
	End   time.Time
}

func (r ReplaceResult) Done() bool {
	return !r.End.IsZero()
}

func (s *Splace) Replace(opt ReplaceOptions) (<-chan ReplaceResult, error) {
	return nil, nil
}
