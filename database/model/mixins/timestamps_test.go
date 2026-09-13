package mixins_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/model/mixins"
)

func TestTimestampsBeforeSave(t *testing.T) {
	t.Run("sets created and updated", func(t *testing.T) {
		ts := &mixins.Timestamps{}
		err := ts.BeforeSave(context.Background(), nil)
		assert.NoError(t, err)
		assert.False(t, ts.CreatedAt.IsZero())
		assert.False(t, ts.UpdatedAt.IsZero())
		assert.Equal(t, ts.CreatedAt, ts.UpdatedAt)
	})

	t.Run("preserves existing created", func(t *testing.T) {
		ts := &mixins.Timestamps{
			CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		err := ts.BeforeSave(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), ts.CreatedAt)
		assert.False(t, ts.UpdatedAt.IsZero())
	})
}
