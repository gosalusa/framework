package builder

import "gosalusa.com/database/dialects"

// Where adds a basic where clause to the query.
func (b *Builder) Where(column, operator string, value any) *Builder {
	b.wheres = b.wheres.Where(column, operator, value)
	return b
}

// Having adds a basic having clause to the query.
func (b *Builder) Having(column, operator string, value any) *Builder {
	b.havings = b.havings.Where(column, operator, value)
	return b
}

// OrWhere adds an or where clause to the query
func (b *Builder) OrWhere(column, operator string, value any) *Builder {
	b.wheres = b.wheres.OrWhere(column, operator, value)
	return b
}

// OrHaving adds an or having clause to the query
func (b *Builder) OrHaving(column, operator string, value any) *Builder {
	b.havings = b.havings.OrWhere(column, operator, value)
	return b
}

// WhereColumn adds a where clause to the query comparing two columns.
func (b *Builder) WhereColumn(column, operator string, valueColumn string) *Builder {
	b.wheres = b.wheres.WhereColumn(column, operator, valueColumn)
	return b
}

// HavingColumn adds a having clause to the query comparing two columns.
func (b *Builder) HavingColumn(column, operator string, valueColumn string) *Builder {
	b.havings = b.havings.WhereColumn(column, operator, valueColumn)
	return b
}

// OrWhereColumn adds an or where clause to the query comparing two columns.
func (b *Builder) OrWhereColumn(column, operator string, valueColumn string) *Builder {
	b.wheres = b.wheres.OrWhereColumn(column, operator, valueColumn)
	return b
}

// OrHavingColumn adds an or having clause to the query comparing two columns.
func (b *Builder) OrHavingColumn(column, operator string, valueColumn string) *Builder {
	b.havings = b.havings.OrWhereColumn(column, operator, valueColumn)
	return b
}

// WhereIn adds a where in clause to the query.
func (b *Builder) WhereIn(column string, values []any) *Builder {
	b.wheres = b.wheres.WhereIn(column, values)
	return b
}

// HavingIn adds a having in clause to the query.
func (b *Builder) HavingIn(column string, values []any) *Builder {
	b.havings = b.havings.WhereIn(column, values)
	return b
}

// OrWhereIn adds an or where in clause to the query.
func (b *Builder) OrWhereIn(column string, values []any) *Builder {
	b.wheres = b.wheres.OrWhereIn(column, values)
	return b
}

// OrHavingIn adds an or having in clause to the query.
func (b *Builder) OrHavingIn(column string, values []any) *Builder {
	b.havings = b.havings.OrWhereIn(column, values)
	return b
}

// WhereExists add an exists clause to the query.
func (b *Builder) WhereExists(query dialects.QueryBuilder) *Builder {
	b.wheres = b.wheres.WhereExists(query)
	return b
}

// HavingExists add an exists clause to the query.
func (b *Builder) HavingExists(query dialects.QueryBuilder) *Builder {
	b.havings = b.havings.WhereExists(query)
	return b
}

// OrWhereExists add an or exists clause to the query.
func (b *Builder) OrWhereExists(query dialects.QueryBuilder) *Builder {
	b.wheres = b.wheres.OrWhereExists(query)
	return b
}

// OrHavingExists add an or exists clause to the query.
func (b *Builder) OrHavingExists(query dialects.QueryBuilder) *Builder {
	b.havings = b.havings.OrWhereExists(query)
	return b
}

// WhereNotExists add a not exists clause to the query.
func (b *Builder) WhereNotExists(query dialects.QueryBuilder) *Builder {
	b.wheres = b.wheres.WhereNotExists(query)
	return b
}

// HavingNotExists add a not exists clause to the query.
func (b *Builder) HavingNotExists(query dialects.QueryBuilder) *Builder {
	b.havings = b.havings.WhereNotExists(query)
	return b
}

// OrWhereNotExists add an or not exists clause to the query.
func (b *Builder) OrWhereNotExists(query dialects.QueryBuilder) *Builder {
	b.wheres = b.wheres.OrWhereNotExists(query)
	return b
}

// OrHavingNotExists add an or not exists clause to the query.
func (b *Builder) OrHavingNotExists(query dialects.QueryBuilder) *Builder {
	b.havings = b.havings.OrWhereNotExists(query)
	return b
}

// WhereSubquery adds a where clause to the query comparing a column and a subquery.
func (b *Builder) WhereSubquery(subquery dialects.QueryBuilder, operator string, value any) *Builder {
	b.wheres = b.wheres.WhereSubquery(subquery, operator, value)
	return b
}

// HavingSubquery adds a having clause to the query comparing a column and a subquery.
func (b *Builder) HavingSubquery(subquery dialects.QueryBuilder, operator string, value any) *Builder {
	b.havings = b.havings.WhereSubquery(subquery, operator, value)
	return b
}

// OrWhereSubquery adds an or where clause to the query comparing a column and a subquery.
func (b *Builder) OrWhereSubquery(subquery dialects.QueryBuilder, operator string, value any) *Builder {
	b.wheres = b.wheres.OrWhereSubquery(subquery, operator, value)
	return b
}

// OrHavingSubquery adds an or having clause to the query comparing a column and a subquery.
func (b *Builder) OrHavingSubquery(subquery dialects.QueryBuilder, operator string, value any) *Builder {
	b.havings = b.havings.OrWhereSubquery(subquery, operator, value)
	return b
}

// WhereHas adds a relationship exists condition to the query with where clauses.
func (b *Builder) WhereHas(relation string, cb func(q *Builder) *Builder) *Builder {
	b.wheres = b.wheres.WhereHas(relation, cb)
	return b
}

// HavingHas adds a relationship exists condition to the query with having clauses.
func (b *Builder) HavingHas(relation string, cb func(q *Builder) *Builder) *Builder {
	b.havings = b.havings.WhereHas(relation, cb)
	return b
}

// OrWhereHas adds a relationship exists condition to the query with where clauses and an or.
func (b *Builder) OrWhereHas(relation string, cb func(q *Builder) *Builder) *Builder {
	b.wheres = b.wheres.OrWhereHas(relation, cb)
	return b
}

// OrHavingHas adds a relationship exists condition to the query with having clauses and an or.
func (b *Builder) OrHavingHas(relation string, cb func(q *Builder) *Builder) *Builder {
	b.havings = b.havings.OrWhereHas(relation, cb)
	return b
}

// WhereRaw adds a raw where clause to the query.
func (b *Builder) WhereRaw(rawSql string, bindings ...any) *Builder {
	b.wheres = b.wheres.WhereRaw(rawSql, bindings...)
	return b
}

// HavingRaw adds a raw having clause to the query.
func (b *Builder) HavingRaw(rawSql string, bindings ...any) *Builder {
	b.havings = b.havings.WhereRaw(rawSql, bindings...)
	return b
}

// OrWhereRaw adds a raw or where clause to the query.
func (b *Builder) OrWhereRaw(rawSql string, bindings ...any) *Builder {
	b.wheres = b.wheres.OrWhereRaw(rawSql, bindings...)
	return b
}

// OrHavingRaw adds a raw or having clause to the query.
func (b *Builder) OrHavingRaw(rawSql string, bindings ...any) *Builder {
	b.havings = b.havings.OrWhereRaw(rawSql, bindings...)
	return b
}

// And adds a group of conditions to the query
func (b *Builder) And(cb func(q *Conditions)) *Builder {
	b.wheres = b.wheres.And(cb)
	return b
}

// HavingAnd adds a group of conditions to the query
func (b *Builder) HavingAnd(cb func(q *Conditions)) *Builder {
	b.havings = b.havings.And(cb)
	return b
}

// Or adds a group of conditions to the query with an or
func (b *Builder) Or(cb func(q *Conditions)) *Builder {
	b.wheres = b.wheres.Or(cb)
	return b
}

// HavingOr adds a group of conditions to the query with an or
func (b *Builder) HavingOr(cb func(q *Conditions)) *Builder {
	b.havings = b.havings.Or(cb)
	return b
}

// WithScope adds a local scope to a query.
func (b *Builder) WithScope(scope *Scope) *Builder {
	b.scopes = b.scopes.WithScope(scope)
	return b
}

// WithoutScope removes the given scope from the local scopes.
func (b *Builder) WithoutScope(scope *Scope) *Builder {
	b.scopes = b.scopes.WithoutScope(scope)
	return b
}

// WithoutGlobalScope removes a global scope from the query.
func (b *Builder) WithoutGlobalScope(scope *Scope) *Builder {
	b.scopes = b.scopes.WithoutGlobalScope(scope)
	return b
}
