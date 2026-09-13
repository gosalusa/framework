package generic

import (
	"sort"

	"gosalusa.com/database/dialects"
)

// EncodeUpdateQuery renders q as an UPDATE statement, including any WHERE
// clause and a RETURNING clause when Returning is set.
func (g *Generic) EncodeUpdateQuery(q *dialects.UpdateQuery) (dialects.RawQuery, error) {
	b := newRawQueryBuilder().
		AddString("UPDATE").
		AddString(g.core.Identifier(q.Table)).
		Add(g.EncodeUpdateSet(q.Values)).
		Add(g.EncodeWheres(q.Wheres))

	if len(q.Returning) > 0 {
		b.AddString("RETURNING")
		b.Add(g.EncodeSelects(&dialects.Select{
			Columns: q.Returning,
		}))
	}
	return b.Build()
}

// EncodeUpdateSet renders the SET clause of an update with the columns sorted
// for deterministic output.
func (g *Generic) EncodeUpdateSet(values map[string]any) (dialects.RawQuery, error) {
	b := newRawQueryBuilder().AddString("SET")
	results := make([]dialects.RawQuery, 0, len(values))
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		r, err := newRawQueryBuilder().
			AddString(g.core.Identifier(k)).
			AddString("=").
			Add(g.EncodeAny(values[k])).
			Build()
		if err != nil {
			return dialects.RawQuery{}, err
		}
		results = append(results, r)
	}
	return b.Add(joinRawQueries(results, ", "), nil).Build()
}
