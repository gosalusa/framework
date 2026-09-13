package schema

import (
	"context"

	"gosalusa.com/database"
	"gosalusa.com/database/dialects"
)

// Drop returns a Runner that drops the named table.
func Drop(table string) Runner {
	return Run(func(ctx context.Context, tx database.DB) error {
		return runDropTable(ctx, tx, &dialects.DropTableQuery{Table: table})
	})
}

// DropIfExists returns a Runner that drops the named table if it exists.
func DropIfExists(table string) Runner {
	return Run(func(ctx context.Context, tx database.DB) error {
		return runDropTable(ctx, tx, &dialects.DropTableQuery{Table: table, IfExists: true})
	})
}
func runDropTable(ctx context.Context, tx database.DB, q *dialects.DropTableQuery) error {
	d, err := dialects.New(tx.DriverName())
	if err != nil {
		return err
	}
	result, err := d.EncodeDropTableQuery(q)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, result.SQL, result.Bindings...)
	return err
}
