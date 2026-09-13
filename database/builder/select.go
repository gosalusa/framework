package builder

import (
	"strings"

	"gosalusa.com/database/dialects"
)

func parseColumn(c string) dialects.Column {
	parts := strings.SplitN(c, "as", 2)
	as := ""
	if len(parts) > 1 {
		as = strings.TrimSpace(parts[1])
	}
	return dialects.Column{
		Column: strings.TrimSpace(parts[0]),
		As:     as,
	}
}

// Select sets the columns to be selected.
func (b *Builder) Select(columns ...string) *Builder {
	identifiers := make([]dialects.Column, len(columns))
	for i, c := range columns {
		identifiers[i] = parseColumn(c)
	}
	b.query.Select.Columns = identifiers
	return b
}

// Select sets the columns to be selected.
func (b *Builder) SelectRaw(columns ...string) *Builder {
	identifiers := make([]dialects.Column, len(columns))
	for i, c := range columns {
		identifiers[i] = dialects.Column{Raw: c}
	}
	b.query.Select.Columns = identifiers
	return b
}

// AddSelect adds new columns to be selected.
func (b *Builder) AddSelect(columns ...string) *Builder {
	for _, c := range columns {
		b.query.Select.Columns = append(b.query.Select.Columns, parseColumn(c))
	}
	return b
}

// AddSelect adds new columns to be selected.
func (b *Builder) AddSelectRaw(columns ...string) *Builder {
	for _, c := range columns {
		b.query.Select.Columns = append(b.query.Select.Columns, dialects.Column{Raw: c})
	}
	return b
}

// SelectSubquery sets a subquery to be selected.
func (b *Builder) SelectSubquery(sb dialects.QueryBuilder, as string) *Builder {
	return b.Select().AddSelectSubquery(sb, as)
}

// AddSelectSubquery adds a subquery to be selected.
func (b *Builder) AddSelectSubquery(sb dialects.QueryBuilder, as string) *Builder {
	b.query.Select.Columns = append(b.query.Select.Columns, dialects.Column{
		SubQuery: sb,
		As:       as,
	})
	return b
}

// SelectFunction sets a column to be selected with a function applied.
func (b *Builder) SelectFunction(function, column string) *Builder {
	return b.Select().AddSelectFunction(function, column)
}

// SelectFunction adds a column to be selected with a function applied.
func (b *Builder) AddSelectFunction(function, column string) *Builder {
	b.query.Select.Columns = append(b.query.Select.Columns, dialects.Column{
		Function: &dialects.FunctionCall{
			Name:      function,
			Arguments: column,
		},
	})

	return b
}

// Distinct forces the query to only return distinct results.
func (b *Builder) Distinct() *Builder {
	b.query.Select.Distinct = true
	return b
}
