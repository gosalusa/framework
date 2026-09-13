package dialects

// AlterTableQueryBuilder is implemented by anything that produces an
// AlterTableQuery.
type AlterTableQueryBuilder interface {
	AlterTableQuery() *AlterTableQuery
}

// AlterTableQuery describes changes to a table: columns to drop, modify, or
// add, plus foreign keys and indexes to add.
type AlterTableQuery struct {
	Table         string
	DropColumns   []string
	ModifyColumns []ColumnDefinition
	AddColumns    []ColumnDefinition
	ForeignKeys   []ForeignKey
	Indexes       []Index
}
