package splace

import (
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"
)

type queryOptions struct {
	table   string
	columns []string
	mode    Mode
	search  string
	limit   int

	update  bool
	replace string
}

type queryBuilder struct {
	b strings.Builder
}

func (b *queryBuilder) build(opt queryOptions) string {
	if opt.update {
		fmt.Fprintf(&b.b, "SELECT * FROM %s ", opt.table)
	} else {
		fmt.Fprintf(&b.b, "UPDATE %s ", opt.table)
		b.set(opt.columns, opt.search, opt.replace, opt.mode)
	}

	b.where(opt.columns, opt.search, opt.mode)

	if opt.limit > 0 {
		fmt.Fprintf(&b.b, "LIMIT %d", opt.limit)
	}

	return b.b.String()
}

func (b *queryBuilder) where(columns []string, search string, mode Mode) {
	b.b.WriteString("WHERE ")
	for i, col := range columns {
		fmt.Fprintf(&b.b, "`%s`", col)

		switch mode {
		case Equals:
			b.b.WriteString(" = ")
		case Like:
			b.b.WriteString(" LIKE ")
		case Regexp:
			b.b.WriteString(" REGEXP ")
		}

		b.b.WriteString(sqlf.Sprintf("'%s'", search).Query(sqlf.PostgresBindVar))

		if i < len(columns)-1 {
			b.b.WriteString(" AND ")
		}
	}
}

func (b *queryBuilder) set(columns []string, search, replace string, mode Mode) {
	b.b.WriteString("SET ")
	for i, col := range columns {
		fmt.Fprintf(&b.b, "`%s` = ", col)

		var q *sqlf.Query
		switch mode {
		case Equals:
			q = sqlf.Sprintf("'%s'", replace)
		case Like:
			q = sqlf.Sprintf("REPLACE(`%s`, '%s', '%s')", col, search, replace)
		case Regexp:
			q = sqlf.Sprintf("REGEXP_REPLACE(`%s`, '%s', '%s')", col, search, replace)
		}
		b.b.WriteString(q.Query(sqlf.PostgresBindVar))

		if i < len(columns)-1 {
			b.b.WriteString(" AND ")
		}
	}
}
