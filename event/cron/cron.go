package cron

import (
	"context"
	"log/slog"
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

func (c *CronService) Run(ctx context.Context) error {
	runner := cron.New()
	for spec, events := range c.events {
		for _, e := range events {
			_, err := runner.AddFunc(spec, func() {
				e.SetTime(time.Now())
				err := c.Dispatch(ctx, e)
				if err != nil {
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
