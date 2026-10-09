package channelpubsub

import (
	"sync"

	"gosalusa.com/pubsub"
)

type PubSub struct {
	topics map[string]chan pubsub.Message
	mtx    sync.Mutex
}

var _ pubsub.PubSub = (*PubSub)(nil)

func New() *PubSub {
	return &PubSub{
		topics: map[string]chan pubsub.Message{},
	}
}

// Topic implements [pubsub.PubSub].
func (p *PubSub) Topic(name string) pubsub.Topic {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	t, ok := p.topics[name]

	if !ok {
		t = make(chan pubsub.Message, 10)
		p.topics[name] = t
	}

	return &Topic{
		ch: t,
	}
}
