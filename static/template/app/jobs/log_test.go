package jobs

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"gosalusa.com/di"
	"gosalusa.com/static/template/app/events"
)

func TestLogJob_Handle(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	di.RegisterSingleton(ctx, func() *slog.Logger {
		return slog.Default()
	})

	l := &LogJob{}
	require.NoError(t, di.Fill(ctx, l))

	err := l.Handle(context.Background(), &events.LogEvent{Message: "test"})
	require.NoError(t, err)
}
