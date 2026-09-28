package cron

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/robfig/cron/v3"
	"gosalusa.com/event"
	"gosalusa.com/kernel"
)

type Event interface {
	event.Event
	SetTime(t time.Time)
}

type CronEvent struct {
	Time time.Time
}

func (b *CronEvent) SetTime(t time.Time) {
	b.Time = t
}

type CronService struct {
	Dispatch event.Dispatch `inject:""`
	Logger   *slog.Logger   `inject:""`

	events map[string][]Event
}

var _ kernel.Service = (*CronService)(nil)

func Service() *CronService {
	return &CronService{
		events: map[string][]Event{},
	}
}

func (c *CronService) Name() string {
	return "cron-service"
}

// cloneEvent returns a shallow copy of e so that every firing of a scheduled
// event works on its own instance. cron runs each job in its own goroutine, so
// sharing a single instance across firings would race on the fire time and let
// one firing observe another's. The copy is shallow: value fields, including
// the fire time, are duplicated, but fields that are themselves references stay
// shared with the template.
func cloneEvent(e Event) (Event, error) {
	v := reflect.ValueOf(e)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return nil, fmt.Errorf("cron: event %T must be a non-nil pointer to be cloned", e)
	}
	clone := reflect.New(v.Type().Elem())
	clone.Elem().Set(v.Elem())
	out, ok := clone.Interface().(Event)
	if !ok {
		return nil, fmt.Errorf("cron: cloned %T does not implement Event", e)
	}
	return out, nil
}

func (c *CronService) Run(ctx context.Context) error {
	runner := cron.New()
	for spec, events := range c.events {
		for _, e := range events {
			_, err := runner.AddFunc(spec, func() {
				fire, err := cloneEvent(e)
				if err != nil {
					c.Logger.Error("failed to prepare cron event", slog.Any("error", err))
					return
				}
				fire.SetTime(time.Now())
				if err := c.Dispatch(ctx, fire); err != nil {
					c.Logger.Error("failed to dispatch event", slog.Any("error", err))
				}
			})
			if err != nil {
				c.Logger.Error("failed to start cron listener", slog.Any("error", err))
			}
		}
	}
	runner.Start()

	return nil
}

func (c *CronService) Schedule(cron string, e Event) *CronService {
	jobs, ok := c.events[cron]
	if !ok {
		jobs = []Event{}
	}
	c.events[cron] = append(jobs, e)
	return c
}
