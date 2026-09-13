package dialects

// DropTableQueryBuilder is implemented by anything that produces a
// DropTableQuery.
type DropTableQueryBuilder interface {
	DropTableQuery() *DropTableQuery
}

// DropTableQuery describes a DROP TABLE statement. IfExists drops the table
// without error if it does not exist.
type DropTableQuery struct {
	Table    string
	IfExists bool
}
