package generic_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/internal/test"
)

func TestGeneric_EncodeUpdateQuery(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeUpdateQuery, []test.EncoderTestCase[*dialects.UpdateQuery]{
		{
			Name: "single",
			Builder: &dialects.UpdateQuery{
				Table:  "foo",
				Values: map[string]any{"a": "b"},
			},
			ExpectedSQL:      "UPDATE `foo` SET `a` = ?",
			ExpectedBindings: []any{"b"},
		},
		{
			Name: "where",
			Builder: &dialects.UpdateQuery{
				Table:  "foo",
				Values: map[string]any{"a": "b"},
				Wheres: []dialects.Condition{
					{
						Column:   dialects.Column{Column: "b"},
						Operator: "=",
						Value:    2,
					},
				},
			},
			ExpectedSQL:      "UPDATE `foo` SET `a` = ? WHERE `b` = ?",
			ExpectedBindings: []any{"b", 2},
		},
	})
}

func TestGeneric_EncodeUpdateSet(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeUpdateSet, []test.EncoderTestCase[map[string]any]{
		{
			Name:             "single",
			Builder:          map[string]any{"a": "b"},
			ExpectedSQL:      "SET `a` = ?",
			ExpectedBindings: []any{"b"},
		},
		{
			Name:             "multiple",
			Builder:          map[string]any{"a": "b", "c": 2},
			ExpectedSQL:      "SET `a` = ?, `c` = ?",
			ExpectedBindings: []any{"b", 2},
		},
		{
			Name:             "column",
			Builder:          map[string]any{"a": dialects.Column{Column: "b"}},
			ExpectedSQL:      "SET `a` = `b`",
			ExpectedBindings: []any{},
		},
	})
}
