package generic_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/internal/test"
)

func TestGeneric_EncodeSelectQuery(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeSelectQuery, []test.EncoderTestCase[*dialects.SelectQuery]{
		// {
		// 	Name:             "empty",
		// 	Builder:          &dialects.SelectQuery{},
		// 	ExpectedSQL:      "",
		// 	ExpectedBindings: []any{},
		// },
		{
			Name: "simple",
			Builder: &dialects.SelectQuery{
				Select: dialects.Select{
					Columns: []dialects.Column{{Column: "foo"}},
				},
				From: "bar",
				Joins: []dialects.Join{
					{
						Direction: "LEFT",
						Table:     "joined_table",
						Conditions: []dialects.Condition{
							{
								Column:   dialects.Column{Column: "joined_table_id"},
								Operator: "=",
								Value:    dialects.Column{Column: "id"},
							},
						},
					},
				},
				Wheres: []dialects.Condition{
					{
						Column:   dialects.Column{Column: "foo"},
						Operator: "=",
						Value:    "baz",
					},
				},
				GroupBys: []string{"foo"},
				Havings: []dialects.Condition{
					{
						Column:   dialects.Column{Column: "foo"},
						Operator: "=",
						Value:    "baz",
					},
				},
				OrderBys: []dialects.OrderColumn{{Column: "foo"}},
				Limit: dialects.Limit{
					Limit:  5,
					Offset: 10,
				},
			},
			ExpectedSQL:      "SELECT `foo` FROM `bar` LEFT JOIN `joined_table` ON `joined_table_id` = `id` WHERE `foo` = ? GROUP BY `foo` HAVING `foo` = ? ORDER BY `foo` LIMIT ? OFFSET ?",
			ExpectedBindings: []any{"baz", "baz", 5, 10},
		},
	})
}
