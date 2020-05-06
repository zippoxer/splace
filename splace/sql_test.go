package splace

import (
	"strings"
	"testing"
)

var queryBuilderTests = []struct {
	in  queryOptions
	out string
}{
	{
		queryOptions{
			table:   "people",
			columns: []string{"name", "address"},
			mode:    Equals,
			search:  "Dvid",
			limit:   1000,
			update:  true,
			replace: "David",
		},
		"UPDATE `people` SET `name` = 'David' , `address` = 'David' WHERE `name` = 'Dvid' OR `address` = 'Dvid' LIMIT 1000",
	},
	{
		queryOptions{
			table:   "people",
			columns: []string{"name", "address"},
			mode:    Equals,
			search:  "Dvid",
			limit:   1000,
		},
		"SELECT * FROM `people` WHERE `name` = 'Dvid' OR `address` = 'Dvid' LIMIT 1000",
	},
	{
		queryOptions{
			table:   "people",
			columns: []string{"name", "address"},
			mode:    Like,
			search:  "%Dvid%",
		},
		"SELECT * FROM `people` WHERE `name` LIKE '%Dvid%' OR `address` LIKE '%Dvid%'",
	},
}

func TestQueryBuilder(t *testing.T) {
	sb := &queryBuilder{}
	for i, test := range queryBuilderTests {
		result := sb.build(test.in)
		if strings.ToLower(strings.TrimSpace(result)) != strings.ToLower(strings.TrimSpace(test.out)) {
			t.Errorf("failed test %d: expected %q, got %q", i, test.out, result)
		}
	}
}
