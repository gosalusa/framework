package generic

import "gosalusa.com/database/dialects"

// EncodeLiteral renders v as a binding placeholder with the value as a
// binding.
func (g *Generic) EncodeLiteral(v any) (dialects.RawQuery, error) {
	return dialects.RawQuery{
		SQL:      g.core.Binding(),
		Bindings: []any{v},
	}, nil
}

// EncodeAny renders a value, dispatching on its type: a QueryBuilder becomes a
// subquery, a []Condition a grouped condition, a Column a column reference, a
// RawQuery or RawString raw SQL, and anything else a literal binding.
func (g *Generic) EncodeAny(v any) (dialects.RawQuery, error) {
	b := newRawQueryBuilder()
	switch v := v.(type) {
	case dialects.QueryBuilder:
		b.Add(group(g.EncodeSelectQuery(v.Query())))
	case []dialects.Condition:
		b.Add(group(g.EncodeConditions(v)))
	case dialects.Column:
		b.Add(g.EncodeColumn(&v))
	case dialects.RawQuery:
		b.Add(v, nil)
	case dialects.RawString:
		b.AddString(string(v))
	default:
		b.Add(g.EncodeLiteral(v))
	}
	return b.Build()
}
