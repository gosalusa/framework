package generic

import "gosalusa.com/database/dialects"

// EncodeDeleteQuery renders q as a DELETE statement, including any WHERE
// clause.
func (g *Generic) EncodeDeleteQuery(q *dialects.DeleteQuery) (dialects.RawQuery, error) {
	return newRawQueryBuilder().
		AddString("DELETE FROM").
		AddString(g.core.Identifier(q.Table)).
		Add(g.EncodeWheres(q.Wheres)).
		Build()
}
