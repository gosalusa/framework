package schema

import "gosalusa.com/database/dialects"

// ForeignKeyBuilder describes a foreign-key constraint created with
// [Blueprint.ForeignKey], referencing a column in another table.
type ForeignKeyBuilder struct {
	relatedTable string
	localKey     string
	relatedKey   string
}

// ForeignKey converts the builder into its dialects.ForeignKey representation.
// The constraint has a column on localKey referencing relatedKey in the
// related table and is named "<localKey>-<relatedTable>-<relatedKey>".
func (b *ForeignKeyBuilder) ForeignKey() *dialects.ForeignKey {
	return &dialects.ForeignKey{
		Name:           b.localKey + "-" + b.relatedTable + "-" + b.relatedKey,
		Columns:        []string{b.localKey},
		ForeignTable:   b.relatedTable,
		ForeignColumns: []string{b.relatedKey},
	}
}
