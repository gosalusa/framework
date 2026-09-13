package builder

import (
	"context"
	"reflect"

	"gosalusa.com/database/dialects"
)

// Conditions is a mutable collection of where clauses that can be built up
// with the same methods exposed on Builder. It is used to build where and
// having clauses as well as the on conditions of a join.
type Conditions struct {
	conditions []dialects.Condition
	parent     any
	ctx        context.Context
}

// Clone returns a copy of the conditions. Mutating the returned conditions, or
// the original, will not affect the other.
func (c *Conditions) Clone() *Conditions {
	return &Conditions{
		conditions: cloneSlice(c.conditions),
		parent:     c.parent,
		ctx:        c.ctx,
	}
}
func (c *Conditions) withParent(parent any) *Conditions {
	c.parent = parent
	return c
}

// NewConditionBuilder returns a new empty Conditions.
func NewConditionBuilder() *Conditions {
	return &Conditions{
		conditions: []dialects.Condition{},
	}
}

// Build returns the conditions that have been added.
func (c *Conditions) Build() []dialects.Condition {
	return c.conditions
}

// Where adds a basic where clause to the query.
func (b *Conditions) Where(column, operator string, value any) *Conditions {
	return b.where(column, operator, value, false)
}

// OrWhere adds an or where clause to the query
func (b *Conditions) OrWhere(column, operator string, value any) *Conditions {
	return b.where(column, operator, value, true)
}

// WhereColumn adds a where clause to the query comparing two columns.
func (b *Conditions) WhereColumn(column, operator string, valueColumn string) *Conditions {
	return b.where(column, operator, dialects.Column{Column: valueColumn}, false)
}

// OrWhereColumn adds an or where clause to the query comparing two columns.
func (b *Conditions) OrWhereColumn(column, operator string, valueColumn string) *Conditions {
	return b.where(column, operator, dialects.Column{Column: valueColumn}, true)
}

// WhereIn adds a where in clause to the query.
func (b *Conditions) WhereIn(column string, values []any) *Conditions {
	return b.whereIn(column, values, false)
}

// OrWhereIn adds an or where in clause to the query.
func (b *Conditions) OrWhereIn(column string, values []any) *Conditions {
	return b.whereIn(column, values, true)
}

func (b *Conditions) whereIn(column string, values []any, or bool) *Conditions {
	return b.where(column, "in", values, or)
}

// WhereExists add an exists clause to the query.
func (b *Conditions) WhereExists(query dialects.QueryBuilder) *Conditions {
	return b.whereExists(query, false)
}

// OrWhereExists add an or exists clause to the query.
func (b *Conditions) OrWhereExists(query dialects.QueryBuilder) *Conditions {
	return b.whereExists(query, true)
}

func (b *Conditions) whereExists(query dialects.QueryBuilder, or bool) *Conditions {
	b.conditions = append(b.conditions, dialects.Condition{
		Operator: "EXISTS",
		Value:    query,
		Or:       or,
	})
	return b
}

// WhereNotExists add a not exists clause to the query.
func (b *Conditions) WhereNotExists(query dialects.QueryBuilder) *Conditions {
	return b.whereNotExists(query, false)
}

// OrWhereNotExists add an or not exists clause to the query.
func (b *Conditions) OrWhereNotExists(query dialects.QueryBuilder) *Conditions {
	return b.whereNotExists(query, true)
}

func (b *Conditions) whereNotExists(query dialects.QueryBuilder, or bool) *Conditions {
	b.conditions = append(b.conditions, dialects.Condition{
		Operator: "NOT EXISTS",
		Value:    query,
		Or:       or,
	})
	return b
}

// WhereSubquery adds a where clause to the query comparing a column and a subquery.
func (b *Conditions) WhereSubquery(subquery dialects.QueryBuilder, operator string, value any) *Conditions {
	return b.whereSubquery(subquery, operator, value, false)
}

// OrWhereSubquery adds an or where clause to the query comparing a column and a subquery.
func (b *Conditions) OrWhereSubquery(subquery dialects.QueryBuilder, operator string, value any) *Conditions {
	return b.whereSubquery(subquery, operator, value, true)
}

func (b *Conditions) whereSubquery(subquery dialects.QueryBuilder, operator string, value any, or bool) *Conditions {
	b.conditions = append(b.conditions, dialects.Condition{
		Column:   dialects.Column{SubQuery: subquery},
		Operator: operator,
		Value:    value,
		Or:       or,
	})
	return b
}

func (b *Conditions) where(column, operator string, value any, or bool) *Conditions {
	b.conditions = append(b.conditions, dialects.Condition{
		Column:   dialects.Column{Column: column},
		Operator: operator,
		Value:    value,
		Or:       or,
	})
	return b
}

// WhereHas adds a relationship exists condition to the query with where clauses.
func (b *Conditions) WhereHas(relation string, cb func(q *Builder) *Builder) *Conditions {
	return b.whereHas(relation, cb, false)
}

// OrWhereHas adds a relationship exists condition to the query with where clauses and an or.
func (b *Conditions) OrWhereHas(relation string, cb func(q *Builder) *Builder) *Conditions {
	return b.whereHas(relation, cb, true)
}
func (b *Conditions) whereHas(relation string, cb func(b *Builder) *Builder, or bool) *Conditions {
	r, ok := getRelation(reflect.ValueOf(b.parent), relation)
	if !ok {
		return b.whereExists(cb(NewBuilder().WithContext(b.ctx)), or)
	}

	return b.whereExists(cb(r.Subquery().WithContext(b.ctx)), or)
}

// WhereRaw adds a raw where clause to the query.
func (b *Conditions) WhereRaw(rawSql string, bindings ...any) *Conditions {
	return b.whereRaw(rawSql, bindings, false)
}

// OrWhereRaw adds a raw or where clause to the query.
func (b *Conditions) OrWhereRaw(rawSql string, bindings ...any) *Conditions {
	return b.whereRaw(rawSql, bindings, true)
}
func (b *Conditions) whereRaw(rawSql string, bindings []any, or bool) *Conditions {
	b.conditions = append(b.conditions, dialects.Condition{
		Value: dialects.Raw(rawSql, bindings...),
		Or:    or,
	})
	return b
}

// And adds a group of conditions to the query
func (b *Conditions) And(cb func(q *Conditions)) *Conditions {
	return b.group(cb, false)
}

// Or adds a group of conditions to the query with an or
func (b *Conditions) Or(cb func(q *Conditions)) *Conditions {
	return b.group(cb, true)
}

func (b *Conditions) group(fn func(q *Conditions), or bool) *Conditions {
	c := &Conditions{
		conditions: []dialects.Condition{},
	}
	fn(c)
	b.conditions = append(b.conditions, dialects.Condition{
		Value: c.conditions,
		Or:    or,
	})
	return b
}
