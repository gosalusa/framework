package schema

import (
	"fmt"

	"gosalusa.com/database/dialects"
)

// IndexBuilder describes an index on a table. Create one with
// [Blueprint.Index] and chain AddColumn, and optionally Unique, on it.
type IndexBuilder struct {
	table   string
	name    string
	columns []string
	unique  bool
}

func newIndexBuilder(table string) *IndexBuilder {
	return &IndexBuilder{
		columns: []string{},
		table:   table,
	}
}

// AddColumn adds a column to the index.
func (b *IndexBuilder) AddColumn(c string) *IndexBuilder {
	b.columns = append(b.columns, c)
	return b
}

// Unique makes the index enforce uniqueness.
func (b *IndexBuilder) Unique() *IndexBuilder {
	b.unique = true
	return b
}

// Index converts the builder into its dialects.Index representation that a
// dialect encodes into SQL.
func (b *IndexBuilder) Index() *dialects.Index {
	return &dialects.Index{
		Table:   b.table,
		Name:    b.name,
		Columns: b.columns,
		Unique:  b.unique,
	}
}

// GoString renders the index's modifiers as a chain of Go method calls.
func (b *IndexBuilder) GoString() string {
	src := ""
	for _, c := range b.columns {
		src += fmt.Sprintf(".AddColumn(%#v)", c)
	}
	if b.unique {
		src += ".Unique()"
	}
	return src
}
