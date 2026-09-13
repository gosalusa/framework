package event

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/pubsub"
)

var errBoom = errors.New("boom")

var (
	testHandled    chan string
	fillableResult string
)

type testDep struct{ V int }

type testEventHandler struct{}

func (testEventHandler) Handle(ctx context.Context, e *TestEvent1) error {
	testHandled <- e.Foo
	return nil
}

type testFailingHandler struct{}

func (testFailingHandler) Handle(ctx context.Context, e *TestEvent1) error {
	return errBoom
}

type testEventHandler2 struct{}

func (testEventHandler2) Handle(ctx context.Context, e *TestEvent2) error {
	return nil
}

type orphanSource struct {
	Common string
}

func (*orphanSource) Type() EventType {
	return "orphan-source"
}

type orphanTarget struct {
	Common string
}

func (*orphanTarget) Type() EventType {
	return "orphan-target"
}

type testFillableHandler struct {
	Dep *testDep `inject:""`
}

func (h *testFillableHandler) Handle(ctx context.Context, e *TestEvent1) error {
	fillableResult = fmt.Sprintf("%s:%d", e.Foo, h.Dep.V)
	return nil
}

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type fakeMessage struct {
	id   string
	data []byte

	mu    sync.Mutex
	acks  int
	nacks int
}

func (m *fakeMessage) ID() string   { return m.id }
func (m *fakeMessage) Data() []byte { return m.data }

func (m *fakeMessage) Ack(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.acks++
	return nil
}

func (m *fakeMessage) Nack(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nacks++
	return nil
}

func (m *fakeMessage) ackCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.acks
}

type fakeTopic struct {
	ch         chan pubsub.Message
	mu         sync.Mutex
	enqueued   [][]byte
	enqueueErr error
	closes     int
}

var _ pubsub.Topic = (*fakeTopic)(nil)

func newFakeTopic() *fakeTopic {
	return &fakeTopic{ch: make(chan pubsub.Message, 16)}
}

func (t *fakeTopic) Enqueue(ctx context.Context, data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.enqueueErr != nil {
		return t.enqueueErr
	}
	t.enqueued = append(t.enqueued, data)
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
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closes++
	return nil
}

func (t *fakeTopic) enqueuedData() [][]byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	data := make([][]byte, len(t.enqueued))
	copy(data, t.enqueued)
	return data
}

type fakePubSub struct {
	topic *fakeTopic

	mu    sync.Mutex
	names []string
}

var _ pubsub.PubSub = (*fakePubSub)(nil)

func newFakePubSub() *fakePubSub {
	return &fakePubSub{topic: newFakeTopic()}
}

func (p *fakePubSub) Topic(name string) pubsub.Topic {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.names = append(p.names, name)
	return p.topic
}

func (p *fakePubSub) topicNames() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string{}, p.names...)
}

func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

func startService(s *EventService) (context.CancelFunc, chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()
	return cancel, done
}

func stopService(t *testing.T, cancel func(), done chan error) {
	t.Helper()
	cancel()
	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for service to stop")
	}
}

func TestHandler(t *testing.T) {
	t.Run("run", func(t *testing.T) {
		testHandled = make(chan string, 1)
		h := &handler[*TestEvent1]{
			value:       &TestEvent1{Foo: "bar"},
			handlerType: reflect.TypeFor[testEventHandler](),
		}
		err := h.Run(context.Background(), di.NewDependencyProvider())
		assert.NoError(t, err)

		select {
		case v := <-testHandled:
			assert.Equal(t, "bar", v)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for handler")
		}
	})

	t.Run("run fillable", func(t *testing.T) {
		dp := di.NewDependencyProvider()
		dp.Register(di.NewSingletonFactory(&testDep{V: 7}))
		h := &handler[*TestEvent1]{
			value:       &TestEvent1{Foo: "bar"},
			handlerType: reflect.TypeFor[*testFillableHandler](),
		}
		err := h.Run(context.Background(), dp)
		assert.NoError(t, err)
		assert.Equal(t, "bar:7", fillableResult)
	})

	t.Run("run fill error", func(t *testing.T) {
		h := &handler[*TestEvent1]{
			value:       &TestEvent1{Foo: "bar"},
			handlerType: reflect.TypeFor[*testFillableHandler](),
		}
		err := h.Run(context.Background(), di.NewDependencyProvider())
		assert.ErrorIs(t, err, di.ErrNotRegistered)
	})

	t.Run("update value", func(t *testing.T) {
		h := &handler[*TestEvent1]{}
		assert.False(t, h.UpdateValue(&TestEvent2{}))
		assert.Nil(t, h.value)
		assert.True(t, h.UpdateValue(&TestEvent1{Foo: "x"}))
		assert.Equal(t, "x", h.value.Foo)
	})

	t.Run("event type", func(t *testing.T) {
		h := &handler[*TestEvent1]{}
		assert.Equal(t, reflect.TypeOf(&TestEvent1{}), h.EventType())
	})
}

func TestNewListener(t *testing.T) {
	l := NewListener[testEventHandler, *TestEvent1]()
	assert.Equal(t, EventType("test-event:1"), l.eventType)
	assert.NotNil(t, l.runner)
	assert.Equal(t, reflect.TypeOf(&TestEvent1{}), l.runner.EventType())
}

func TestService(t *testing.T) {
	s := Service(
		NewListener[testEventHandler, *TestEvent1](),
		NewListener[testFailingHandler, *TestEvent1](),
		NewListener[testEventHandler2, *TestEvent2](),
	)
	assert.Equal(t, "event-service", s.Name())
	assert.Len(t, s.listeners, 2)
	assert.Len(t, s.listeners[(&TestEvent1{}).Type()], 2)
	assert.Len(t, s.listeners[(&TestEvent2{}).Type()], 1)
}

func TestEventServiceRun(t *testing.T) {
	newService := func(t *testing.T, ps *fakePubSub, listeners ...*Listener) (*EventService, *safeBuffer) {
		t.Helper()
		logs := &safeBuffer{}
		s := Service(listeners...)
		s.PubSub = ps
		s.Logger = slog.New(slog.NewTextHandler(logs, nil))
		s.DP = di.NewDependencyProvider()
		return s, logs
	}

	t.Run("handles event", func(t *testing.T) {
		ps := newFakePubSub()
		testHandled = make(chan string, 1)
		s, logs := newService(t, ps, NewListener[testEventHandler, *TestEvent1]())

		b, err := encodeEvent(&TestEvent1{Foo: "hi"})
		assert.NoError(t, err)
		msg := &fakeMessage{id: "1", data: b}
		ps.topic.ch <- msg

		cancel, done := startService(s)

		var got string
		select {
		case got = <-testHandled:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for event to be handled")
		}
		assert.Equal(t, "hi", got)
		assert.True(t, waitFor(5*time.Second, func() bool {
			return msg.ackCount() == 1
		}), "message was not acked")

		stopService(t, cancel, done)

		assert.Contains(t, logs.String(), "could not dequeue event")
		assert.Equal(t, 1, ps.topic.closes)
		assert.Equal(t, []string{""}, ps.topicNames())
	})

	t.Run("decode failure", func(t *testing.T) {
		ps := newFakePubSub()
		testHandled = make(chan string, 1)
		s, logs := newService(t, ps, NewListener[testEventHandler, *TestEvent1]())

		ps.topic.ch <- &fakeMessage{id: "1", data: []byte("test-event:1|notgobdata")}
		ps.topic.ch <- &fakeMessage{id: "2", data: []byte("unknown-type|junk")}

		cancel, done := startService(s)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return strings.Count(logs.String(), "could not decode event") >= 2
		}), "expected decode failures to be logged")

		stopService(t, cancel, done)
		assert.Empty(t, testHandled)
	})

	t.Run("no listeners for event", func(t *testing.T) {
		ps := newFakePubSub()
		s, logs := newService(t, ps, &Listener{
			eventType: (&orphanSource{}).Type(),
			runner:    &handler[*orphanTarget]{},
		})

		b, err := encodeEvent(&orphanSource{Common: "orphan"})
		assert.NoError(t, err)
		ps.topic.ch <- &fakeMessage{id: "1", data: b}

		cancel, done := startService(s)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return strings.Contains(logs.String(), "no listeners for event with matching type")
		}), "expected missing listener warning")

		stopService(t, cancel, done)
	})

	t.Run("mismatched event and type", func(t *testing.T) {
		ps := newFakePubSub()
		testHandled = make(chan string, 1)
		s, logs := newService(t, ps,
			NewListener[testEventHandler, *TestEvent1](),
			&Listener{
				eventType: (&TestEvent1{}).Type(),
				runner:    &handler[*TestEvent2]{},
			},
		)

		b, err := encodeEvent(&TestEvent1{Foo: "ok"})
		assert.NoError(t, err)
		ps.topic.ch <- &fakeMessage{id: "1", data: b}

		cancel, done := startService(s)

		select {
		case got := <-testHandled:
			assert.Equal(t, "ok", got)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for event to be handled")
		}
		assert.True(t, waitFor(5*time.Second, func() bool {
			return strings.Contains(logs.String(), "mismatched event and type, there may be a conflict")
		}), "expected mismatch warning")

		stopService(t, cancel, done)
	})

	t.Run("handler failure", func(t *testing.T) {
		ps := newFakePubSub()
		s, logs := newService(t, ps, NewListener[testFailingHandler, *TestEvent1]())

		b, err := encodeEvent(&TestEvent1{Foo: "fail"})
		assert.NoError(t, err)
		msg := &fakeMessage{id: "1", data: b}
		ps.topic.ch <- msg

		cancel, done := startService(s)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return strings.Contains(logs.String(), "handler failed")
		}), "expected handler failure warning")
		assert.True(t, waitFor(5*time.Second, func() bool {
			return msg.ackCount() == 1
		}), "message was not acked")

		stopService(t, cancel, done)
	})
}
