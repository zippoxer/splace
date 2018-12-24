package splace

import "time"

type SearchOptions struct {
	Search string
	Mode   Mode

	// Tables is a map of selected table names and column names.
	Tables map[string][]string

	// Limit sets the maximum amount of rows updated with each query.
	// A lower limit would provide frequent progress updates, while a higher limit would perform better.
	// Set to 0 for no limit.
	Limit int
}

type SearchResult struct {
	Table   string
	SQL     string
	Columns []string
	Rows    <-chan []string
	Start   time.Time
	End     time.Time
}

func (s *Splace) Search(opt SearchOptions) (<-chan SearchResult, error) {
	return nil, nil
}
