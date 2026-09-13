package generic

import "gosalusa.com/database/dialects"

// EncodeFrom renders the FROM clause of a query, or an empty query if from is
// empty.
func (g *Generic) EncodeFrom(from string) (dialects.RawQuery, error) {
	if from == "" {
		return dialects.RawQuery{}, nil
	}
	return dialects.RawQuery{
		SQL: "FROM " + g.core.Identifier(from),
	}, nil
}
