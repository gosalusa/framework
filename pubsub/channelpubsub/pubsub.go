package channelpubsub

import (
	"gosalusa.com/pubsub"
)

type PubSub struct {
	topics map[string]chan pubsub.Message
}

var _ pubsub.PubSub = (*PubSub)(nil)

func New() *PubSub {
	return &PubSub{
		topics: map[string]chan pubsub.Message{},
	}
}

// Topic implements [pubsub.PubSub].
func (p *PubSub) Topic(name string) pubsub.Topic {
	t, ok := p.topics[name]

	if !ok {
		t = make(chan pubsub.Message, 10)
		p.topics[name] = t
	}

	return &Topic{
		ch: t,
	}
}
