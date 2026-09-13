package generic

import (
	"strings"

	"gosalusa.com/database/dialects"
)

// EncodeIndex renders i as a CREATE [UNIQUE] INDEX statement.
func (g *Generic) EncodeIndex(i *dialects.Index) (dialects.RawQuery, error) {

	r := newRawQueryBuilder().AddString("CREATE")
	if i.Unique {
		r.AddString("UNIQUE")
	}

	columns := make([]string, len(i.Columns))

	for i, c := range i.Columns {
		columns[i] = g.core.Identifier(c)
	}

	r.AddString("INDEX IF NOT EXISTS").
		AddString(g.core.Identifier(i.Name)).
		AddString("ON").
		AddString(g.core.Identifier(i.Table)).
		AddString("(" + strings.Join(columns, ", ") + ")")

	return r.Build()
}
