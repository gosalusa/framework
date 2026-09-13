package schema

import (
	"context"
	"fmt"

	"gosalusa.com/database"
	"gosalusa.com/database/dialects"
)

// CreateTableBuilder builds a CREATE TABLE operation. Create returns one, and
// it may be configured with IfNotExists or Temporary before Run executes it.
// It implements Blueprinter and Runner.
type CreateTableBuilder struct {
	blueprint   *Blueprint
	ifNotExists bool
	temporary   bool
}

var _ Blueprinter = &CreateTableBuilder{}

// Create starts a CREATE TABLE operation for the named table. The callback
// populates the table's Blueprint with columns, indexes, and constraints.
func Create(name string, cb func(b *Blueprint)) *CreateTableBuilder {
	b := NewBlueprint(name)
	cb(b)
	return &CreateTableBuilder{
		blueprint: b,
	}
}

// GetBlueprint returns the Blueprint describing the table.
func (b *CreateTableBuilder) GetBlueprint() *Blueprint {
	return b.blueprint
}

// Type returns [BlueprintTypeCreate].
func (b *CreateTableBuilder) Type() BlueprintType {
	return BlueprintTypeCreate
}

// CreateTableQuery converts the builder into a dialects.CreateTableQuery that
// a dialect can encode into SQL.
func (b *CreateTableBuilder) CreateTableQuery() *dialects.CreateTableQuery {
	columns := make([]dialects.ColumnDefinition, len(b.blueprint.columns))
	for i, c := range b.blueprint.columns {
		columns[i] = *c.ColumnDefinition()
	}

	foreignKeys := make([]dialects.ForeignKey, len(b.blueprint.foreignKeys))
	for i, fk := range b.blueprint.foreignKeys {
		foreignKeys[i] = *fk.ForeignKey()
	}

	indexes := make([]dialects.Index, len(b.blueprint.indexes))
	for i, fk := range b.blueprint.indexes {
		indexes[i] = *fk.Index()
	}

	return &dialects.CreateTableQuery{
		IfNotExists: b.ifNotExists,
		Temporary:   b.temporary,
		Table:       b.blueprint.TableName(),
		Columns:     columns,
		PrimaryKeys: b.blueprint.primaryKeys,
		ForeignKeys: foreignKeys,
		Indexes:     indexes,
	}
}

// GoString renders the operation as a schema.Create(...) call.
func (b *CreateTableBuilder) GoString() string {
	return fmt.Sprintf(
		"schema.Create(%#v, %#v)",
		b.blueprint.name,
		b.blueprint,
	)
}

// Run selects the dialect for the transaction's driver and executes the
// rendered CREATE TABLE statement.
func (b *CreateTableBuilder) Run(ctx context.Context, tx database.DB) error {
	q := b.CreateTableQuery()
	d, err := dialects.New(tx.DriverName())
	if err != nil {
		return err
	}
	result, err := d.EncodeCreateTableQuery(q)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, result.SQL, result.Bindings...)
	return err
}
// IfNotExists causes the statement to be generated with IF NOT EXISTS.
func (b *CreateTableBuilder) IfNotExists() *CreateTableBuilder {
	b.ifNotExists = true
	return b
}

// Temporary causes a temporary table to be created.
func (b *CreateTableBuilder) Temporary() *CreateTableBuilder {
	b.temporary = true
	return b
}
