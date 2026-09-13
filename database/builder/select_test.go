package builder_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestSelect(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "one select",
			Builder:          NewTestBuilder().Select("a"),
			ExpectedSQLite:   "SELECT \"a\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "two select",
			Builder:          NewTestBuilder().Select("a", "b"),
			ExpectedSQLite:   "SELECT \"a\", \"b\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "different table",
			Builder:          NewTestBuilder().Select("a.b"),
			ExpectedSQLite:   "SELECT \"a\".\"b\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "distinct",
			Builder:          NewTestBuilder().Select("a").Distinct(),
			ExpectedSQLite:   "SELECT DISTINCT \"a\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "subquery",
			Builder:          NewTestBuilder().SelectSubquery(NewTestBuilder().Select("a"), "test"),
			ExpectedSQLite:   "SELECT (SELECT \"a\" FROM \"foos\") AS \"test\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "all columns",
			Builder:          NewTestBuilder().Select("*"),
			ExpectedSQLite:   "SELECT * FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "all columns from table",
			Builder:          NewTestBuilder().Select("foos.*"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "as",
			Builder:          NewTestBuilder().Select("foo.bar as baz"),
			ExpectedSQLite:   `SELECT "foo"."bar" AS "baz" FROM "foos"`,
			ExpectedBindings: []any{},
		},
		{
			Name:             "raw",
			Builder:          NewTestBuilder().SelectRaw("foo.bar as baz"),
			ExpectedSQLite:   `SELECT foo.bar as baz FROM "foos"`,
			ExpectedBindings: []any{},
		},
	})
}
