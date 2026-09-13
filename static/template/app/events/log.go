package events

import (
	"gosalusa.com/event"
	"gosalusa.com/event/cron"
)

type LogEvent struct {
	cron.CronEvent
	Message string
}

var _ event.Event = (*LogEvent)(nil)

func (e *LogEvent) Type() event.EventType {
	return "template:example-event"
}
