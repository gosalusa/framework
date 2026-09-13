package builder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
)

func TestContext(t *testing.T) {
	q := NewTestBuilder()

	assert.Equal(t, context.Background(), q.Context())
	ctx := context.WithValue(context.Background(), "foo", "bar")
	q = q.WithContext(ctx).Where("1", "=", 2)

	assert.Same(t, ctx, q.Context())
	assert.NotNil(t, q.Context())
}

func TestContext_sub_builder(t *testing.T) {
	q := NewTestBuilder()

	assert.Equal(t, context.Background(), q.Context())

	q = q.WithContext(context.WithValue(context.Background(), "foo", "bar")).Where("1", "=", 2)

	q.WhereHas("Bar", func(q *builder.Builder) *builder.Builder {
		assert.NotEqual(t, context.Background(), q.Context())
		assert.NotNil(t, q.Context())
		return q
	})
}
