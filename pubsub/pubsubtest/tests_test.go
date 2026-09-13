package pubsubtest

import (
	"context"
	"testing"

	"gosalusa.com/pubsub"
)

type fakeMessage struct {
	data []byte
	ch   chan pubsub.Message
}

var _ pubsub.Message = (*fakeMessage)(nil)

func (m *fakeMessage) ID() string {
	return ""
}

func (m *fakeMessage) Data() []byte {
	return m.data
}

func (m *fakeMessage) Ack(ctx context.Context) error {
	return nil
}

func (m *fakeMessage) Nack(ctx context.Context) error {
	m.ch <- m
	return nil
}

type fakeTopic struct {
	ch chan pubsub.Message
}

var _ pubsub.Topic = (*fakeTopic)(nil)

func newFakeTopic() *fakeTopic {
	return &fakeTopic{ch: make(chan pubsub.Message, 10)}
}

func (t *fakeTopic) Enqueue(ctx context.Context, data []byte) error {
	t.ch <- &fakeMessage{data: data, ch: t.ch}
	return nil
}

func (t *fakeTopic) Dequeue(ctx context.Context) (pubsub.Message, error) {
	select {
	case m := <-t.ch:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (t *fakeTopic) Close() error {
	return nil
}

type fakePubSub struct {
	topic *fakeTopic
}

var _ pubsub.PubSub = (*fakePubSub)(nil)

func newFakePubSub() *fakePubSub {
	return &fakePubSub{topic: newFakeTopic()}
}

func (p *fakePubSub) Topic(name string) pubsub.Topic {
	return p.topic
}

func TestRunStandardTests(t *testing.T) {
	RunStandardTests(t, func(t *testing.T, name string, fn func(*testing.T, pubsub.PubSub)) {
		t.Run(name, func(t *testing.T) {
			fn(t, newFakePubSub())
		})
	})
}
