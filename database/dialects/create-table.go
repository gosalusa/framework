package dialects

// CreateTableQueryBuilder is implemented by anything that produces a
// CreateTableQuery.
type CreateTableQueryBuilder interface {
	CreateTableQuery() *CreateTableQuery
}

// CreateTableQuery describes a CREATE TABLE statement, its columns, composite
// primary key and foreign keys, and any indexes to create alongside it.
type CreateTableQuery struct {
	IfNotExists bool
	Temporary   bool
	Table       string
	Columns     []ColumnDefinition
	PrimaryKeys []string
	ForeignKeys []ForeignKey
	Indexes     []Index
}

// ColumnDefinition describes a single column of a table: its name, data type,
// and constraints such as nullability, primary key, default value, and
// auto-increment. DefaultCurrentTime makes the default the current timestamp.
type ColumnDefinition struct {
	Name               string
	Datatype           DataType
	Nullable           bool
	Primary            bool
	AutoIncrement      bool
	DefaultValue       any
	Unique             bool
	DefaultCurrentTime bool
}

// ForeignKey describes a FOREIGN KEY constraint. Columns are the local
// columns, and ForeignTable with ForeignColumns is the referenced table.
type ForeignKey struct {
	Name           string
	Columns        []string
	ForeignTable   string
	ForeignColumns []string
	// OnUpdate       string
	// OnDelete       string
}
