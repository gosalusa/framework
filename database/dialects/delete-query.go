package dialects

// DeleteQueryBuilder is implemented by anything that produces a DeleteQuery.
type DeleteQueryBuilder interface {
	DeleteQuery() *DeleteQuery
}

// DeleteQuery describes a DELETE statement. Wheres selects the rows to delete.
type DeleteQuery struct {
	Table  string
	Wheres []Condition
}
