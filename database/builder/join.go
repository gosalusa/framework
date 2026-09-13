package builder

import (
	"gosalusa.com/database/dialects"
)

// Join adds a join clause to the query.
func (b *Builder) Join(table, localColumn, operator, foreignColumn string) *Builder {
	return b.join("", table, localColumn, operator, foreignColumn)
}

// LeftJoin adds a left join clause to the query.
func (b *Builder) LeftJoin(table, localColumn, operator, foreignColumn string) *Builder {
	return b.join("LEFT", table, localColumn, operator, foreignColumn)
}

// RightJoin adds a right join clause to the query.
func (b *Builder) RightJoin(table, localColumn, operator, foreignColumn string) *Builder {
	return b.join("RIGHT", table, localColumn, operator, foreignColumn)
}

// InnerJoin adds an inner join clause to the query.
func (b *Builder) InnerJoin(table, localColumn, operator, foreignColumn string) *Builder {
	return b.join("INNER", table, localColumn, operator, foreignColumn)
}

// CrossJoin adds a cross join clause to the query.
func (b *Builder) CrossJoin(table, localColumn, operator, foreignColumn string) *Builder {
	return b.join("CROSS", table, localColumn, operator, foreignColumn)
}
func (b *Builder) join(direction, table, localColumn, operator, foreignColumn string) *Builder {
	return b.joinOn(direction, table, func(q *Conditions) {
		q.WhereColumn(localColumn, operator, foreignColumn)
	})
}

// JoinOn adds a join clause to the query with a complex on statement.
func (b *Builder) JoinOn(table string, cb func(q *Conditions)) *Builder {
	return b.joinOn("", table, cb)
}

// LeftJoinOn adds a left join clause to the query with a complex on statement.
func (b *Builder) LeftJoinOn(table string, cb func(q *Conditions)) *Builder {
	return b.joinOn("LEFT", table, cb)
}

// RightJoinOn adds a right join clause to the query with a complex on statement.
func (b *Builder) RightJoinOn(table string, cb func(q *Conditions)) *Builder {
	return b.joinOn("RIGHT", table, cb)
}

// InnerJoinOn adds an inner join clause to the query with a complex on statement.
func (b *Builder) InnerJoinOn(table string, cb func(q *Conditions)) *Builder {
	return b.joinOn("INNER", table, cb)
}

// CrossJoinOn adds a cross join clause to the query with a complex on statement.
func (b *Builder) CrossJoinOn(table string, cb func(q *Conditions)) *Builder {
	return b.joinOn("CROSS", table, cb)
}
func (b *Builder) joinOn(direction string, table string, cb func(q *Conditions)) *Builder {
	c := NewConditionBuilder()
	cb(c)
	b.query.Joins = append(b.query.Joins, dialects.Join{
		Direction:  direction,
		Table:      table,
		Conditions: c.conditions,
	})
	return b
}
