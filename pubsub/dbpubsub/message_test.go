package dbpubsub_test

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database"
	"gosalusa.com/internal/test"
	"gosalusa.com/pubsub"
	"gosalusa.com/pubsub/dbpubsub"
)

var errBoom = errors.New("boom")

type messageTestEnv struct {
	p       pubsub.PubSub
	failing *atomic.Bool
}

func runMessageTest(t *testing.T, cb func(t *testing.T, ctx context.Context, env messageTestEnv)) {
	t.Helper()
	test.Run(t, "messages", func(t *testing.T, tx *sqlx.Tx) {
		if tx.DriverName() == "mysql" {
			t.SkipNow()
		}
		for _, m := range dbpubsub.Migrations {
			err := m.Up.Run(t.Context(), tx)
			if !assert.NoError(t, err) {
				return
			}
		}

		defer func() {
			v := recover()
			m := slices.Clone(dbpubsub.Migrations)
			slices.Reverse(m)
			for _, m := range m {
				err := m.Down.Run(t.Context(), tx)
				if !assert.NoError(t, err) {
					return
				}
			}

			if v != nil {
				panic(v)
			}
		}()

		var failing atomic.Bool
		u := database.Update(func(f func(tx *sqlx.Tx) error) error {
			if failing.Load() {
				return errBoom
			}
			return f(tx)
		})

		cb(t, t.Context(), messageTestEnv{
			p:       dbpubsub.New(u),
			failing: &failing,
		})
	})
}

func TestMessageID(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q1 := env.p.Topic("one")
		q2 := env.p.Topic("two")

		if !assert.NoError(t, q1.Enqueue(ctx, []byte("a"))) {
			return
		}
		if !assert.NoError(t, q2.Enqueue(ctx, []byte("b"))) {
			return
		}

		m1, err := q1.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}
		m2, err := q2.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		assert.NotEmpty(t, m1.ID())
		assert.NotEmpty(t, m2.ID())
		assert.NotEqual(t, m1.ID(), m2.ID())
	})
}

func TestMessageAckTwice(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q := env.p.Topic("default")

		if !assert.NoError(t, q.Enqueue(ctx, []byte("a"))) {
			return
		}
		m, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		assert.NoError(t, m.Ack(ctx))
		assert.ErrorIs(t, m.Ack(ctx), dbpubsub.ErrMessageFinished)
		assert.ErrorIs(t, m.Nack(ctx), dbpubsub.ErrMessageFinished)
	})
}

func TestMessageNackTwice(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q := env.p.Topic("default")

		if !assert.NoError(t, q.Enqueue(ctx, []byte("a"))) {
			return
		}
		m, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		assert.NoError(t, m.Nack(ctx))
		assert.ErrorIs(t, m.Nack(ctx), dbpubsub.ErrMessageFinished)
		assert.ErrorIs(t, m.Ack(ctx), dbpubsub.ErrMessageFinished)
	})
}

func TestDequeueContextCanceled(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q := env.p.Topic("default")

		ctx, cancel := context.WithCancel(ctx)
		cancel()

		m, err := q.Dequeue(ctx)
		assert.Nil(t, m)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestDequeueUpdateError(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q := env.p.Topic("default")

		env.failing.Store(true)
		m, err := q.Dequeue(ctx)
		assert.Nil(t, m)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestMessageAckUpdateError(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q := env.p.Topic("default")

		if !assert.NoError(t, q.Enqueue(ctx, []byte("a"))) {
			return
		}
		m, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		env.failing.Store(true)
		err = m.Ack(ctx)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestMessageNackUpdateError(t *testing.T) {
	runMessageTest(t, func(t *testing.T, ctx context.Context, env messageTestEnv) {
		q := env.p.Topic("default")

		if !assert.NoError(t, q.Enqueue(ctx, []byte("a"))) {
			return
		}
		m, err := q.Dequeue(ctx)
		if !assert.NoError(t, err) {
			return
		}

		env.failing.Store(true)
		err = m.Nack(ctx)
		assert.ErrorIs(t, err, errBoom)
	})
}
