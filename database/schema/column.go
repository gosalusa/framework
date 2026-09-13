package schema

import (
	"fmt"

	"gosalusa.com/database/dialects"
)

// ColumnBuilder describes a single column and collects its modifiers. It is
// created by the Blueprint shorthand methods or NewColumn, and modifiers such
// as Nullable and Default mutate it and return it so calls can be chained.
type ColumnBuilder struct {
	name     string
	datatype dialects.DataType

	nullable           bool
	primary            bool
	autoIncrement      bool
	change             bool
	defaultValue       any
	afterColumn        string
	unique             bool
	defaultCurrentTime bool
	index              bool
}

// NewColumn returns a column builder with the given name and datatype.
func NewColumn(name string, datatype dialects.DataType) *ColumnBuilder {
	return &ColumnBuilder{
		name:     name,
		datatype: datatype,
	}
}

// Equals reports whether b describes the same column definition as newB. It
// compares the name, datatype, nullability, auto-increment, primary-key, and
// index settings.
func (b *ColumnBuilder) Equals(newB *ColumnBuilder) bool {
	return b.datatype == newB.datatype &&
		b.nullable == newB.nullable &&
		b.autoIncrement == newB.autoIncrement &&
		b.index == newB.index &&
		b.name == newB.name &&
		b.primary == newB.primary
}

// Name returns the column name.
func (b *ColumnBuilder) Name() string {
	return b.name
}

// Nullable allows the column to store NULL.
func (b *ColumnBuilder) Nullable() *ColumnBuilder {
	b.nullable = true
	return b
}

// NotNullable forbids NULL in the column. Columns are not nullable by default.
func (b *ColumnBuilder) NotNullable() *ColumnBuilder {
	b.nullable = false
	return b
}

// Primary marks the column as the primary key. For a composite primary key,
// use [Blueprint.PrimaryKey].
func (b *ColumnBuilder) Primary() *ColumnBuilder {
	b.primary = true
	return b
}

// AutoIncrement makes the database populate the column automatically on
// insert. It is typically combined with Primary.
func (b *ColumnBuilder) AutoIncrement() *ColumnBuilder {
	b.autoIncrement = true
	return b
}

// After positions the column after the named column. It is only supported by
// MySQL.
func (b *ColumnBuilder) After(column string) *ColumnBuilder {
	b.afterColumn = column
	return b
}

// Change marks the column as a modification of an existing column when used
// with an [UpdateTableBuilder], producing ALTER TABLE ... MODIFY instead of
// ADD.
func (b *ColumnBuilder) Change() *ColumnBuilder {
	b.change = true
	return b
}

// Default sets a constant default value for the column.
func (b *ColumnBuilder) Default(v any) *ColumnBuilder {
	b.defaultValue = v
	return b
}

// Type overrides the column's data type.
func (b *ColumnBuilder) Type(datatype dialects.DataType) *ColumnBuilder {
	b.datatype = datatype
	return b
}

// Unique adds a unique constraint to the column.
func (b *ColumnBuilder) Unique() *ColumnBuilder {
	b.unique = true
	return b
}

// DefaultCurrentTime sets the column default to CURRENT_TIMESTAMP.
func (b *ColumnBuilder) DefaultCurrentTime() *ColumnBuilder {
	b.defaultCurrentTime = true
	return b
}

// Index creates an index on the column.
func (b *ColumnBuilder) Index() *ColumnBuilder {
	b.index = true
	return b
}

// Size sets the column size, for example the length of a string column.
func (b *ColumnBuilder) Size(s int) *ColumnBuilder {
	b.datatype.Size = s
	return b
}

// ColumnDefinition converts the column into its dialects.ColumnDefinition, the
// query-level representation that a dialect encodes into SQL.
func (b *ColumnBuilder) ColumnDefinition() *dialects.ColumnDefinition {
	return &dialects.ColumnDefinition{
		Name:               b.name,
		Datatype:           b.datatype,
		Nullable:           b.nullable,
		Primary:            b.primary,
		AutoIncrement:      b.autoIncrement,
		DefaultValue:       b.defaultValue,
		Unique:             b.unique,
		DefaultCurrentTime: b.defaultCurrentTime,
	}
}

// GoString renders the column's modifiers as a chain of Go method calls.
func (b *ColumnBuilder) GoString() string {
	src := ""
	if b.primary {
		src += ".Primary()"
	}
	if b.autoIncrement {
		src += ".AutoIncrement()"
	}
	if b.nullable {
		src += ".Nullable()"
	}
	if b.defaultValue != nil {
		src += fmt.Sprintf(".Default(%#v)", b.defaultValue)
	}
	if b.defaultCurrentTime {
		src += ".DefaultCurrentTime()"
	}
	if b.unique {
		src += ".Unique()"
	}
	if b.index {
		src += ".Index()"
	}
	if b.change {
		src += ".Change()"
	}
	if b.datatype.Size != 0 {
		src += fmt.Sprintf(".Size(%#v)", b.datatype.Size)
	}
	return src
}

// Columns replaces the blueprint's columns with the given ones.
func (b *CreateTableBuilder) Columns(columns ...*ColumnBuilder) *CreateTableBuilder {
	b.blueprint.columns = columns
	return b
}

// AddColumns appends the given columns to the blueprint.
func (b *CreateTableBuilder) AddColumns(columns ...*ColumnBuilder) *CreateTableBuilder {
	b.blueprint.columns = append(b.blueprint.columns, columns...)
	return b
}
