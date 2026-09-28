package cron

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/event"
)

type testEvent struct {
	CronEvent
}

func (e *testEvent) Type() event.EventType {
	return "test"
}

// sharedFieldEvent has a reference-typed field, to show that cloning is
// shallow and does not duplicate it.
type sharedFieldEvent struct {
	CronEvent
	Tags []string
}

func (e *sharedFieldEvent) Type() event.EventType {
	return "shared"
}

// valueEvent implements Event on its value type, so it can be handed to
// Schedule even though its fire time can never actually be set.
type valueEvent struct {
	CronEvent
}

func (e valueEvent) Type() event.EventType {
	return "value"
}

func (e valueEvent) SetTime(time.Time) {}

func TestCronEvent_SetTime(t *testing.T) {
	e := &testEvent{}
	now := time.Now()
	e.SetTime(now)
	assert.Equal(t, now, e.Time)
}

func TestServiceAndName(t *testing.T) {
	s := Service()
	assert.Equal(t, "cron-service", s.Name())
	assert.Empty(t, s.events)
}

func TestSchedule(t *testing.T) {
	s := Service()
	e1 := &testEvent{}
	e2 := &testEvent{}

	result := s.Schedule("* * * * *", e1)
	assert.Same(t, s, result)
	assert.Len(t, s.events["* * * * *"], 1)

	s.Schedule("* * * * *", e2)
	assert.Len(t, s.events["* * * * *"], 2)

	s.Schedule("*/5 * * * *", e1)
	assert.Len(t, s.events["*/5 * * * *"], 1)
}

func TestRun(t *testing.T) {
	s := Service()
	s.Dispatch = func(ctx context.Context, e event.Event) error {
		return nil
	}
	s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	s.Schedule("* * * * *", &testEvent{})
	s.Schedule("bad-spec", &testEvent{})

	err := s.Run(context.Background())
	assert.NoError(t, err)
}

func TestRunFiresEvent(t *testing.T) {
	ch := make(chan event.Event, 1)
	s := Service()
	s.Dispatch = func(ctx context.Context, e event.Event) error {
		ch <- e
		return nil
	}
	s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	s.Schedule("@every 100ms", &testEvent{})

	err := s.Run(context.Background())
	assert.NoError(t, err)

	select {
	case e := <-ch:
		te, ok := e.(*testEvent)
		assert.True(t, ok)
		assert.False(t, te.Time.IsZero())
	case <-time.After(2 * time.Second):
		t.Fatal("expected cron event to fire")
	}
}

func TestCloneEvent(t *testing.T) {
	t.Run("copies value fields", func(t *testing.T) {
		now := time.Now()
		orig := &testEvent{}
		orig.Time = now

		clone, err := cloneEvent(orig)
		assert.NoError(t, err)
		assert.Equal(t, now, clone.(*testEvent).Time)
		assert.NotSame(t, orig, clone, "a firing must not reuse the registered instance")

		// The copy is independent, so setting the time on it leaves the
		// template untouched.
		clone.SetTime(now.Add(time.Hour))
		assert.Equal(t, now, orig.Time)
	})

	t.Run("shares reference fields", func(t *testing.T) {
		// A shallow copy keeps reference-typed fields shared with the
		// template, which is why handlers must not mutate them unsafely.
		clone, err := cloneEvent(&sharedFieldEvent{Tags: []string{"a"}})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a"}, clone.(*sharedFieldEvent).Tags)
	})

	t.Run("rejects non-pointer", func(t *testing.T) {
		_, err := cloneEvent(valueEvent{})
		assert.Error(t, err, "a value cannot have its fire time set")
	})

	t.Run("rejects nil pointer", func(t *testing.T) {
		_, err := cloneEvent((*testEvent)(nil))
		assert.Error(t, err, "a nil pointer cannot be cloned")
	})
}

// robfig/cron runs every job in its own goroutine, so a tick that fires while
// the previous dispatch is still in flight puts two goroutines on the same
// registered event at once. Each firing has to get its own copy, otherwise they
// race setting the fire time and can read each other's value.
func TestRunOverlappingFiringsAreIndependent(t *testing.T) {
	const (
		tick    = 10 * time.Millisecond
		overlap = 3
	)

	entered := make(chan struct{}, 64)
	fired := make(chan *testEvent, 64)
	release := make(chan struct{})

	s := Service()
	s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	s.Dispatch = func(ctx context.Context, e event.Event) error {
		select {
		case entered <- struct{}{}:
		default:
		}
		// Hold the dispatch open so the next tick fires while this one is
		// still running.
		select {
		case <-release:
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
		select {
		case fired <- e.(*testEvent):
		default:
		}
		return nil
	}

	s.Schedule("@every "+tick.String(), &testEvent{})
	assert.NoError(t, s.Run(context.Background()))

	// None of these are released yet, so reaching the target means that many
	// firings are inside the dispatch simultaneously.
	deadline := time.After(5 * time.Second)
	for i := 0; i < overlap; i++ {
		select {
		case <-entered:
		case <-deadline:
			t.Fatalf("only %d of %d firings overlapped", i, overlap)
		}
	}
	close(release)

	distinct := map[*testEvent]struct{}{}
	for i := 0; i < overlap; i++ {
		select {
		case e := <-fired:
			assert.False(t, e.Time.IsZero(), "every firing should carry its own fire time")
			distinct[e] = struct{}{}
		case <-time.After(5 * time.Second):
			t.Fatalf("only %d of %d firings completed", i, overlap)
		}
	}

	assert.GreaterOrEqual(t, len(distinct), overlap,
		"overlapping firings must each dispatch their own event instance, not a shared one")
}

// The same event registered under two specs produces two independent cron
// jobs, so it has to be copied per firing just the same.
func TestRunSameEventUnderTwoSpecs(t *testing.T) {
	fired := make(chan *testEvent, 64)

	s := Service()
	s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	s.Dispatch = func(ctx context.Context, e event.Event) error {
		select {
		case fired <- e.(*testEvent):
		default:
		}
		return nil
	}

	shared := &testEvent{}
	s.Schedule("@every 5ms", shared)
	s.Schedule("@every 7ms", shared)
	assert.NoError(t, s.Run(context.Background()))

	distinct := map[*testEvent]struct{}{}
	deadline := time.After(5 * time.Second)
	for i := 0; i < 6; i++ {
		select {
		case e := <-fired:
			distinct[e] = struct{}{}
		case <-deadline:
		}
	}

	assert.GreaterOrEqual(t, len(distinct), 2,
		"the same registered event under two specs must dispatch a fresh instance per firing")
}
