package generic

import (
	"fmt"
	"strings"

	"gosalusa.com/database/dialects"
	"gosalusa.com/stream"
)

// EncodeCreateTableQuery renders q as a CREATE TABLE statement followed by any
// index statements.
func (g *Generic) EncodeCreateTableQuery(q *dialects.CreateTableQuery) (dialects.RawQuery, error) {
	b := newRawQueryBuilder().AddString("CREATE")
	if q.Temporary {
		b.AddString("TEMPORARY")
	}
	b.AddString("TABLE")
	if q.IfNotExists {
		b.AddString("IF NOT EXISTS")
	}
	b.AddString(g.core.Identifier(q.Table))

	columnsAndConstraints := make([]dialects.RawQuery, 0, len(q.Columns)+len(q.ForeignKeys)+1)
	for _, c := range q.Columns {
		c, err := g.EncodeColumnDefinition(&c)
		if err != nil {
			return dialects.RawQuery{}, err
		}
		columnsAndConstraints = append(columnsAndConstraints, c)
	}
	if len(q.PrimaryKeys) > 0 {

		columnsAndConstraints = append(columnsAndConstraints, dialects.Raw(fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(stream.Of(q.PrimaryKeys).Map(g.core.Identifier).Slice(), ", ")),
		))
	}
	for _, f := range q.ForeignKeys {
		f, err := g.EncodeForeignKey(&f)
		if err != nil {
			return dialects.RawQuery{}, err
		}
		columnsAndConstraints = append(columnsAndConstraints, f)
	}

	b.Add(group(joinRawQueries(columnsAndConstraints, ", "), nil))
	b.AddStringNoSpace(";")

	for _, index := range q.Indexes {
		b.Add(g.EncodeIndex(&index))
		b.AddStringNoSpace(";")
	}

	return b.Build()
}

// EncodeColumnDefinition renders a single column definition with its data type
// and constraints.
func (g *Generic) EncodeColumnDefinition(c *dialects.ColumnDefinition) (dialects.RawQuery, error) {
	r := newRawQueryBuilder()
	r.AddString(g.core.Identifier(c.Name))
	r.AddString(g.core.DataType(c.Datatype))

	if c.AutoIncrement {
		r.AddString(g.core.AutoIncrement())
	} else if c.Primary {
		r.AddString("PRIMARY KEY")
	}
	if !c.Nullable {
		r.AddString("NOT NULL")
	}
	if c.Unique {
		r.AddString("UNIQUE")
	}

	if c.DefaultValue != nil {
		r.AddString("DEFAULT").
			AddString(g.core.Escape(c.DefaultValue))
	} else if c.DefaultCurrentTime {
		r.AddString("DEFAULT").
			AddString(g.core.CurrentTime())
	}
	return r.Build()
}

// EncodeForeignKey renders a CONSTRAINT ... FOREIGN KEY clause.
func (g *Generic) EncodeForeignKey(f *dialects.ForeignKey) (dialects.RawQuery, error) {
	return dialects.Raw(fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		g.core.Identifier(f.Name),
		strings.Join(stream.Of(f.Columns).Map(g.core.Identifier).Slice(), ", "),
		g.core.Identifier(f.ForeignTable),
		strings.Join(stream.Of(f.ForeignColumns).Map(g.core.Identifier).Slice(), ", "),
	)), nil
}
