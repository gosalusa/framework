package generic

import (
	"strings"

	"gosalusa.com/database/dialects"
)

// EncodeOrderBy renders the ORDER BY clause of a query, or nothing if there
// are no order columns.
func (g *Generic) EncodeOrderBy(orderBys []dialects.OrderColumn) (dialects.RawQuery, error) {
	if len(orderBys) == 0 {
		return dialects.RawQuery{}, nil
	}

	identifiers := make([]string, len(orderBys))
	for i, group := range orderBys {
		identifiers[i] = g.core.Identifier(group.Column)
		if group.Descending {
			identifiers[i] += " DESC"
		}
	}
	return dialects.Raw("ORDER BY " + strings.Join(identifiers, ", ")), nil
}
