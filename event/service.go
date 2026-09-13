package event

import (
	"context"
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

type runner interface {
	UpdateValue(v Event) bool
	Run(ctx context.Context, dp *di.DependencyProvider) error
	EventType() reflect.Type
}
type Listener struct {
	eventType EventType
	runner    runner
}

type handler[E Event] struct {
	value       E
	handlerType reflect.Type
}

func (j *handler[E]) Run(ctx context.Context, dp *di.DependencyProvider) error {
	t := j.handlerType
	h := helpers.Create(t).Interface().(Handler[E])

	if di.IsFillable(h) {
		err := dp.Fill(ctx, h)
		if err != nil {
			return err
		}
	}
	return h.Handle(ctx, j.value)
}

func (j *handler[E]) UpdateValue(v Event) bool {
	ev, ok := v.(E)
	if !ok {
		return false
	}
	j.value = ev
	return true
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
			value:       e,
			handlerType: reflect.TypeFor[H](),
		},
	}
}

type EventService struct {
	PubSub pubsub.PubSub          `inject:""`
	Logger *slog.Logger           `inject:""`
	DP     *di.DependencyProvider `inject:""`

	listeners map[EventType][]runner
	topic     string
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

		go s.run(ctx, m, events)
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
		if r.UpdateValue(e) {
			err := r.Run(ctx, s.DP)
			if err != nil {
				s.Logger.Warn("handler failed", slog.Any("error", err))
			}
		} else {
			s.Logger.Warn("mismatched event and type, there may be a conflict")
		}
	}
}
