package schema

import (
	"context"
	"fmt"

	"gosalusa.com/database"
)

// ViewBuilder builds a CREATE VIEW operation.
type ViewBuilder struct {
	// Name is the name of the view.
	Name string
	// Query is the SELECT statement that defines the view.
	Query string
}

// View returns a ViewBuilder for a view named name backed by the given query.
func View(name string, query string) *ViewBuilder {
	return &ViewBuilder{
		Name:  name,
		Query: query,
	}
}

// GoString renders the builder as a schema.View(...) call.
func (b *ViewBuilder) GoString() string {
	return fmt.Sprintf(
		"schema.View(%#v, %#v)",
		b.Name,
		b.Query,
	)
}
// Run executes the CREATE VIEW statement against the transaction.
func (b *ViewBuilder) Run(ctx context.Context, tx database.DB) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("CREATE VIEW %s as %s", b.Name, b.Query))
	return err
}
