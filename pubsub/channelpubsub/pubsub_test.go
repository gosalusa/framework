package channelpubsub_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/pubsub"
	"gosalusa.com/pubsub/channelpubsub"
	"gosalusa.com/pubsub/pubsubtest"
)

func TestStandard(t *testing.T) {
	pubsubtest.RunStandardTests(t, func(t *testing.T, name string, fn func(*testing.T, pubsub.PubSub)) {
		t.Run(name, func(t *testing.T) {
			p := channelpubsub.New()
			fn(t, p)
		})
	})
}

func TestDequeueContextCanceled(t *testing.T) {
	p := channelpubsub.New()
	q := p.Topic("default")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m, err := q.Dequeue(ctx)
	assert.Nil(t, m)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMessageID(t *testing.T) {
	ctx := context.Background()
	p := channelpubsub.New()
	q := p.Topic("default")

	if !assert.NoError(t, q.Enqueue(ctx, []byte("a"))) {
		return
	}
	if !assert.NoError(t, q.Enqueue(ctx, []byte("b"))) {
		return
	}

	m1, err := q.Dequeue(ctx)
	if !assert.NoError(t, err) {
		return
	}
	m2, err := q.Dequeue(ctx)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, m1.ID())
	assert.NotEmpty(t, m2.ID())
	assert.NotEqual(t, m1.ID(), m2.ID())
}

type registerDeps struct {
	PS    pubsub.PubSub `inject:""`
	Topic pubsub.Topic  `inject:"default"`
}

func TestRegister(t *testing.T) {
	ctx := di.TestDependencyProviderContext()

	channelpubsub.Register(ctx)

	d, err := di.Resolve[registerDeps](ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, d.PS)

	bg := context.Background()
	if !assert.NoError(t, d.Topic.Enqueue(bg, []byte("hello"))) {
		return
	}

	m, err := d.Topic.Dequeue(bg)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []byte("hello"), m.Data())

	d2, err := di.Resolve[registerDeps](ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Same(t, d.PS, d2.PS)

	q2 := d2.PS.Topic("default")
	assert.NoError(t, q2.Enqueue(bg, []byte("world")))
	m2, err := d.Topic.Dequeue(bg)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []byte("world"), m2.Data())
}
