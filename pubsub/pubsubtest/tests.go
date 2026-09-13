package pubsubtest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/pubsub"
)

func RunStandardTests(t *testing.T, run func(t *testing.T, name string, fn func(*testing.T, pubsub.PubSub))) {
	t.Helper()
	run(t, "push pop", func(t *testing.T, p pubsub.PubSub) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		q := p.Topic("default")
		defer q.Close()

		q.Enqueue(ctx, []byte("a"))

		m, err := q.Dequeue(ctx)
		if assert.NoError(t, err) {
			assert.Equal(t, []byte("a"), m.Data())
		}
	})

	run(t, "multi push", func(t *testing.T, p pubsub.PubSub) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		q := p.Topic("default")
		defer q.Close()

		q.Enqueue(ctx, []byte("a"))
		q.Enqueue(ctx, []byte("b"))

		a, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []byte("a"), a.Data())

		b, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []byte("b"), b.Data())
	})

	run(t, "ack", func(t *testing.T, p pubsub.PubSub) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		q := p.Topic("default")
		defer q.Close()

		q.Enqueue(ctx, []byte("a"))

		m, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, []byte("a"), m.Data())
		err = m.Ack(ctx)
		if !assert.NoError(t, err) {
			return
		}
	})

	run(t, "nack", func(t *testing.T, p pubsub.PubSub) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		q := p.Topic("default")
		defer q.Close()

		q.Enqueue(ctx, []byte("a"))

		m1, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, []byte("a"), m1.Data())

		err = m1.Nack(ctx)
		if !assert.NoError(t, err) {
			return
		}

		m2, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, []byte("a"), m2.Data())
	})
}
