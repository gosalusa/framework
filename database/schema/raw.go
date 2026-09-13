package schema

import (
	"context"

	"gosalusa.com/database"
)

// Raw is a schema operation containing a raw SQL statement to execute as-is.
type Raw string

// Run implements [Runner].
func (v Raw) Run(ctx context.Context, tx database.DB) error {
	_, err := tx.ExecContext(ctx, string(v))
	return err
}
