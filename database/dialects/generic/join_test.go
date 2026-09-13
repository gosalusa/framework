package generic_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/internal/test"
)

func TestGeneric_EncodeJoin(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeJoin, []test.EncoderTestCase[*dialects.Join]{
		// {
		// 	Name:             "empty",
		// 	Builder:          &dialects.Join{},
		// 	ExpectedSQL:      "",
		// 	ExpectedBindings: []any{},
		// },
		{
			Name: "simple",
			Builder: &dialects.Join{
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
			ExpectedSQL:      "LEFT JOIN `joined_table` ON `joined_table_id` = `id`",
			ExpectedBindings: []any{},
		},
		{
			Name: "no conditions",
			Builder: &dialects.Join{
				Table: "joined_table",
			},
			ExpectedSQL:      "JOIN `joined_table`",
			ExpectedBindings: []any{},
		},
	})
}
