package channelpubsub

import (
	"context"

	"github.com/google/uuid"
	"gosalusa.com/pubsub"
)

type Topic struct {
	ch chan pubsub.Message
}

// Close implements [pubsub.Topic].
func (t *Topic) Close() error {
	return nil
}

// Enqueue implements [pubsub.Topic].
func (t *Topic) Enqueue(ctx context.Context, data []byte) error {
	t.ch <- &Message{
		id:    uuid.NewString(),
		data:  data,
		topic: t,
	}
	return nil
}

// Dequeue implements [pubsub.Topic].
func (t *Topic) Dequeue(ctx context.Context) (pubsub.Message, error) {
	select {
	case m := <-t.ch:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
