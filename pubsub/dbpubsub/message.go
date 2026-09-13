package dbpubsub

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"gosalusa.com/database"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/pubsub"
)

var ErrMessageFinished = errors.New("message finished")

type Message struct {
	event    *Event
	update   database.Update
	finished bool
}

var _ pubsub.Message = (*Message)(nil)

// Data implements [event.Message].
func (m *Message) Data() []byte {
	return m.event.Data
}

// ID implements [event.Message].
func (m *Message) ID() string {
	return strconv.FormatInt(int64(m.event.ID), 10)
}

// Ack implements [event.Message].
func (m *Message) Ack(ctx context.Context) error {
	if m.finished {
		return ErrMessageFinished
	}
	m.finished = true
	return m.update(func(tx *sqlx.Tx) error {
		m.event.Status = EventFinished
		return model.SaveContext(ctx, tx, m.event)
	})
}

// Nack implements [event.Message].
func (m *Message) Nack(ctx context.Context) error {
	if m.finished {
		return ErrMessageFinished
	}
	m.finished = true
	return m.update(func(tx *sqlx.Tx) error {
		m.event.Status = EventPending
		m.event.RunAt = time.Now().Add(helpers.ExponentialFalloff(m.event.Retries))
		m.event.Retries++
		return model.SaveContext(ctx, tx, m.event)
	})
}
