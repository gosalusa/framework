package pubsub_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/pubsub"
)

type fakeTopic struct {
	name string
}

var _ pubsub.Topic = (*fakeTopic)(nil)

func (t *fakeTopic) Enqueue(ctx context.Context, data []byte) error {
	return nil
}

func (t *fakeTopic) Dequeue(ctx context.Context) (pubsub.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Second):
		return nil, nil
	}
}

func (t *fakeTopic) Close() error {
	return nil
}

type fakePubSub struct {
	mu     sync.Mutex
	topics map[string]*fakeTopic
}

var _ pubsub.PubSub = (*fakePubSub)(nil)

func newFakePubSub() *fakePubSub {
	return &fakePubSub{topics: map[string]*fakeTopic{}}
}

func (p *fakePubSub) Topic(name string) pubsub.Topic {
	p.mu.Lock()
	defer p.mu.Unlock()
	t, ok := p.topics[name]
	if !ok {
		t = &fakeTopic{name: name}
		p.topics[name] = t
	}
	return t
}

func (p *fakePubSub) topic(name string) *fakeTopic {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.topics[name]
}

func TestRegisterTopic(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	ps := newFakePubSub()
	di.RegisterSingleton(ctx, func() pubsub.PubSub { return ps })

	pubsub.RegisterTopic(ctx)

	type deps struct {
		Default pubsub.Topic `inject:"default"`
		Other   pubsub.Topic `inject:"other"`
	}

	d, err := di.Resolve[deps](ctx)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, d.Default)
	assert.NotNil(t, d.Other)
	assert.NotEqual(t, d.Default, d.Other)

	dt, ok := d.Default.(*fakeTopic)
	if assert.True(t, ok, "topic should be the fake topic") {
		assert.Equal(t, "default", dt.name)
		assert.Same(t, ps.topic("default"), dt)
	}

	ot, ok := d.Other.(*fakeTopic)
	if assert.True(t, ok, "topic should be the fake topic") {
		assert.Equal(t, "other", ot.name)
		assert.Same(t, ps.topic("other"), ot)
	}
}
