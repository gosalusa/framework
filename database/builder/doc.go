// Package builder provides a fluent, chainable query builder for the database
// package. It lets you construct and execute SELECT, UPDATE, and DELETE
// statements against models without writing raw SQL.
//
// There are two types of builders. Builder represents a query before any model
// is attached and exposes the full set of query operations, while the generic
// ModelBuilder[T] is bound to a model type and adds the type-safe terminal
// operations for loading, counting, updating, and deleting records.
//
// Start a query with New, From, or NewEmpty and chain methods to apply
// conditions, joins, ordering, and pagination:
//
//	users, err := builder.From[User]().
//		Where("active", "=", true).
//		OrderByDesc("name").
//		Limit(10).
//		Get(db)
//
// Relationships declared as fields on a model can be eager loaded with With,
// constrained with WhereHas, or loaded after the fact with Load and
// LoadMissing.
//
// Builder methods generally mutate the receiver and return it so calls can be
// chained. Use Clone to obtain an independent copy of a query before applying
// further modifications.
package builder
