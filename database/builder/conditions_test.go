package builder_test

import (
	"testing"

	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestWhere(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "one where",
			Builder:          NewTestBuilder().Where("a", "=", "b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = ?",
			ExpectedBindings: []any{"b"},
		},
		{
			Name:             "2 wheres",
			Builder:          NewTestBuilder().Where("a", "=", "b").Where("c", "=", "d"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = ? AND \"c\" = ?",
			ExpectedBindings: []any{"b", "d"},
		},
		{
			Name:             "null",
			Builder:          NewTestBuilder().Where("a", "=", nil),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" IS NULL",
			ExpectedBindings: []any{},
		},
		{
			Name:             "not null",
			Builder:          NewTestBuilder().Where("a", "!=", nil),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" IS NOT NULL",
			ExpectedBindings: []any{},
		},
		{
			Name:             "specified table",
			Builder:          NewTestBuilder().Where("foo.a", "=", "b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"foo\".\"a\" = ?",
			ExpectedBindings: []any{"b"},
		},
		{
			Name:             "or where",
			Builder:          NewTestBuilder().Where("a", "=", "b").OrWhere("c", "=", "d"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = ? OR \"c\" = ?",
			ExpectedBindings: []any{"b", "d"},
		},
		{
			Name: "and group",
			Builder: NewTestBuilder().And(func(wl *builder.Conditions) {
				wl.Where("a", "=", "a").OrWhere("b", "=", "b")
			}).And(func(wl *builder.Conditions) {
				wl.Where("c", "=", "c").OrWhere("d", "=", "d")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE (\"a\" = ? OR \"b\" = ?) AND (\"c\" = ? OR \"d\" = ?)",
			ExpectedBindings: []any{"a", "b", "c", "d"},
		},
		{
			Name: "or group",
			Builder: NewTestBuilder().Or(func(wl *builder.Conditions) {
				wl.Where("a", "=", "a").Where("b", "=", "b")
			}).Or(func(wl *builder.Conditions) {
				wl.Where("c", "=", "c").Where("d", "=", "d")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE (\"a\" = ? AND \"b\" = ?) OR (\"c\" = ? AND \"d\" = ?)",
			ExpectedBindings: []any{"a", "b", "c", "d"},
		},
		{
			Name:             "subquery",
			Builder:          NewTestBuilder().Where("a", "=", NewTestBuilder().Select("a").Where("id", "=", 1)),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = (SELECT \"a\" FROM \"foos\" WHERE \"id\" = ?)",
			ExpectedBindings: []any{1},
		},
		{
			Name:             "wherein",
			Builder:          NewTestBuilder().WhereIn("a", []any{1, 2, 3}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" in (?, ?, ?)",
			ExpectedBindings: []any{1, 2, 3},
		},
		{
			Name:             "where subquery",
			Builder:          NewTestBuilder().WhereSubquery(NewTestBuilder().Select("a").Where("id", "=", 1), "=", "a"),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" WHERE (SELECT "a" FROM "foos" WHERE "id" = ?) = ?`,
			ExpectedBindings: []any{1, "a"},
		},
		{
			Name:             "where exists",
			Builder:          NewTestBuilder().WhereExists(NewTestBuilder().Select("a").Where("id", "=", 1)),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" WHERE EXISTS (SELECT "a" FROM "foos" WHERE "id" = ?)`,
			ExpectedBindings: []any{1},
		},
		{
			Name: "whereHas HasOne",
			Builder: NewTestBuilder().WhereHas("Bar", func(q *builder.Builder) *builder.Builder {
				return q.Where("id", "=", "b")
			}),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" WHERE EXISTS (SELECT "bars".* FROM "bars" WHERE "foo_id" = "foos"."id" AND "id" = ?)`,
			ExpectedBindings: []any{"b"},
		},
		{
			Name: "whereHas HasMany",
			Builder: NewTestBuilder().WhereHas("Bars", func(q *builder.Builder) *builder.Builder {
				return q.Where("id", "=", "b")
			}),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" WHERE EXISTS (SELECT "bars".* FROM "bars" WHERE "foo_id" = "foos"."id" AND "id" = ?)`,
			ExpectedBindings: []any{"b"},
		},
		{
			Name: "whereHas BelongsTo",
			Builder: builder.From[*test.Bar]().WhereHas("Foo", func(q *builder.Builder) *builder.Builder {
				return q.Where("id", "=", "b")
			}),
			ExpectedSQLite:   `SELECT "bars".* FROM "bars" WHERE EXISTS (SELECT "foos".* FROM "foos" WHERE "id" = "bars"."foo_id" AND "id" = ?)`,
			ExpectedBindings: []any{"b"},
		},
		{
			Name:             "whereRaw",
			Builder:          NewTestBuilder().WhereRaw("function(a) = ?", "b"),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" WHERE function(a) = ?`,
			ExpectedBindings: []any{"b"},
		},
	})
}
