package clog_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/clog"
	"gosalusa.com/di"
)

func register() (context.Context, *bytes.Buffer) {
	ctx := di.ContextWithDependencyProvider(
		context.Background(),
		di.NewDependencyProvider(),
	)

	b := bytes.NewBuffer([]byte{})

	clog.RegisterWith(ctx, slog.NewTextHandler(b, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, time.Time{})
			}
			return a
		},
	}))

	return ctx, b
}

func TestWith(t *testing.T) {
	t.Run("with string", func(t *testing.T) {
		ctx, b := register()

		ctx = clog.With(ctx, slog.String("foo", "bar"))

		l, err := di.Resolve[*slog.Logger](ctx)
		assert.NoError(t, err)

		l.Warn("test")
		assert.Equal(t, "time=0001-01-01T00:00:00.000Z level=WARN msg=test foo=bar\n", b.String())
	})

	t.Run("with multiple", func(t *testing.T) {
		ctx, b := register()

		ctx = clog.With(ctx, slog.String("a", "1"))
		ctx = clog.With(ctx, slog.String("b", "2"))

		l, err := di.Resolve[*slog.Logger](ctx)
		assert.NoError(t, err)

		l.Warn("test")
		assert.Equal(t, "time=0001-01-01T00:00:00.000Z level=WARN msg=test a=1 b=2\n", b.String())
	})
}

func TestResolve(t *testing.T) {
	t.Run("no handler", func(t *testing.T) {
		ctx := di.ContextWithDependencyProvider(
			context.Background(),
			di.NewDependencyProvider(),
		)

		clog.RegisterDefault(ctx)

		l, err := di.Resolve[*slog.Logger](ctx)
		assert.NoError(t, err)
		assert.NotNil(t, l)
	})
}

type testConfig struct {
	level slog.Level
	b     *bytes.Buffer
}

func (c *testConfig) GetHTTPPort() int {
	return 8080
}
func (c *testConfig) GetBaseURL() string {
	return "https://example.com"
}
func (c *testConfig) LoggerConfig() clog.Config {
	return &testLoggerConfig{level: c.level, b: c.b}
}

type testLoggerConfig struct {
	level slog.Level
	b     *bytes.Buffer
}

func (c *testLoggerConfig) Handler() (slog.Handler, error) {
	return slog.NewTextHandler(c.b, &slog.HandlerOptions{
		Level: c.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, time.Time{})
			}
			return a
		},
	}), nil
}

func TestNewDefaultConfig(t *testing.T) {
	c := clog.NewDefaultConfig(slog.LevelWarn)
	assert.NotNil(t, c)
	h, err := c.Handler()
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

func TestDefaultHandler(t *testing.T) {
	t.Run("not a tty", func(t *testing.T) {
		h := clog.DefaultHandler(slog.LevelDebug)
		assert.IsType(t, &slog.TextHandler{}, h)
	})

	t.Run("tty", func(t *testing.T) {
		devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
		assert.NoError(t, err)
		defer devNull.Close()

		origStderr := os.Stderr
		os.Stderr = devNull
		defer func() { os.Stderr = origStderr }()

		h := clog.DefaultHandler(slog.LevelDebug)
		_, isText := h.(*slog.TextHandler)
		assert.False(t, isText)

		r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
		r.AddAttrs(
			slog.String("foo", "bar"),
			slog.Any("err", errors.New("boom")),
		)
		assert.NoError(t, h.Handle(context.Background(), r))
	})
}

func TestRegister(t *testing.T) {
	t.Run("without logger config", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		clog.RegisterDefault(ctx)

		logger, err := di.Resolve[*clog.RootLogger](ctx)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("with config", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		ctx := di.ContextWithDependencyProvider(
			context.Background(),
			di.NewDependencyProvider(),
		)

		clog.Register(ctx, &testLoggerConfig{level: slog.LevelWarn, b: b})

		logger, err := di.Resolve[*clog.RootLogger](ctx)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		l, err := di.Resolve[*slog.Logger](ctx)
		assert.NoError(t, err)
		assert.NotNil(t, l)

		l.Warn("test")
		assert.Equal(t, "time=0001-01-01T00:00:00.000Z level=WARN msg=test\n", b.String())

		b.Reset()
		l.Info("test")
		assert.Empty(t, b.String())
	})

	t.Run("handler error", func(t *testing.T) {
		ctx := di.ContextWithDependencyProvider(
			context.Background(),
			di.NewDependencyProvider(),
		)

		clog.Register(ctx, errLoggerConfig{})

		_, err := di.Resolve[*clog.RootLogger](ctx)
		assert.ErrorIs(t, err, errTestHandler)
	})
}

var errTestHandler = errors.New("handler error")

type errLoggerConfig struct{}

func (errLoggerConfig) Handler() (slog.Handler, error) {
	return nil, errTestHandler
}

type plainConfig struct{}

func (plainConfig) GetHTTPPort() int {
	return 8080
}
func (plainConfig) GetBaseURL() string {
	return "https://example.com"
}

func TestUse(t *testing.T) {
	t.Run("resolves", func(t *testing.T) {
		ctx, _ := register()
		logger := clog.Use(ctx)
		assert.NotNil(t, logger)
	})

	t.Run("fallback to default", func(t *testing.T) {
		ctx := di.ContextWithDependencyProvider(
			context.Background(),
			di.NewDependencyProvider(),
		)
		logger := clog.Use(ctx)
		assert.NotNil(t, logger)
	})
}
