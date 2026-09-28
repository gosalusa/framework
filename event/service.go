package event

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"gosalusa.com/di"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/kernel"
	"gosalusa.com/pubsub"
)

type Handler[E Event] interface {
	Handle(ctx context.Context, event E) error
}

// ErrEventTypeMismatch is returned by a runner when the decoded event is not
// the type the runner was registered for.
var ErrEventTypeMismatch = errors.New("event type does not match runner")

// runner invokes a handler for a single decoded event. Runners are shared by
// every message the service handles, so they must not hold per-message state:
// the event is passed in rather than stored.
type runner interface {
	Run(ctx context.Context, dp *di.DependencyProvider, e Event) error
	EventType() reflect.Type
}
type Listener struct {
	eventType EventType
	runner    runner
}

type handler[E Event] struct {
	handlerType reflect.Type
}

// Run builds a handler for the event and invokes it. It returns
// ErrEventTypeMismatch without building a handler if e is not an E.
func (j *handler[E]) Run(ctx context.Context, dp *di.DependencyProvider, e Event) error {
	v, ok := e.(E)
	if !ok {
		return ErrEventTypeMismatch
	}

	t := j.handlerType
	h := helpers.Create(t).Interface().(Handler[E])

	if di.IsFillable(h) {
		err := dp.Fill(ctx, h)
		if err != nil {
			return err
		}
	}
	return h.Handle(ctx, v)
}

func (j *handler[E]) EventType() reflect.Type {
	var e E
	return reflect.TypeOf(e)
}

func NewListener[H Handler[E], E Event]() *Listener {
	var e E

	return &Listener{
		eventType: e.Type(),
		runner: &handler[E]{
			handlerType: reflect.TypeFor[H](),
		},
	}
}

type EventService struct {
	PubSub pubsub.PubSub          `inject:""`
	Logger *slog.Logger           `inject:""`
	DP     *di.DependencyProvider `inject:""`

	listeners   map[EventType][]runner
	topic       string
	synchronous bool
}

var _ kernel.Service = (*EventService)(nil)

func Service(listeners ...*Listener) *EventService {
	s := &EventService{
		listeners: map[EventType][]runner{},
	}
	for _, l := range listeners {
		jobs, ok := s.listeners[l.eventType]
		if !ok {
			jobs = []runner{}
		}
		s.listeners[l.eventType] = append(jobs, l.runner)
	}
	return s
}

func (s *EventService) Name() string {
	return "event-service"
}
// Synchronous makes the service handle each message inline, on the dequeue
// loop's own goroutine, rather than dispatching it to a new one. The loop then
// waits for the handler to return before dequeuing the next message. By default
// handlers run concurrently, one goroutine per message.
func (s *EventService) Synchronous() *EventService {
	s.synchronous = true
	return s
}

func (s *EventService) Run(ctx context.Context) error {

	events := map[EventType]reflect.Type{}
	for eventType, runners := range s.listeners {
		events[eventType] = runners[0].EventType()
	}

	t := s.PubSub.Topic(s.topic)
	defer t.Close()

	fails := 0
	for ctx.Err() == nil {
		m, err := t.Dequeue(ctx)
		if err != nil {
			s.Logger.Warn("could not dequeue event", "error", err)

			if ctx.Err() == nil {
				time.Sleep(helpers.ExponentialFalloff(fails))
				fails++
			}
			continue
		}
		fails = 0

		if s.synchronous {
			s.run(ctx, m, events)
		} else {
			go s.run(ctx, m, events)
		}
	}
	return ctx.Err()
}

func (s *EventService) run(ctx context.Context, m pubsub.Message, events map[EventType]reflect.Type) {
	defer m.Ack(ctx)

	e, err := decodeEvent(m.Data(), events)
	if err != nil {
		s.Logger.Warn("could not decode event", "error", err)
		return
	}
	runners, ok := s.listeners[e.Type()]
	if !ok {
		s.Logger.Warn("no listeners for event with matching type", slog.Any("type", e.Type()))
		return
	}

	for _, r := range runners {
		err := r.Run(ctx, s.DP, e)
		switch {
		case errors.Is(err, ErrEventTypeMismatch):
			s.Logger.Warn("mismatched event and type, there may be a conflict")
		case err != nil:
			s.Logger.Warn("handler failed", slog.Any("error", err))
		}
	}
}
