package generic

import (
	"errors"
	"sort"
	"strings"

	"gosalusa.com/database/dialects"
	"gosalusa.com/stream"
)

// ErrInsertNoRows is returned by EncodeInsertQuery when there are no rows to
// insert.
var ErrInsertNoRows = errors.New("no rows to insert")

// ErrInsertMismatchedValueKeys is returned by EncodeInsertQuery when the rows
// to insert have different sets of column names.
var ErrInsertMismatchedValueKeys = errors.New("mismatched value keys")

// EncodeInsertQuery renders q as an INSERT statement. Column names are sorted
// for deterministic output, and RETURNING is added when Returning is set.
func (g *Generic) EncodeInsertQuery(q *dialects.InsertQuery) (dialects.RawQuery, error) {
	if len(q.Values) == 0 {
		return dialects.RawQuery{}, ErrInsertNoRows
	}
	numColumns := len(q.Values[0])

	columns := make([]string, 0, numColumns)
	values := make([][]any, len(q.Values))
	for k := range q.Values[0] {
		columns = append(columns, k)
	}
	sort.Strings(columns)

	for i, newRow := range q.Values {
		if len(q.Values[i]) != numColumns {
			return dialects.RawQuery{}, ErrInsertMismatchedValueKeys
		}
		values[i] = make([]any, numColumns)
		for j, column := range columns {
			values[i][j] = newRow[column]
		}
	}

	b := newRawQueryBuilder().
		AddString("INSERT INTO").
		AddString(g.core.Identifier(q.Table)).
		AddString("(" + strings.Join(stream.Of(columns).Map(g.core.Identifier).Slice(), ", ") + ")").
		AddString("VALUES")

	for i, v := range values {
		if i > 0 {
			b.AddStringNoSpace(",")
		}
		b.Add(group(mapJoinRawQueries(v, ", ", g.EncodeLiteral)))
	}

	returning := make([]string, len(q.Returning))
	for i, r := range q.Returning {
		returning[i] = g.core.Identifier(r)
	}
	if len(returning) > 0 {
		b.AddString("RETURNING " + strings.Join(returning, ", "))
	}

	return b.Build()
}
