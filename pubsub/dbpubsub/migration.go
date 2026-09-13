package dbpubsub

import (
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/schema"
)

var Migrations = []*migrate.Migration{
	{
		Name: "20260811_065216-Event",
		Up: schema.Create("events", func(table *schema.Blueprint) {
			table.Int("id").Primary().AutoIncrement()
			table.DateTime("created_at")
			table.DateTime("updated_at")

			table.DateTime("run_at")
			table.String("status")
			table.String("topic")
			table.Blob("data")
			table.Int("retries")

			table.Index("idx_job_queue_status_id").
				AddColumn("status").
				AddColumn("topic").
				AddColumn("run_at").
				AddColumn("id")
		}),
		Down: schema.DropIfExists("events"),
	},
}
