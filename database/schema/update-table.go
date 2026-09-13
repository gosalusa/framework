package schema

import (
	"context"
	"fmt"

	"gosalusa.com/database"
	"gosalusa.com/database/dialects"
)

// UpdateTableBuilder builds an ALTER TABLE operation. Table returns one and
// Run executes it. Columns not marked with Change are added, columns marked
// with Change are modified, DropColumn names columns to remove, and ForeignKey
// and Index append new constraints. It implements Blueprinter and Runner.
type UpdateTableBuilder struct {
	blueprint *Blueprint
}

var _ Blueprinter = &UpdateTableBuilder{}

// Table starts an ALTER TABLE operation for the named table. The callback
// populates the Blueprint with the changes to apply.
func Table(name string, cb func(table *Blueprint)) *UpdateTableBuilder {
	table := NewBlueprint(name)
	cb(table)
	return &UpdateTableBuilder{
		blueprint: table,
	}
}

// GetBlueprint returns the Blueprint describing the changes.
func (b *UpdateTableBuilder) GetBlueprint() *Blueprint {
	return b.blueprint
}

// Type returns [BlueprintTypeUpdate].
func (b *UpdateTableBuilder) Type() BlueprintType {
	return BlueprintTypeUpdate
}

// AlterTableQuery converts the builder into a dialects.AlterTableQuery that a
// dialect can encode into SQL.
func (b *UpdateTableBuilder) AlterTableQuery() *dialects.AlterTableQuery {
	addColumns := make([]dialects.ColumnDefinition, 0, len(b.blueprint.columns))
	modifyColumns := make([]dialects.ColumnDefinition, 0, len(b.blueprint.columns))

	for _, column := range b.blueprint.columns {
		if column.change {
			modifyColumns = append(modifyColumns, *column.ColumnDefinition())
		} else {
			addColumns = append(addColumns, *column.ColumnDefinition())
		}
	}
	foreignKeys := make([]dialects.ForeignKey, 0, len(b.blueprint.foreignKeys))
	for _, foreignKey := range b.blueprint.foreignKeys {
		foreignKeys = append(foreignKeys, *foreignKey.ForeignKey())
	}
	indexes := make([]dialects.Index, 0, len(b.blueprint.indexes))
	for _, index := range b.blueprint.indexes {
		indexes = append(indexes, *index.Index())
	}

	return &dialects.AlterTableQuery{
		Table:         b.blueprint.TableName(),
		DropColumns:   b.blueprint.dropColumns,
		AddColumns:    addColumns,
		ModifyColumns: modifyColumns,
		ForeignKeys:   foreignKeys,
		Indexes:       indexes,
	}
}
// GoString renders the operation as a schema.Table(...) call.
func (b *UpdateTableBuilder) GoString() string {
	return fmt.Sprintf(
		"schema.Table(%#v, %#v)",
		b.blueprint.name,
		b.blueprint,
	)
}

// Run selects the dialect for the transaction's driver and executes the
// rendered ALTER TABLE statement.
func (b *UpdateTableBuilder) Run(ctx context.Context, tx database.DB) error {
	d, err := dialects.New(tx.DriverName())
	if err != nil {
		return err
	}
	result, err := d.EncodeAlterTableQuery(b.AlterTableQuery())
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, result.SQL, result.Bindings...)
	return err
}
