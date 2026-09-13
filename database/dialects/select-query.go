package dialects

// QueryBuilder is implemented by anything that produces a SelectQuery, such as
// the query builders used as subqueries.
type QueryBuilder interface {
	Query() *SelectQuery
}

// SelectQuery describes a SELECT statement with its columns, source table,
// joins, conditions, grouping, ordering, limit, and locking clause.
type SelectQuery struct {
	Select    Select
	From      string
	Joins     []Join
	Wheres    []Condition
	Havings   []Condition
	GroupBys  []string
	OrderBys  []OrderColumn
	Limit     Limit
	ForUpdate ForUpdate
}

// NewSelectQuery returns a SelectQuery with its slices initialized.
func NewSelectQuery() SelectQuery {
	return SelectQuery{
		Select:   NewSelect(),
		Joins:    []Join{},
		Wheres:   []Condition{},
		GroupBys: []string{},
		Havings:  []Condition{},
		OrderBys: []OrderColumn{},
	}
}

// OrderColumn is a single column in an ORDER BY clause. Descending makes the
// sort descending.
type OrderColumn struct {
	Column     string
	Descending bool
}

// Select is the SELECT clause of a query: an optional DISTINCT followed by the
// columns to select.
type Select struct {
	Distinct bool
	Columns  []Column
}

// NewSelect returns a Select with its column slice initialized.
func NewSelect() Select {
	return Select{
		Columns: []Column{},
	}
}

// Join is a single join clause. Direction is a join keyword such as "LEFT" or
// "INNER", Table is the table being joined, and Conditions are the ON
// conditions.
type Join struct {
	Direction  string
	Table      string
	Conditions []Condition
}

// Condition is a single predicate on a column or expression. Operator is a
// comparison such as "=" or "!=" and Value is the value compared against. Or
// joins the condition to the previous one with OR instead of AND.
type Condition struct {
	Column   Column
	Operator string
	Value    any
	Or       bool
}

// Column is an expression in a query: a column name, a function call, a
// subquery, or raw SQL. As is an optional alias.
type Column struct {
	Column   string
	Function *FunctionCall
	SubQuery QueryBuilder
	Raw      string

	As string
}

// FunctionCall is a SQL function applied to Arguments, for example "count" with
// arguments "*".
type FunctionCall struct {
	Name      string
	Arguments string
}

// Limit is the LIMIT and OFFSET of a query.
type Limit struct {
	Limit  int
	Offset int
}

// ForUpdate is a row locking clause for SELECT statements.
type ForUpdate string

const (
	// ForUpdateDefault locks the selected rows, waiting for other transactions
	// that hold the locks to commit.
	ForUpdateDefault = ForUpdate("default")
	// ForUpdateSkipLocked locks the selected rows but skips any rows already
	// locked by other transactions.
	ForUpdateSkipLocked = ForUpdate("skip-locked")
)

// RawString is a raw SQL fragment that can be used as a value in conditions.
type RawString string

// Raw returns a RawQuery from a SQL string and its optional bind values.
func Raw(sql string, bindings ...any) RawQuery {
	return RawQuery{
		SQL:      sql,
		Bindings: bindings,
	}
}
