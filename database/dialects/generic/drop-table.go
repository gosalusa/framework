package generic

import "gosalusa.com/database/dialects"

// return runQuery(ctx, tx, helpers.Concat(helpers.Raw("DROP TABLE IF EXISTS "), helpers.Identifier(table)))
// EncodeDropTableQuery renders q as a DROP TABLE statement, including IF
// EXISTS when set.
func (g *Generic) EncodeDropTableQuery(q *dialects.DropTableQuery) (dialects.RawQuery, error) {
	b := newRawQueryBuilder().AddString("DROP TABLE")
	if q.IfExists {
		b.AddString("IF EXISTS")
	}
	return b.AddString(g.core.Identifier(q.Table)).Build()
}
