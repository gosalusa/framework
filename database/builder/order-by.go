package builder

import "gosalusa.com/database/dialects"

// OrderBy adds an order by clause to the query.
func (b *Builder) OrderBy(column string) *Builder {
	b.query.OrderBys = append(b.query.OrderBys, dialects.OrderColumn{Column: column})
	return b
}

// OrderByDesc adds a descending order by clause to the query.
func (b *Builder) OrderByDesc(column string) *Builder {
	b.query.OrderBys = append(b.query.OrderBys, dialects.OrderColumn{Column: column, Descending: true})
	return b
}

// Unordered removes all order by clauses from the query.
func (b *Builder) Unordered() *Builder {
	b.query.OrderBys = []dialects.OrderColumn{}
	return b
}
