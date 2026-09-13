package dialects

import "strings"

// RawQuery is a SQL statement together with the positional bind values it
// needs.
type RawQuery struct {
	SQL      string
	Bindings []any
}

// JoinQueries combines several RawQueries into a single multi-statement query,
// separating each statement with a space and semicolon and concatenating the
// bindings in order.
func JoinQueries(results []RawQuery) RawQuery {
	if len(results) == 0 {
		return RawQuery{
			SQL:      "",
			Bindings: []any{},
		}
	}
	if len(results) == 1 {
		return RawQuery{
			SQL:      results[0].SQL + ";",
			Bindings: results[0].Bindings,
		}
	}
	queryLen := 2*len(results) - 1
	bindingCount := 0

	for _, r := range results {
		queryLen += len(r.SQL)
		bindingCount += len(r.Bindings)
	}

	query := strings.Builder{}
	query.Grow(queryLen)
	bindings := make([]any, 0, bindingCount)

	for i, r := range results {
		if i > 0 {
			query.WriteString(" ")
		}
		query.WriteString(r.SQL)
		query.WriteString(";")
		bindings = append(bindings, r.Bindings...)
	}

	return RawQuery{
		SQL:      query.String(),
		Bindings: bindings,
	}
}
