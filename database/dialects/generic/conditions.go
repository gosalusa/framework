package generic

import (
	"fmt"

	"gosalusa.com/database/dialects"
)

// EncodeWheres renders the WHERE clause of a query, or nothing if there are no
// conditions.
func (g *Generic) EncodeWheres(c []dialects.Condition) (dialects.RawQuery, error) {
	return g.encodeConditionsPrefix("WHERE", c)
}

// EncodeHavings renders the HAVING clause of a query, or nothing if there are
// no conditions.
func (g *Generic) EncodeHavings(c []dialects.Condition) (dialects.RawQuery, error) {
	return g.encodeConditionsPrefix("HAVING", c)
}

func (g *Generic) encodeConditionsPrefix(prefix string, c []dialects.Condition) (dialects.RawQuery, error) {
	if len(c) == 0 {
		return dialects.RawQuery{}, nil
	}
	return newRawQueryBuilder().AddString(prefix).Add(g.EncodeConditions(c)).Build()
}

// EncodeConditions renders the conditions joined by AND, or OR for a condition
// with Or set. A nil value with an = or != operator becomes IS NULL or IS NOT
// NULL, and a []any value becomes an IN (...) list.
func (g *Generic) EncodeConditions(c []dialects.Condition) (dialects.RawQuery, error) {
	b := newRawQueryBuilder()
	for i, c := range c {
		if i != 0 {
			if c.Or {
				b.AddString("OR")
			} else {
				b.AddString("AND")
			}
		}
		if (c.Column != dialects.Column{}) {
			b.Add(g.EncodeColumn(&c.Column))

			if c.Operator == "" {
				return dialects.RawQuery{}, fmt.Errorf("the operator must be set when the column is set")
			}
		}

		if c.Value == nil {
			switch c.Operator {
			case "=":
				b.AddString("IS NULL")
			case "!=":
				b.AddString("IS NOT NULL")
			default:
				return dialects.RawQuery{}, fmt.Errorf("wheres checking nil only support = and !=")
			}
		} else {
			if c.Operator != "" {
				b.AddString(c.Operator)
			}
			if inList, ok := c.Value.([]any); ok {
				b.Add(group(mapJoinRawQueries(inList, ", ", g.EncodeAny)))
			} else {
				b.Add(g.EncodeAny(c.Value))
			}
		}
	}

	return b.Build()
}
