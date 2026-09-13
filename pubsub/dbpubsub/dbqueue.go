package dbpubsub

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"gosalusa.com/database"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/pubsub"
)

type PubSub struct {
	table  string
	update database.Update
}

var _ pubsub.PubSub = (*PubSub)(nil)

func New(u database.Update) *PubSub {
	return &PubSub{
		update: u,
	}
}

// Topic implements [pubsub.PubSub].
func (p *PubSub) Topic(name string) pubsub.Topic {
	return &Topic{
		topic:  name,
		update: p.update,
	}
}

type Topic struct {
	topic  string
	update database.Update
}

var _ pubsub.Topic = (*Topic)(nil)

// Close implements [pubsub.Queue].
func (t *Topic) Close() error {
	return nil
}

// // Consume implements [pubsub.Queue].
// func (t *Topic) Consume(ctx context.Context) (<-chan pubsub.Message, error) {
// 	messages := make(chan pubsub.Message)
// 	go func() {
// 		t.listen(ctx, messages)
// 		close(messages)
// 	}()
// 	return messages, nil
// }

// Enqueue implements [pubsub.Queue].
func (t *Topic) Enqueue(ctx context.Context, data []byte) error {
	return t.update(func(tx *sqlx.Tx) error {
		return model.Save(tx, &Event{
			Data:   data,
			RunAt:  time.Now(),
			Topic:  t.topic,
			Status: EventPending,
		})
	})
}

// Dequeue implements [pubsub.Topic].
func (t *Topic) Dequeue(ctx context.Context) (pubsub.Message, error) {
	tick := time.Tick(time.Second)

	for {
		events, err := database.Value(t.update, func(tx *sqlx.Tx) ([]*Event, error) {
			now := time.Now()
			return EventQuery().
				Where("id", "=", EventQuery().
					Select("id").
					Where("status", "=", "pending").
					Where("topic", "=", t.topic).
					Where("run_at", "<", now).
					OrderBy("id").
					Limit(1)).
				ForUpdateSkipLocked().
				UpdateReturning(tx, builder.Updates{
					"status":     "processing",
					"updated_at": now,
				})
		})
		if err != nil {
			return nil, err
		}
		if len(events) > 0 {
			return &Message{
				event:  events[0],
				update: t.update,
			}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tick:
		}
	}
}
