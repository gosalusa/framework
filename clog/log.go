package clog

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"gosalusa.com/di"
)

type key uint8

const (
	withKey key = iota
)

type RootLogger slog.Logger

type Config interface {
	Handler() (slog.Handler, error)
}

type DefaultConfig struct {
	Level slog.Level
}

var _ Config = (*DefaultConfig)(nil)

func NewDefaultConfig(level slog.Level) *DefaultConfig {
	return &DefaultConfig{Level: level}
}

func (c *DefaultConfig) Handler() (slog.Handler, error) {
	return DefaultHandler(c.Level), nil
}

func RegisterDefault(ctx context.Context) {
	di.RegisterLazySingleton(ctx, func() (*RootLogger, error) {
		logger := slog.New(DefaultHandler(slog.LevelInfo))

		slog.SetDefault(logger)
		return (*RootLogger)(logger), nil
	})

	registerLogger(ctx)
}
func Register(ctx context.Context, cfg Config) {
	di.RegisterLazySingleton(ctx, func() (*RootLogger, error) {
		h, err := cfg.Handler()
		if err != nil {
			return nil, err
		}

		logger := slog.New(h)

		slog.SetDefault(logger)
		return (*RootLogger)(logger), nil
	})

	registerLogger(ctx)
}
func RegisterWith(ctx context.Context, h slog.Handler) {
	di.RegisterLazySingleton(ctx, func() (*RootLogger, error) {
		logger := slog.New(h)

		slog.SetDefault(logger)
		return (*RootLogger)(logger), nil
	})

	registerLogger(ctx)
}
func DefaultHandler(level slog.Level) slog.Handler {
	fi, err := os.Stderr.Stat()
	isTTY := err == nil && (fi.Mode()&os.ModeCharDevice) != 0
	if !isTTY {
		return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
	}
	return tint.NewTextHandler(os.Stderr, &tint.Options{
		Level: level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			err, ok := attr.Value.Any().(error)
			if !ok {
				return attr
			}
			errAttr := tint.Err(err)
			errAttr.Key = attr.Key
			return errAttr
		},
	})
}
func registerLogger(ctx context.Context) {
	di.RegisterWith(ctx, func(ctx context.Context, tag string, logger *RootLogger) (*slog.Logger, error) {
		with := ctx.Value(withKey)
		sLogger := (*slog.Logger)(logger)
		if with != nil {
			return sLogger.With(with.([]any)...), nil
		}
		return sLogger, nil
	})
}

func Use(ctx context.Context) *slog.Logger {
	logger, err := di.Resolve[*slog.Logger](ctx)
	if err != nil {
		return slog.Default()
	}
	return logger
}

func With(ctx context.Context, attrs ...slog.Attr) context.Context {
	with := get(ctx)

	for _, attr := range attrs {
		with = append(with, attr)
	}
	return context.WithValue(ctx, withKey, with)
}

func get(ctx context.Context) []any {
	iWith := ctx.Value(withKey)
	if iWith == nil {
		return []any{}
	}
	with, ok := iWith.([]any)
	if !ok {
		return []any{}
	}
	return with
}
