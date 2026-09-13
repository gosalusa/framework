package generic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/internal/test"
)

func TestGeneric_EncodeSelects(t *testing.T) {
	g := generic.New(&testCore{})
	test.EncoderTest(t, g.EncodeSelects, []test.EncoderTestCase[*dialects.Select]{
		{
			Name:             "empty",
			Builder:          &dialects.Select{},
			ExpectedSQL:      "",
			ExpectedBindings: []any{},
		},
		{
			Name: "single",
			Builder: &dialects.Select{
				Columns: []dialects.Column{
					{Column: "table.column"},
				},
			},
			ExpectedSQL:      "`table`.`column`",
			ExpectedBindings: []any{},
		},
		{
			Name: "distinct",
			Builder: &dialects.Select{
				Distinct: true,
				Columns: []dialects.Column{
					{Column: "table.column"},
				},
			},
			ExpectedSQL:      "DISTINCT `table`.`column`",
			ExpectedBindings: []any{},
		},
		{
			Name: "multi",
			Builder: &dialects.Select{
				Columns: []dialects.Column{
					{Column: "table1.column1"},
					{Column: "table2.column2"},
				},
			},
			ExpectedSQL:      "`table1`.`column1`, `table2`.`column2`",
			ExpectedBindings: []any{},
		},
		{
			Name: "as",
			Builder: &dialects.Select{
				Columns: []dialects.Column{
					{Column: "table.column", As: "name"},
				},
			},
			ExpectedSQL:      "`table`.`column` AS `name`",
			ExpectedBindings: []any{},
		},
		{
			Name: "function",
			Builder: &dialects.Select{
				Columns: []dialects.Column{
					{Function: &dialects.FunctionCall{
						Name:      "count",
						Arguments: "*",
					}},
				},
			},
			ExpectedSQL:      "count(*)",
			ExpectedBindings: []any{},
		},
		{
			Name: "disinct",
			Builder: &dialects.Select{
				Columns: []dialects.Column{
					{Column: "foo"},
				},
				Distinct: true,
			},
			ExpectedSQL:      "DISTINCT `foo`",
			ExpectedBindings: []any{},
		},
		{
			Name: "as",
			Builder: &dialects.Select{
				Columns: []dialects.Column{
					{Column: "foo", As: "bar"},
				},
				Distinct: true,
			},
			ExpectedSQL:      "DISTINCT `foo` AS `bar`",
			ExpectedBindings: []any{},
		},
	})
}

func TestGeneric_EncodeColumn(t *testing.T) {
	g := generic.New(&testCore{})
	t.Run("function", func(t *testing.T) {
		r, err := g.EncodeColumn(&dialects.Column{
			Function: &dialects.FunctionCall{Name: "count", Arguments: "*"},
		})
		require.NoError(t, err)
		assert.Equal(t, "count(*)", r.SQL)
	})
	t.Run("subquery", func(t *testing.T) {
		r, err := g.EncodeColumn(&dialects.Column{
			SubQuery: &selectQuery{&dialects.SelectQuery{
				Select: dialects.Select{Columns: []dialects.Column{{Column: "id"}}},
				From:   "foos",
			}},
			As: "x",
		})
		require.NoError(t, err)
		assert.Equal(t, "(SELECT `id` FROM `foos`) AS `x`", r.SQL)
	})
	t.Run("as", func(t *testing.T) {
		r, err := g.EncodeColumn(&dialects.Column{Column: "a", As: "b"})
		require.NoError(t, err)
		assert.Equal(t, "`a` AS `b`", r.SQL)
	})
	t.Run("empty", func(t *testing.T) {
		r, err := g.EncodeColumn(&dialects.Column{})
		require.NoError(t, err)
		assert.Equal(t, "", r.SQL)
	})
}
