package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/event"
)

func TestLogEventType(t *testing.T) {
	e := &LogEvent{Message: "hello"}
	assert.Equal(t, event.EventType("template:example-event"), e.Type())
}
