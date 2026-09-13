package generic_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/internal/test"
)

func TestGeneric_EncodeConditions(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeConditions, []test.EncoderTestCase[[]dialects.Condition]{
		{
			Name:             "empty",
			Builder:          []dialects.Condition{},
			ExpectedSQL:      "",
			ExpectedBindings: []any{},
		},
		{
			Name: "single simple",
			Builder: []dialects.Condition{
				{
					Column:   dialects.Column{Column: "foo"},
					Operator: "=",
					Value:    "bar",
				},
			},
			ExpectedSQL:      "`foo` = ?",
			ExpectedBindings: []any{"bar"},
		},
		{
			Name: "single nil",
			Builder: []dialects.Condition{
				{
					Column:   dialects.Column{Column: "foo"},
					Operator: "=",
					Value:    nil,
				},
			},
			ExpectedSQL:      "`foo` IS NULL",
			ExpectedBindings: []any{},
		},
		{
			Name: "single not nil",
			Builder: []dialects.Condition{
				{
					Column:   dialects.Column{Column: "foo"},
					Operator: "!=",
					Value:    nil,
				},
			},
			ExpectedSQL:      "`foo` IS NOT NULL",
			ExpectedBindings: []any{},
		},
		{
			Name: "and simple",
			Builder: []dialects.Condition{
				{
					Column:   dialects.Column{Column: "foo"},
					Operator: "=",
					Value:    "bar",
				},
				{
					Column:   dialects.Column{Column: "baz"},
					Operator: "=",
					Value:    "foo",
				},
			},
			ExpectedSQL:      "`foo` = ? AND `baz` = ?",
			ExpectedBindings: []any{"bar", "foo"},
		},
		{
			Name: "sub conditions",
			Builder: []dialects.Condition{
				{
					Value: []dialects.Condition{
						{
							Column:   dialects.Column{Column: "foo"},
							Operator: "=",
							Value:    "bar",
							Or:       true,
						},
						{
							Column:   dialects.Column{Column: "baz"},
							Operator: "=",
							Value:    "foo",
							Or:       true,
						}},
				},
				{
					Column:   dialects.Column{Column: "foo"},
					Operator: "=",
					Value:    "bar",
				},
			},
			ExpectedSQL:      "(`foo` = ? OR `baz` = ?) AND `foo` = ?",
			ExpectedBindings: []any{"bar", "foo", "bar"},
		},
	})
}
