package splace

import (
	"fmt"
	"strings"
)

type queryOptions struct {
	table   string
	columns []string
	mode    Mode
	search  string
	offset  int
	limit   int

	update  bool
	replace string
}

type queryBuilder struct {
	b strings.Builder
}

func (b *queryBuilder) build(opt queryOptions) string {
	if opt.update {
		fmt.Fprintf(&b.b, "UPDATE `%s` ", opt.table)
		b.set(opt.columns, opt.search, opt.replace, opt.mode)
	} else {
		fmt.Fprintf(&b.b, "SELECT * FROM `%s` ", opt.table)
	}

	b.where(opt.columns, opt.search, opt.mode)

	if opt.limit > 0 {
		if opt.offset > 0 {
			fmt.Fprintf(&b.b, "LIMIT %d, %d", opt.offset, opt.limit)
		} else {
			fmt.Fprintf(&b.b, "LIMIT %d", opt.limit)
		}
	}

	s := b.b.String()
	b.b.Reset()
	return s
}

func (b *queryBuilder) where(columns []string, search string, mode Mode) {
	b.b.WriteString("WHERE ")
	for i, col := range columns {
		fmt.Fprintf(&b.b, "`%s` ", col)

		switch mode {
		case Equals:
			b.b.WriteString(querySprintf("= '%s' ", search))
		case Contains:
			escapedSearch := strings.Replace(search, `%`, `\%`, -1)
			b.b.WriteString("LIKE BINARY '%" + querySprintf("%s", escapedSearch) + "%' ")
		case Like:
			b.b.WriteString(querySprintf("LIKE BINARY '%s' ", search))
		case Regexp:
			b.b.WriteString(querySprintf("REGEXP '%s' ", search))
		}

		if i < len(columns)-1 {
			b.b.WriteString("OR ")
		}
	}
}

func (b *queryBuilder) set(columns []string, search, replace string, mode Mode) {
	b.b.WriteString("SET ")
	for i, col := range columns {
		fmt.Fprintf(&b.b, "`%s` = ", col)

		switch mode {
		case Equals:
			b.b.WriteString(querySprintf("'%s' ", replace))
		case Contains:
			b.b.WriteString(querySprintf("REPLACE(`%s`, '%s', '%s') ", col, search, replace))
		case Like:
			panic("queryBuilder.set: update queries don't support Like")
		case Regexp:
			b.b.WriteString(querySprintf("REGEXP_REPLACE(`%s`, '%s', '%s') ", col, search, replace))
		}

		if i < len(columns)-1 {
			b.b.WriteString(", ")
		}
	}
}

func querySprintf(format string, args ...interface{}) string {
	for i := range args {
		args[i] = queryEscape(fmt.Sprint(args[i]))
	}
	return fmt.Sprintf(format, args...)
}

func queryEscape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]

		escape = 0

		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"': /* Better safe than sorry */
			escape = '"'
			break
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}

	return string(dest)
}
