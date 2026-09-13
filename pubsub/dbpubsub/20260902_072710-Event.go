package dbpubsub

import (
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/schema"
)

func init() {
	migrations.Add(&migrate.Migration{
		Name: "20260902_072710-Event",
		Up: schema.Create("events", func(table *schema.Blueprint) {
			table.DateTime("created_at")
			table.DateTime("updated_at")
			table.Int("id").Primary().AutoIncrement()
			table.DateTime("run_at")
			table.String("status")
			table.String("topic")
			table.Blob("data")
			table.Int("retries")
		}),
		Down: schema.DropIfExists("events"),
	})
}
