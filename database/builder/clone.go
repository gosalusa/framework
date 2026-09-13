package builder

import "gosalusa.com/database/dialects"

// Clone returns an independent copy of the query. Mutating the returned
// builder, or the original, will not affect the other.
func (b *ModelBuilder[T]) Clone() *ModelBuilder[T] {
	return &ModelBuilder[T]{
		builder:       b.builder.Clone(),
		withs:         cloneSlice(b.withs),
		withoutScopes: b.withoutScopes.Clone(),
	}
}

// Clone returns an independent copy of the query. Mutating the returned
// builder, or the original, will not affect the other.
func (b *Builder) Clone() *Builder {
	return &Builder{
		query: dialects.SelectQuery{
			Select: dialects.Select{
				Distinct: b.query.Select.Distinct,
				Columns:  cloneSlice(b.query.Select.Columns),
			},
			From:     b.query.From,
			Joins:    cloneSlice(b.query.Joins),
			Wheres:   cloneSlice(b.query.Wheres),
			Havings:  cloneSlice(b.query.Havings),
			GroupBys: cloneSlice(b.query.GroupBys),
			OrderBys: cloneSlice(b.query.OrderBys),
			Limit:    b.query.Limit,
		},
		scopes:  b.scopes.Clone(),
		wheres:  b.wheres.Clone(),
		havings: b.havings.Clone(),
		ctx:     b.ctx,
	}
}

func cloneSlice[T any](arr []T) []T {
	l := make([]T, len(arr))
	copy(l, arr)
	return l
}
