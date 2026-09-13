package generic

import (
	"gosalusa.com/database/dialects"
)

// EncodeAlterTableQuery renders q as one or more ALTER TABLE statements,
// one for each column, foreign key, or index change.
func (g *Generic) EncodeAlterTableQuery(q *dialects.AlterTableQuery) (dialects.RawQuery, error) {
	b := newRawQueryBuilder()

	alterTable := "ALTER TABLE " + g.core.Identifier(q.Table)
	for _, column := range q.DropColumns {
		b.AddStringf("%s DROP COLUMN %s;", alterTable, g.core.Identifier(column))
	}
	for _, column := range q.ModifyColumns {
		b.AddStringf("%s MODIFY COLUMN", alterTable)
		b.Add(g.EncodeColumnDefinition(&column))
		b.AddStringNoSpace(";")
	}
	for _, column := range q.AddColumns {
		b.AddStringf("%s ADD", alterTable)
		b.Add(g.EncodeColumnDefinition(&column))
		b.AddStringNoSpace(";")
	}
	for _, foreignKey := range q.ForeignKeys {
		b.AddString(alterTable)
		b.AddString("ADD")
		b.Add(g.EncodeForeignKey(&foreignKey))
		b.AddStringNoSpace(";")
	}
	for _, index := range q.Indexes {
		b.Add(g.EncodeIndex(&index))
		b.AddStringNoSpace(";")
	}

	return b.Build()
}
