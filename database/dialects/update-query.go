package dialects

// UpdateQueryBuilder is implemented by anything that produces an UpdateQuery.
type UpdateQueryBuilder interface {
	UpdateQuery() *UpdateQuery
}

// UpdateQuery describes an UPDATE statement. Values maps column names to their
// new values, Returning is the list of columns to return, and Wheres selects
// the rows to update.
type UpdateQuery struct {
	Table     string
	Values    map[string]any
	Returning []Column
	Wheres    []Condition
}
