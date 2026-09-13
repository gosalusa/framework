package pubsub

import (
	"context"

	"gosalusa.com/di"
)

type PubSub interface {
	Topic(name string) Topic
}

type Topic interface {
	// Enqueue adds a message to the topic.
	Enqueue(ctx context.Context, data []byte) error

	// Dequeue fetches a single message. It should block until a message is
	// available or the context is canceled.
	Dequeue(ctx context.Context) (Message, error)

	// Close cleans up connections or background goroutines.
	Close() error
}

type Message interface {
	ID() string
	Data() []byte

	// Ack signals the message was processed successfully and can be deleted.
	Ack(ctx context.Context) error

	// Nack signals processing failed. The message should be re-queued or
	// dead-lettered.
	Nack(ctx context.Context) error
}

func RegisterTopic(ctx context.Context) {
	di.RegisterWith(ctx, func(ctx context.Context, tag string, with PubSub) (Topic, error) {
		return with.Topic(tag), nil
	})
}
