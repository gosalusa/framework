package generic

import (
	"gosalusa.com/database/dialects"
)

// EncodeSelects renders the columns of a SELECT clause, prefixing DISTINCT
// when set.
func (g *Generic) EncodeSelects(s *dialects.Select) (dialects.RawQuery, error) {
	if len(s.Columns) == 0 {
		return dialects.RawQuery{
			Bindings: []any{},
		}, nil
	}

	b := newRawQueryBuilder()
	if s.Distinct {
		b.AddString("DISTINCT")
	}

	columns := make([]dialects.RawQuery, len(s.Columns))
	var err error
	for i, c := range s.Columns {
		columns[i], err = g.EncodeColumn(&c)
		if err != nil {
			return dialects.RawQuery{}, err
		}
	}
	b.Add(joinRawQueries(columns, ", "), nil)

	return b.Build()
}

// EncodeFunctionCall renders a function applied to its arguments, such as
// count(*).
func (g *Generic) EncodeFunctionCall(fc *dialects.FunctionCall) (dialects.RawQuery, error) {
	return newRawQueryBuilder().
		AddString(fc.Name + "(" + g.core.Identifier(fc.Arguments) + ")").
		Build()
}

// EncodeColumn renders a single selectable expression: a column name, a
// function call, a subquery, or raw SQL, with an optional AS alias.
func (g *Generic) EncodeColumn(c *dialects.Column) (dialects.RawQuery, error) {
	if c.Raw != "" {
		return dialects.Raw(c.Raw), nil
	}

	b := newRawQueryBuilder()
	if c.Column != "" {
		b.AddString(g.core.Identifier(c.Column))
	} else if c.Function != nil {
		b.Add(g.EncodeFunctionCall(c.Function))
	} else if c.SubQuery != nil {
		b.Add(group(g.EncodeSelectQuery(c.SubQuery.Query())))
	}
	if c.As != "" {
		b.AddString("AS").AddString(g.core.Identifier(c.As))
	}
	return b.Build()
}
