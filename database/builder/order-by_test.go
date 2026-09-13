package builder_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestOrderBy(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "one group",
			Builder:          NewTestBuilder().OrderBy("a"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" ORDER BY \"a\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "two groups",
			Builder:          NewTestBuilder().OrderBy("a").OrderBy("b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" ORDER BY \"a\", \"b\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "different table",
			Builder:          NewTestBuilder().OrderBy("a.b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" ORDER BY \"a\".\"b\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "descending",
			Builder:          NewTestBuilder().OrderByDesc("a"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" ORDER BY \"a\" DESC",
			ExpectedBindings: []any{},
		},
	})
}
