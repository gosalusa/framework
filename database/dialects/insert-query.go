package dialects

// InsertQueryBuilder is implemented by anything that produces an InsertQuery.
type InsertQueryBuilder interface {
	InsertQuery() *InsertQuery
}

// InsertQuery describes an INSERT statement. Values is a list of rows, each a
// map of column names to values, and Returning is the list of columns to
// return if the dialect supports RETURNING.
type InsertQuery struct {
	Table     string
	Values    []map[string]any
	Returning []string
}
