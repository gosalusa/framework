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

// blockingHandler signals that it has started, then blocks until released, so
// a test can observe whether the dequeue loop keeps running while a handler is
// still in flight. Listeners are built reflectively from their type, so the
// channels arrive through dependency injection rather than a captured value.
type blockingHandler struct {
	Started chan string   `inject:""`
	Release chan struct{} `inject:""`
}

func (h *blockingHandler) Handle(ctx context.Context, e *TestEvent1) error {
	h.Started <- e.Foo
	<-h.Release
	return nil
}

// recordingHandler reports the event it was handed so a test can check that
// concurrently handled messages do not cross-talk.
type recordingHandler struct {
	Seen chan string `inject:""`
}

func (h *recordingHandler) Handle(ctx context.Context, e *TestEvent1) error {
	h.Seen <- e.Foo
	return nil
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
	dequeues   int
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
	t.mu.Lock()
	t.dequeues++
	t.mu.Unlock()
	select {
	case m := <-t.ch:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// dequeueCount reports how many times the loop has pulled a message, which is
// how the tests observe whether handling a message blocks the loop.
func (t *fakeTopic) dequeueCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dequeues
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
			handlerType: reflect.TypeFor[testEventHandler](),
		}
		err := h.Run(context.Background(), di.NewDependencyProvider(), &TestEvent1{Foo: "bar"})
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
		dp.RegisterSingleton(func() *testDep {
			return &testDep{V: 7}
		})
		h := &handler[*TestEvent1]{
			handlerType: reflect.TypeFor[*testFillableHandler](),
		}
		err := h.Run(context.Background(), dp, &TestEvent1{Foo: "bar"})
		assert.NoError(t, err)
		assert.Equal(t, "bar:7", fillableResult)
	})

	t.Run("run fill error", func(t *testing.T) {
		h := &handler[*TestEvent1]{
			handlerType: reflect.TypeFor[*testFillableHandler](),
		}
		err := h.Run(context.Background(), di.NewDependencyProvider(), &TestEvent1{Foo: "bar"})
		assert.ErrorIs(t, err, di.ErrNotRegistered)
	})

	t.Run("mismatched event", func(t *testing.T) {
		testHandled = make(chan string, 1)
		h := &handler[*TestEvent1]{
			handlerType: reflect.TypeFor[testEventHandler](),
		}
		err := h.Run(context.Background(), di.NewDependencyProvider(), &TestEvent2{})
		assert.ErrorIs(t, err, ErrEventTypeMismatch)
		assert.Empty(t, testHandled, "a mismatched event should not reach the handler")
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

	t.Run("synchronous builder", func(t *testing.T) {
		s := Service()
		assert.False(t, s.synchronous, "handlers are concurrent by default")
		assert.Same(t, s, s.Synchronous(), "Synchronous should return the service for chaining")
		assert.True(t, s.synchronous)
	})
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

	// The default runs each handler in its own goroutine, so a slow handler
	// must not hold up the dequeue loop.
	t.Run("concurrent by default", func(t *testing.T) {
		ps := newFakePubSub()
		started := make(chan string, 4)
		release := make(chan struct{})
		s, _ := newService(t, ps, NewListener[*blockingHandler, *TestEvent1]())
		s.DP.RegisterSingleton(func() chan string { return started })
		s.DP.RegisterSingleton(func() chan struct{} { return release })

		first, err := encodeEvent(&TestEvent1{Foo: "first"})
		assert.NoError(t, err)
		second, err := encodeEvent(&TestEvent1{Foo: "second"})
		assert.NoError(t, err)
		ps.topic.ch <- &fakeMessage{id: "1", data: first}
		ps.topic.ch <- &fakeMessage{id: "2", data: second}

		cancel, done := startService(s)

		select {
		case got := <-started:
			// Both handlers run concurrently, so which one reports first is
			// not deterministic.
			assert.Contains(t, []string{"first", "second"}, got)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for a handler to start")
		}

		assert.True(t, waitFor(5*time.Second, func() bool {
			return ps.topic.dequeueCount() >= 2
		}), "the default mode should keep dequeuing while a handler is still running")

		close(release)
		stopService(t, cancel, done)
	})

	// Synchronous runs the handler on the loop's own goroutine, so nothing
	// else is dequeued until it returns.
	t.Run("synchronous blocks the loop", func(t *testing.T) {
		ps := newFakePubSub()
		started := make(chan string, 4)
		release := make(chan struct{})
		s, _ := newService(t, ps, NewListener[*blockingHandler, *TestEvent1]())
		s.DP.RegisterSingleton(func() chan string { return started })
		s.DP.RegisterSingleton(func() chan struct{} { return release })
		s.Synchronous()

		first, err := encodeEvent(&TestEvent1{Foo: "first"})
		assert.NoError(t, err)
		second, err := encodeEvent(&TestEvent1{Foo: "second"})
		assert.NoError(t, err)
		ps.topic.ch <- &fakeMessage{id: "1", data: first}
		ps.topic.ch <- &fakeMessage{id: "2", data: second}

		cancel, done := startService(s)

		select {
		case got := <-started:
			assert.Equal(t, "first", got)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for the first handler to start")
		}

		assert.False(t, waitFor(250*time.Millisecond, func() bool {
			return ps.topic.dequeueCount() >= 2
		}), "synchronous mode should not dequeue while a handler is still running")

		close(release)
		assert.True(t, waitFor(5*time.Second, func() bool {
			return ps.topic.dequeueCount() >= 2
		}), "synchronous mode should dequeue once the handler returns")

		stopService(t, cancel, done)
	})

	// Runners are shared by every message the service handles, so a
	// concurrently handled event must still reach the handler that was meant
	// to get it.
	t.Run("concurrent messages do not cross-talk", func(t *testing.T) {
		ps := newFakePubSub()
		seen := make(chan string, 16)
		s, _ := newService(t, ps, NewListener[*recordingHandler, *TestEvent1]())
		s.DP.RegisterSingleton(func() chan string { return seen })

		const total = 8
		for i := 0; i < total; i++ {
			b, err := encodeEvent(&TestEvent1{Foo: fmt.Sprintf("e%d", i)})
			assert.NoError(t, err)
			ps.topic.ch <- &fakeMessage{id: fmt.Sprintf("%d", i), data: b}
		}

		cancel, done := startService(s)

		counts := map[string]int{}
		for i := 0; i < total; i++ {
			select {
			case got := <-seen:
				counts[got]++
			case <-time.After(5 * time.Second):
				t.Fatalf("timed out waiting for handlers, handled %d of %d", i, total)
			}
		}
		stopService(t, cancel, done)

		for i := 0; i < total; i++ {
			assert.Equal(t, 1, counts[fmt.Sprintf("e%d", i)], "each event should be handled exactly once")
		}
	})
}
