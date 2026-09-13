package migrations

import (
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/schema"
)

func init() {
	migrations.Add(&migrate.Migration{
		Name: "20240322_143359-Foo",
		Up: schema.Create("foos", func(table *schema.Blueprint) {
			table.Int("id").Primary().AutoIncrement()
		}),
		Down: schema.DropIfExists("foos"),
	})
}
