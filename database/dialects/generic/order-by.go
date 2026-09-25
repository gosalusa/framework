package generic

import (
	"gosalusa.com/database/dialects"
)

// EncodeOrderBy renders the ORDER BY clause of a query, or nothing if there
// are no order columns.
func (g *Generic) EncodeOrderBy(orderBys []dialects.OrderColumn) (dialects.RawQuery, error) {
	if len(orderBys) == 0 {
		return dialects.RawQuery{}, nil
	}

	b := newRawQueryBuilder()
	b.AddString("ORDER BY")
	for i, group := range orderBys {
		if i > 0 {
			b.AddStringNoSpace(",")
		}
		b.Add(g.EncodeColumn(&group.Column))
		if group.Descending {
			b.AddString("DESC")
		}
	}
	return b.Build()
}
