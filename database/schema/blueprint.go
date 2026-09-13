package schema

import (
	"fmt"
	"slices"
	"strings"

	"gosalusa.com/database/dialects"
	"gosalusa.com/extra/sets"
	"gosalusa.com/stream"
)

// BlueprintType distinguishes a table creation from a table alteration.
type BlueprintType int

const (
	// BlueprintTypeCreate marks a Blueprint used to create a table.
	BlueprintTypeCreate BlueprintType = iota
	// BlueprintTypeUpdate marks a Blueprint used to alter an existing table.
	BlueprintTypeUpdate
)

// Blueprinter is implemented by table builders. Both CreateTableBuilder and
// UpdateTableBuilder satisfy it.
type Blueprinter interface {
	GetBlueprint() *Blueprint
	Type() BlueprintType
}

// Blueprint describes a table: its columns, dropped columns, indexes, foreign
// keys, and primary key. A Blueprint is usually created with NewBlueprint and
// populated inside the callback passed to Create or Table.
type Blueprint struct {
	name        string
	columns     []*ColumnBuilder
	dropColumns []string
	indexes     []*IndexBuilder
	foreignKeys []*ForeignKeyBuilder
	primaryKeys []string
}

// NewBlueprint returns an empty Blueprint for the named table.
func NewBlueprint(name string) *Blueprint {
	return &Blueprint{
		name:        name,
		columns:     []*ColumnBuilder{},
		dropColumns: []string{},
		indexes:     []*IndexBuilder{},
		foreignKeys: []*ForeignKeyBuilder{},
	}
}

func (b *Blueprint) findColumn(name string) (*ColumnBuilder, bool) {
	return stream.Of(b.columns).Find(func(c *ColumnBuilder) bool {
		return c.name == name
	})
}

// GetBlueprint returns the Blueprint underlying the builder.
func (t *Blueprint) GetBlueprint() *Blueprint {
	return t
}

// TableName returns the name of the table the blueprint describes.
func (t *Blueprint) TableName() string {
	return t.name
}

// OfType adds a column of the given datatype to the blueprint and returns its
// builder. The shorthand methods such as String and Int call this.
func (t *Blueprint) OfType(datatype dialects.DataType, name string) *ColumnBuilder {
	c := NewColumn(name, datatype)
	t.AddColumn(c)
	return c
}

// AddColumn appends a column to the blueprint.
func (t *Blueprint) AddColumn(c *ColumnBuilder) *Blueprint {
	t.columns = append(t.columns, c)
	return t
}

// String adds a string column and returns its builder.
func (t *Blueprint) String(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeString, name)
}

// Text adds a text column and returns its builder.
func (t *Blueprint) Text(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeText, name)
}

// Bool adds a boolean column and returns its builder.
func (t *Blueprint) Bool(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeBoolean, name)
}

// Int adds a 32-bit signed integer column and returns its builder.
func (t *Blueprint) Int(name string) *ColumnBuilder {
	return t.Int32(name)
}

// Int8 adds an 8-bit signed integer column and returns its builder.
func (t *Blueprint) Int8(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeInt8, name)
}

// Int16 adds a 16-bit signed integer column and returns its builder.
func (t *Blueprint) Int16(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeInt16, name)
}

// Int32 adds a 32-bit signed integer column and returns its builder.
func (t *Blueprint) Int32(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeInt32, name)
}

// Int64 adds a 64-bit signed integer column and returns its builder.
func (t *Blueprint) Int64(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeInt64, name)
}

// UInt adds a 32-bit unsigned integer column and returns its builder.
func (t *Blueprint) UInt(name string) *ColumnBuilder {
	return t.UInt32(name)
}

// UInt8 adds an 8-bit unsigned integer column and returns its builder.
func (t *Blueprint) UInt8(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeUInt8, name)
}

// UInt16 adds a 16-bit unsigned integer column and returns its builder.
func (t *Blueprint) UInt16(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeUInt16, name)
}

// UInt32 adds a 32-bit unsigned integer column and returns its builder.
func (t *Blueprint) UInt32(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeUInt32, name)
}

// UInt64 adds a 64-bit unsigned integer column and returns its builder.
func (t *Blueprint) UInt64(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeUInt64, name)
}

// Float adds a 32-bit floating point column and returns its builder.
func (t *Blueprint) Float(name string) *ColumnBuilder {
	return t.Float32(name)
}

// Float32 adds a 32-bit floating point column and returns its builder.
func (t *Blueprint) Float32(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeFloat32, name)
}

// Float64 adds a 64-bit floating point column and returns its builder.
func (t *Blueprint) Float64(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeFloat64, name)
}

// JSON adds a JSON column and returns its builder.
func (t *Blueprint) JSON(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeJSON, name)
}

// Date adds a date column and returns its builder.
func (t *Blueprint) Date(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeDate, name)
}

// DateTime adds a date-time column and returns its builder.
func (t *Blueprint) DateTime(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeDateTime, name)
}

// Blob adds a blob column and returns its builder.
func (t *Blueprint) Blob(name string) *ColumnBuilder {
	return t.OfType(dialects.DataTypeBlob, name)
}

// Index begins describing an index on the table and returns its builder. Add
// columns to the index with IndexBuilder.AddColumn, and call
// IndexBuilder.Unique for a unique index.
func (t *Blueprint) Index(name string) *IndexBuilder {
	c := newIndexBuilder(t.TableName())
	c.name = name
	t.indexes = append(t.indexes, c)
	return c
}

// ForeignKey adds a foreign-key constraint making localKey reference relatedKey
// in relatedTable. The constraint is named
// "<localKey>-<relatedTable>-<relatedKey>".
func (t *Blueprint) ForeignKey(localKey, relatedTable, relatedKey string) {
	f := &ForeignKeyBuilder{
		localKey:     localKey,
		relatedTable: relatedTable,
		relatedKey:   relatedKey,
	}
	t.foreignKeys = append(t.foreignKeys, f)
}

// PrimaryKey sets the primary key of the table to the given columns. Use it
// for a composite primary key; for a single-column key, prefer
// [ColumnBuilder.Primary].
func (t *Blueprint) PrimaryKey(columns ...string) {
	t.primaryKeys = columns
}

// DropColumn marks a column to be dropped when the blueprint is executed.
func (t *Blueprint) DropColumn(column string) {
	t.dropColumns = append(t.dropColumns, column)
}

// GoString renders the blueprint as a Go function literal of the form
// "func(table *schema.Blueprint) { ... }".
func (b *Blueprint) GoString() string {
	src := strings.Builder{}
	src.WriteString("func(table *schema.Blueprint) {\n")
	for _, c := range b.columns {
		m := map[string]string{
			dialects.DataTypeBlob.Name:     "Blob",
			dialects.DataTypeBoolean.Name:  "Bool",
			dialects.DataTypeDate.Name:     "Date",
			dialects.DataTypeDateTime.Name: "DateTime",
			dialects.DataTypeFloat32.Name:  "Float",
			dialects.DataTypeFloat64.Name:  "Float64",
			dialects.DataTypeInt8.Name:     "Int8",
			dialects.DataTypeInt16.Name:    "Int16",
			dialects.DataTypeInt32.Name:    "Int",
			dialects.DataTypeInt64.Name:    "Int64",
			dialects.DataTypeJSON.Name:     "JSON",
			dialects.DataTypeString.Name:   "String",
			dialects.DataTypeUInt8.Name:    "UInt8",
			dialects.DataTypeUInt16.Name:   "UInt16",
			dialects.DataTypeUInt32.Name:   "UInt",
			dialects.DataTypeUInt64.Name:   "UInt64",
		}
		fmt.Fprintf(&src, "\ttable.%s(%#v)%s\n", m[c.datatype.Name], c.name, c.GoString())
	}

	for _, index := range b.indexes {
		fmt.Fprintf(&src, "\ttable.Index(%#v)%s\n", index.name, index.GoString())
	}

	for _, c := range b.dropColumns {
		fmt.Fprintf(&src, "\ttable.DropColumn(%#v)\n", c)
	}

	for _, foreignKey := range b.foreignKeys {
		fmt.Fprintf(&src, "\ttable.ForeignKey(%#v, %#v, %#v)\n", foreignKey.localKey, foreignKey.relatedTable, foreignKey.relatedKey)
	}

	if len(b.primaryKeys) > 1 {
		args := strings.Join(
			stream.Of(b.primaryKeys).Map(func(pKey string) string {
				return fmt.Sprintf("%#v", pKey)
			}).Slice(),
			", ",
		)
		fmt.Fprintf(&src, "\ttable.PrimaryKey(%s)\n", args)
	}

	src.WriteString("}")
	return src.String()
}

// Merge applies newBlueprint onto t. Columns marked with Change in
// newBlueprint replace existing columns by name, other columns are appended,
// and columns dropped in newBlueprint are removed. New foreign keys, indexes,
// and a primary key are taken from newBlueprint when present.
func (t *Blueprint) Merge(newBlueprint *Blueprint) {
	if t.name != newBlueprint.name {
		return
	}

	for _, newColumn := range newBlueprint.columns {
		if newColumn.change {
			for i, c := range t.columns {
				if c.name == newColumn.name {
					t.columns[i] = newColumn
					break
				}
			}
		} else {
			t.columns = append(t.columns, newColumn)
		}
	}

	t.columns = stream.Of(t.columns).Filter(func(c *ColumnBuilder) bool {
		return !slices.Contains(newBlueprint.dropColumns, c.name)
	}).Slice()

	t.foreignKeys = append(t.foreignKeys, newBlueprint.foreignKeys...)
	t.indexes = append(t.indexes, newBlueprint.indexes...)
	if newBlueprint.primaryKeys != nil {
		t.primaryKeys = newBlueprint.primaryKeys
	}
}

// Update rewrites t to describe the change from oldBlueprint to newBlueprint.
// Added columns are appended, modified columns are marked with Change, and
// columns that only exist in the old blueprint are dropped. New foreign keys
// and indexes are appended. It returns true if any change was detected.
//
// Primary key changes, and the removal of existing foreign keys or indexes,
// are not yet supported.
func (t *Blueprint) Update(oldBlueprint, newBlueprint *Blueprint) bool {
	addedColumns := sets.New[string]()
	hasChanges := false
	for _, newColumn := range newBlueprint.columns {
		oldColumn, ok := oldBlueprint.findColumn(newColumn.name)
		if ok {
			addedColumns.Add(newColumn.name)
			if newColumn.Equals(oldColumn) {
				continue
			}
		}

		newColumn.change = ok
		hasChanges = true
		t.AddColumn(newColumn)
	}
	for _, oldColumn := range oldBlueprint.columns {
		if !addedColumns.Has(oldColumn.name) {
			hasChanges = true
			t.DropColumn(oldColumn.name)
		}
	}

	for _, newKey := range newBlueprint.foreignKeys {
		_, ok := stream.Of(oldBlueprint.foreignKeys).Find(func(oldKey *ForeignKeyBuilder) bool {
			return newKey.localKey == oldKey.localKey &&
				newKey.relatedKey == oldKey.relatedKey &&
				newKey.relatedTable == oldKey.relatedTable
		})
		if !ok {
			t.foreignKeys = append(t.foreignKeys, newKey)
			hasChanges = true
		}
	}
	for _, newIndex := range newBlueprint.indexes {
		_, ok := stream.Of(oldBlueprint.indexes).Find(func(oldIndex *IndexBuilder) bool {
			return newIndex.name == oldIndex.name
		})
		if !ok {
			t.indexes = append(t.indexes, newIndex)
			hasChanges = true
		}
	}

	// TODO: add support for primary keys
	return hasChanges
}
