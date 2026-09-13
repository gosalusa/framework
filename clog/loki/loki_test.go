package loki_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/clog"
	"gosalusa.com/clog/loki"
)

func TestConfig_Handler(t *testing.T) {
	var _ clog.Config = (*loki.Config)(nil)

	t.Run("valid url", func(t *testing.T) {
		c := &loki.Config{
			Level:    slog.LevelInfo,
			URL:      "https://loki.example.com/loki/api/v1/push",
			TenantID: "tenant-1",
		}
		h, err := c.Handler()
		assert.NoError(t, err)
		assert.NotNil(t, h)
	})

	t.Run("invalid url", func(t *testing.T) {
		c := &loki.Config{
			Level: slog.LevelDebug,
			URL:   ":::not-a-url:::",
		}
		h, err := c.Handler()
		assert.Error(t, err)
		assert.Nil(t, h)
	})
}
