package generic

import (
	"strings"

	"gosalusa.com/database/dialects"
)

// EncodeGroupBy renders the GROUP BY clause of a query, or nothing if there
// are no groups.
func (g *Generic) EncodeGroupBy(groups []string) (dialects.RawQuery, error) {
	if len(groups) == 0 {
		return dialects.RawQuery{}, nil
	}
	b := newRawQueryBuilder()
	b.AddString("GROUP BY")

	identifiers := make([]string, len(groups))
	for i, group := range groups {
		identifiers[i] = g.core.Identifier(group)
	}
	return b.AddString(strings.Join(identifiers, ", ")).Build()
}
