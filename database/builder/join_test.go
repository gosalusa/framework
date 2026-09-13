package builder_test

import (
	"testing"

	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestJoin(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "join",
			Builder:          NewTestBuilder().Join("bars", "bars.foo_id", "=", "foos.id"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" JOIN \"bars\" ON \"bars\".\"foo_id\" = \"foos\".\"id\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "left join",
			Builder:          NewTestBuilder().LeftJoin("bars", "bars.foo_id", "=", "foos.id"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" LEFT JOIN \"bars\" ON \"bars\".\"foo_id\" = \"foos\".\"id\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "right join",
			Builder:          NewTestBuilder().RightJoin("bars", "bars.foo_id", "=", "foos.id"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" RIGHT JOIN \"bars\" ON \"bars\".\"foo_id\" = \"foos\".\"id\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "inner join",
			Builder:          NewTestBuilder().InnerJoin("bars", "bars.foo_id", "=", "foos.id"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" INNER JOIN \"bars\" ON \"bars\".\"foo_id\" = \"foos\".\"id\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "cross join",
			Builder:          NewTestBuilder().CrossJoin("bars", "bars.foo_id", "=", "foos.id"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" CROSS JOIN \"bars\" ON \"bars\".\"foo_id\" = \"foos\".\"id\"",
			ExpectedBindings: []any{},
		},
		{
			Name: "join on",
			Builder: NewTestBuilder().JoinOn("bars", func(q *builder.Conditions) {
				q.Where("a", ">", 4).WhereColumn("b", "=", "c")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" JOIN \"bars\" ON \"a\" > ? AND \"b\" = \"c\"",
			ExpectedBindings: []any{4},
		},
		{
			Name: "left join on",
			Builder: NewTestBuilder().LeftJoinOn("bars", func(q *builder.Conditions) {
				q.Where("a", ">", 4).WhereColumn("b", "=", "c")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" LEFT JOIN \"bars\" ON \"a\" > ? AND \"b\" = \"c\"",
			ExpectedBindings: []any{4},
		},
		{
			Name: "right join on",
			Builder: NewTestBuilder().RightJoinOn("bars", func(q *builder.Conditions) {
				q.Where("a", ">", 4).WhereColumn("b", "=", "c")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" RIGHT JOIN \"bars\" ON \"a\" > ? AND \"b\" = \"c\"",
			ExpectedBindings: []any{4},
		},
		{
			Name: "inner join on",
			Builder: NewTestBuilder().InnerJoinOn("bars", func(q *builder.Conditions) {
				q.Where("a", ">", 4).WhereColumn("b", "=", "c")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" INNER JOIN \"bars\" ON \"a\" > ? AND \"b\" = \"c\"",
			ExpectedBindings: []any{4},
		},
		{
			Name: "cross join on",
			Builder: NewTestBuilder().CrossJoinOn("bars", func(q *builder.Conditions) {
				q.Where("a", ">", 4).WhereColumn("b", "=", "c")
			}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" CROSS JOIN \"bars\" ON \"a\" > ? AND \"b\" = \"c\"",
			ExpectedBindings: []any{4},
		},
		{
			Name:             "multiple joins",
			Builder:          NewTestBuilder().Join("a", "b", "=", "c").Join("d", "e", "=", "f"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" JOIN \"a\" ON \"b\" = \"c\" JOIN \"d\" ON \"e\" = \"f\"",
			ExpectedBindings: []any{},
		},
	})
}
