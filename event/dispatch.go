package event

import (
	"context"

	"gosalusa.com/di"
	"gosalusa.com/pubsub"
)

type Dispatch func(ctx context.Context, e Event) error

func NewDispatch(t pubsub.Topic) Dispatch {
	return func(ctx context.Context, e Event) error {
		b, err := encodeEvent(e)
		if err != nil {
			return err
		}
		return t.Enqueue(ctx, b)
	}
}

func Register(ctx context.Context) {
	di.RegisterWith[Dispatch, pubsub.PubSub](ctx, func(ctx context.Context, tag string, with pubsub.PubSub) (Dispatch, error) {
		return NewDispatch(with.Topic(tag)), nil
	})
}
