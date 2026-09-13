package builder

import (
	"context"

	"gosalusa.com/database"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/model"
	"gosalusa.com/extra/sets"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/internal/relationship"
)

// Builder is a query builder that constructs SQL statements without any
// knowledge of a specific model. It exposes the query operations used by
// ModelBuilder and can be built up directly, for example to pass a subquery to
// WhereExists or WhereSubquery.
//
//go:generate go run ../../internal/build/build.go
type Builder struct {
	query   dialects.SelectQuery
	wheres  *Conditions
	havings *Conditions

	scopes *scopes
	ctx    context.Context
}

var _ dialects.QueryBuilder = (*Builder)(nil)

// NewBuilder creates a new Builder with no columns selected.
func NewBuilder() *Builder {
	return &Builder{
		query:   dialects.NewSelectQuery(),
		wheres:  NewConditionBuilder(),
		havings: NewConditionBuilder(),
		scopes:  newScopes(),
		ctx:     context.Background(),
	}
}

// Query implements [dialects.QueryBuilder].
func (b *Builder) Query() *dialects.SelectQuery {
	current := b.Clone()
	for _, s := range b.ActiveScopes() {
		if s.Query != nil {
			current = s.Query(current)
		}
	}
	q := &current.query
	q.Wheres = current.wheres.conditions
	q.Havings = current.havings.conditions
	return q
}

// ModelBuilder represents a query bound to a model type T, along with the
// relationships to eager load and the scopes to apply.
//
//go:generate go run ../../internal/build/build.go
type ModelBuilder[T model.Model] struct {
	builder       *Builder
	withs         []string
	withoutScopes sets.Set[string]
}

// New creates a new query from the model's table with * selected.
func New[T model.Model]() *ModelBuilder[T] {
	return NewEmpty[T]().Select("*")
}

// From creates a new query from the model's table and with table.* selected.
func From[T model.Model]() *ModelBuilder[T] {
	var m T
	table := database.GetTable(m)
	return NewEmpty[T]().Select(table + ".*").From(table)
}

// NewEmpty creates a new query with nothing selected and without a table set.
func NewEmpty[T model.Model]() *ModelBuilder[T] {
	m := helpers.CreateFor[T]().Interface().(T)

	_ = relationship.InitializeRelationships(m)

	sb := NewBuilder()
	sb.wheres.withParent(m)
	sb.havings.withParent(m)
	sb.scopes.withParent(m)

	return &ModelBuilder[T]{
		builder:       sb,
		withs:         []string{},
		withoutScopes: sets.New[string](),
	}
}

// Query implements [dialects.QueryBuilder].
func (b *ModelBuilder[T]) Query() *dialects.SelectQuery {
	return b.builder.Query()
}
