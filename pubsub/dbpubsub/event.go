package dbpubsub

import (
	"time"

	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/database/model/mixins"
)

//go:generate spice generate:migration
type Event struct {
	model.BaseModel
	mixins.Timestamps

	ID      int         `db:"id,primary,autoincrement"`
	RunAt   time.Time   `db:"run_at"`
	Status  EventStatus `db:"status"`
	Topic   string      `db:"topic"`
	Data    []byte      `db:"data"`
	Retries int         `db:"retries"`
}

type EventStatus string

var (
	EventPending    = EventStatus("pending")
	EventProcessing = EventStatus("processing")
	EventFinished   = EventStatus("finished")
	EventError      = EventStatus("error")
)

func EventQuery() *builder.ModelBuilder[*Event] {
	return builder.From[*Event]()
}
