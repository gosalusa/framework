package schema

import (
	"context"

	"gosalusa.com/database"
)

// Runner is any database schema change that executes against a database.DB,
// usually a transaction. Create and Table operations, the results of Drop and
// DropIfExists, View, and Raw all implement it.
type Runner interface {
	Run(ctx context.Context, tx database.DB) error
}

// RunnerFunc adapts a function to the Runner interface.
type RunnerFunc func(ctx context.Context, tx database.DB) error

// Run implements [Runner].
func (f RunnerFunc) Run(ctx context.Context, tx database.DB) error {
	return f(ctx, tx)
}

// Run wraps f as a Runner so it can be used anywhere a schema operation is
// expected.
func Run(f RunnerFunc) Runner {
	return f
}
