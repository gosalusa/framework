package builder

import (
	"context"

	"gosalusa.com/database/dialects"
)

// WithContext adds a context to the query that will be used when fetching results.
func (b *ModelBuilder[T]) WithContext(ctx context.Context) *ModelBuilder[T] {
	b.builder = b.builder.WithContext(ctx)
	return b
}

// ForUpdate adds a FOR UPDATE clause to the query, locking the selected rows
// until the transaction is committed.
func (b *ModelBuilder[T]) ForUpdate() *ModelBuilder[T] {
	b.builder = b.builder.ForUpdate()
	return b
}

// ForUpdateSkipLocked adds a FOR UPDATE SKIP LOCKED clause to the query,
// locking the selected rows while skipping any rows that are already locked.
func (b *ModelBuilder[T]) ForUpdateSkipLocked() *ModelBuilder[T] {
	b.builder = b.builder.ForUpdateSkipLocked()
	return b
}

// From sets the table which the query is targeting.
func (b *ModelBuilder[T]) From(table string) *ModelBuilder[T] {
	b.builder = b.builder.From(table)
	return b
}

// Where adds a basic where clause to the query.
func (b *ModelBuilder[T]) Where(column, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.Where(column, operator, value)
	return b
}

// Having adds a basic having clause to the query.
func (b *ModelBuilder[T]) Having(column, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.Having(column, operator, value)
	return b
}

// OrWhere adds an or where clause to the query
func (b *ModelBuilder[T]) OrWhere(column, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.OrWhere(column, operator, value)
	return b
}

// OrHaving adds an or having clause to the query
func (b *ModelBuilder[T]) OrHaving(column, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.OrHaving(column, operator, value)
	return b
}

// WhereColumn adds a where clause to the query comparing two columns.
func (b *ModelBuilder[T]) WhereColumn(column, operator string, valueColumn string) *ModelBuilder[T] {
	b.builder = b.builder.WhereColumn(column, operator, valueColumn)
	return b
}

// HavingColumn adds a having clause to the query comparing two columns.
func (b *ModelBuilder[T]) HavingColumn(column, operator string, valueColumn string) *ModelBuilder[T] {
	b.builder = b.builder.HavingColumn(column, operator, valueColumn)
	return b
}

// OrWhereColumn adds an or where clause to the query comparing two columns.
func (b *ModelBuilder[T]) OrWhereColumn(column, operator string, valueColumn string) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereColumn(column, operator, valueColumn)
	return b
}

// OrHavingColumn adds an or having clause to the query comparing two columns.
func (b *ModelBuilder[T]) OrHavingColumn(column, operator string, valueColumn string) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingColumn(column, operator, valueColumn)
	return b
}

// WhereIn adds a where in clause to the query.
func (b *ModelBuilder[T]) WhereIn(column string, values []any) *ModelBuilder[T] {
	b.builder = b.builder.WhereIn(column, values)
	return b
}

// HavingIn adds a having in clause to the query.
func (b *ModelBuilder[T]) HavingIn(column string, values []any) *ModelBuilder[T] {
	b.builder = b.builder.HavingIn(column, values)
	return b
}

// OrWhereIn adds an or where in clause to the query.
func (b *ModelBuilder[T]) OrWhereIn(column string, values []any) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereIn(column, values)
	return b
}

// OrHavingIn adds an or having in clause to the query.
func (b *ModelBuilder[T]) OrHavingIn(column string, values []any) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingIn(column, values)
	return b
}

// WhereExists add an exists clause to the query.
func (b *ModelBuilder[T]) WhereExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.WhereExists(query)
	return b
}

// HavingExists add an exists clause to the query.
func (b *ModelBuilder[T]) HavingExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.HavingExists(query)
	return b
}

// OrWhereExists add an or exists clause to the query.
func (b *ModelBuilder[T]) OrWhereExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereExists(query)
	return b
}

// OrHavingExists add an or exists clause to the query.
func (b *ModelBuilder[T]) OrHavingExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingExists(query)
	return b
}

// WhereNotExists add a not exists clause to the query.
func (b *ModelBuilder[T]) WhereNotExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.WhereNotExists(query)
	return b
}

// HavingNotExists add a not exists clause to the query.
func (b *ModelBuilder[T]) HavingNotExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.HavingNotExists(query)
	return b
}

// OrWhereNotExists add an or not exists clause to the query.
func (b *ModelBuilder[T]) OrWhereNotExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereNotExists(query)
	return b
}

// OrHavingNotExists add an or not exists clause to the query.
func (b *ModelBuilder[T]) OrHavingNotExists(query dialects.QueryBuilder) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingNotExists(query)
	return b
}

// WhereSubquery adds a where clause to the query comparing a column and a subquery.
func (b *ModelBuilder[T]) WhereSubquery(subquery dialects.QueryBuilder, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.WhereSubquery(subquery, operator, value)
	return b
}

// HavingSubquery adds a having clause to the query comparing a column and a subquery.
func (b *ModelBuilder[T]) HavingSubquery(subquery dialects.QueryBuilder, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.HavingSubquery(subquery, operator, value)
	return b
}

// OrWhereSubquery adds an or where clause to the query comparing a column and a subquery.
func (b *ModelBuilder[T]) OrWhereSubquery(subquery dialects.QueryBuilder, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereSubquery(subquery, operator, value)
	return b
}

// OrHavingSubquery adds an or having clause to the query comparing a column and a subquery.
func (b *ModelBuilder[T]) OrHavingSubquery(subquery dialects.QueryBuilder, operator string, value any) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingSubquery(subquery, operator, value)
	return b
}

// WhereHas adds a relationship exists condition to the query with where clauses.
func (b *ModelBuilder[T]) WhereHas(relation string, cb func(q *Builder) *Builder) *ModelBuilder[T] {
	b.builder = b.builder.WhereHas(relation, cb)
	return b
}

// HavingHas adds a relationship exists condition to the query with having clauses.
func (b *ModelBuilder[T]) HavingHas(relation string, cb func(q *Builder) *Builder) *ModelBuilder[T] {
	b.builder = b.builder.HavingHas(relation, cb)
	return b
}

// OrWhereHas adds a relationship exists condition to the query with where clauses and an or.
func (b *ModelBuilder[T]) OrWhereHas(relation string, cb func(q *Builder) *Builder) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereHas(relation, cb)
	return b
}

// OrHavingHas adds a relationship exists condition to the query with having clauses and an or.
func (b *ModelBuilder[T]) OrHavingHas(relation string, cb func(q *Builder) *Builder) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingHas(relation, cb)
	return b
}

// WhereRaw adds a raw where clause to the query.
func (b *ModelBuilder[T]) WhereRaw(rawSql string, bindings ...any) *ModelBuilder[T] {
	b.builder = b.builder.WhereRaw(rawSql, bindings...)
	return b
}

// HavingRaw adds a raw having clause to the query.
func (b *ModelBuilder[T]) HavingRaw(rawSql string, bindings ...any) *ModelBuilder[T] {
	b.builder = b.builder.HavingRaw(rawSql, bindings...)
	return b
}

// OrWhereRaw adds a raw or where clause to the query.
func (b *ModelBuilder[T]) OrWhereRaw(rawSql string, bindings ...any) *ModelBuilder[T] {
	b.builder = b.builder.OrWhereRaw(rawSql, bindings...)
	return b
}

// OrHavingRaw adds a raw or having clause to the query.
func (b *ModelBuilder[T]) OrHavingRaw(rawSql string, bindings ...any) *ModelBuilder[T] {
	b.builder = b.builder.OrHavingRaw(rawSql, bindings...)
	return b
}

// And adds a group of conditions to the query
func (b *ModelBuilder[T]) And(cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.And(cb)
	return b
}

// HavingAnd adds a group of conditions to the query
func (b *ModelBuilder[T]) HavingAnd(cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.HavingAnd(cb)
	return b
}

// Or adds a group of conditions to the query with an or
func (b *ModelBuilder[T]) Or(cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.Or(cb)
	return b
}

// HavingOr adds a group of conditions to the query with an or
func (b *ModelBuilder[T]) HavingOr(cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.HavingOr(cb)
	return b
}

// WithScope adds a local scope to a query.
func (b *ModelBuilder[T]) WithScope(scope *Scope) *ModelBuilder[T] {
	b.builder = b.builder.WithScope(scope)
	return b
}

// WithoutScope removes the given scope from the local scopes.
func (b *ModelBuilder[T]) WithoutScope(scope *Scope) *ModelBuilder[T] {
	b.builder = b.builder.WithoutScope(scope)
	return b
}

// WithoutGlobalScope removes a global scope from the query.
func (b *ModelBuilder[T]) WithoutGlobalScope(scope *Scope) *ModelBuilder[T] {
	b.builder = b.builder.WithoutGlobalScope(scope)
	return b
}

// GroupBy sets the "group by" clause to the query.
func (b *ModelBuilder[T]) GroupBy(columns ...string) *ModelBuilder[T] {
	b.builder = b.builder.GroupBy(columns...)
	return b
}

// GroupBy adds a "group by" clause to the query.
func (b *ModelBuilder[T]) AddGroupBy(columns ...string) *ModelBuilder[T] {
	b.builder = b.builder.AddGroupBy(columns...)
	return b
}

// Join adds a join clause to the query.
func (b *ModelBuilder[T]) Join(table, localColumn, operator, foreignColumn string) *ModelBuilder[T] {
	b.builder = b.builder.Join(table, localColumn, operator, foreignColumn)
	return b
}

// LeftJoin adds a left join clause to the query.
func (b *ModelBuilder[T]) LeftJoin(table, localColumn, operator, foreignColumn string) *ModelBuilder[T] {
	b.builder = b.builder.LeftJoin(table, localColumn, operator, foreignColumn)
	return b
}

// RightJoin adds a right join clause to the query.
func (b *ModelBuilder[T]) RightJoin(table, localColumn, operator, foreignColumn string) *ModelBuilder[T] {
	b.builder = b.builder.RightJoin(table, localColumn, operator, foreignColumn)
	return b
}

// InnerJoin adds an inner join clause to the query.
func (b *ModelBuilder[T]) InnerJoin(table, localColumn, operator, foreignColumn string) *ModelBuilder[T] {
	b.builder = b.builder.InnerJoin(table, localColumn, operator, foreignColumn)
	return b
}

// CrossJoin adds a cross join clause to the query.
func (b *ModelBuilder[T]) CrossJoin(table, localColumn, operator, foreignColumn string) *ModelBuilder[T] {
	b.builder = b.builder.CrossJoin(table, localColumn, operator, foreignColumn)
	return b
}

// JoinOn adds a join clause to the query with a complex on statement.
func (b *ModelBuilder[T]) JoinOn(table string, cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.JoinOn(table, cb)
	return b
}

// LeftJoinOn adds a left join clause to the query with a complex on statement.
func (b *ModelBuilder[T]) LeftJoinOn(table string, cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.LeftJoinOn(table, cb)
	return b
}

// RightJoinOn adds a right join clause to the query with a complex on statement.
func (b *ModelBuilder[T]) RightJoinOn(table string, cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.RightJoinOn(table, cb)
	return b
}

// InnerJoinOn adds an inner join clause to the query with a complex on statement.
func (b *ModelBuilder[T]) InnerJoinOn(table string, cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.InnerJoinOn(table, cb)
	return b
}

// CrossJoinOn adds a cross join clause to the query with a complex on statement.
func (b *ModelBuilder[T]) CrossJoinOn(table string, cb func(q *Conditions)) *ModelBuilder[T] {
	b.builder = b.builder.CrossJoinOn(table, cb)
	return b
}

// Limit set the maximum number of rows to return.
func (b *ModelBuilder[T]) Limit(limit int) *ModelBuilder[T] {
	b.builder = b.builder.Limit(limit)
	return b
}

// Offset sets the number of rows to skip before returning the result.
func (b *ModelBuilder[T]) Offset(offset int) *ModelBuilder[T] {
	b.builder = b.builder.Offset(offset)
	return b
}

// OrderBy adds an order by clause to the query.
func (b *ModelBuilder[T]) OrderBy(column string) *ModelBuilder[T] {
	b.builder = b.builder.OrderBy(column)
	return b
}

// OrderByDesc adds a descending order by clause to the query.
func (b *ModelBuilder[T]) OrderByDesc(column string) *ModelBuilder[T] {
	b.builder = b.builder.OrderByDesc(column)
	return b
}

// Unordered removes all order by clauses from the query.
func (b *ModelBuilder[T]) Unordered() *ModelBuilder[T] {
	b.builder = b.builder.Unordered()
	return b
}

// Select sets the columns to be selected.
func (b *ModelBuilder[T]) Select(columns ...string) *ModelBuilder[T] {
	b.builder = b.builder.Select(columns...)
	return b
}

// Select sets the columns to be selected.
func (b *ModelBuilder[T]) SelectRaw(columns ...string) *ModelBuilder[T] {
	b.builder = b.builder.SelectRaw(columns...)
	return b
}

// AddSelect adds new columns to be selected.
func (b *ModelBuilder[T]) AddSelect(columns ...string) *ModelBuilder[T] {
	b.builder = b.builder.AddSelect(columns...)
	return b
}

// AddSelect adds new columns to be selected.
func (b *ModelBuilder[T]) AddSelectRaw(columns ...string) *ModelBuilder[T] {
	b.builder = b.builder.AddSelectRaw(columns...)
	return b
}

// SelectSubquery sets a subquery to be selected.
func (b *ModelBuilder[T]) SelectSubquery(sb dialects.QueryBuilder, as string) *ModelBuilder[T] {
	b.builder = b.builder.SelectSubquery(sb, as)
	return b
}

// AddSelectSubquery adds a subquery to be selected.
func (b *ModelBuilder[T]) AddSelectSubquery(sb dialects.QueryBuilder, as string) *ModelBuilder[T] {
	b.builder = b.builder.AddSelectSubquery(sb, as)
	return b
}

// SelectFunction sets a column to be selected with a function applied.
func (b *ModelBuilder[T]) SelectFunction(function, column string) *ModelBuilder[T] {
	b.builder = b.builder.SelectFunction(function, column)
	return b
}

// SelectFunction adds a column to be selected with a function applied.
func (b *ModelBuilder[T]) AddSelectFunction(function, column string) *ModelBuilder[T] {
	b.builder = b.builder.AddSelectFunction(function, column)
	return b
}

// Distinct forces the query to only return distinct results.
func (b *ModelBuilder[T]) Distinct() *ModelBuilder[T] {
	b.builder = b.builder.Distinct()
	return b
}

// Dump prints the encoded SQL statement for the query to stdout and returns
// the receiver so it can be used in the middle of a chain.
func (b *ModelBuilder[T]) Dump() *ModelBuilder[T] {
	b.builder = b.builder.Dump()
	return b
}
