package di_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
)

type Generic[T any] func(value T) T

func TestResolve(t *testing.T) {
	t.Run("context", func(t *testing.T) {

		expectedContext := di.ContextWithDependencyProvider(
			context.WithValue(context.Background(), "foo", "bar"),
			di.NewDependencyProvider(),
		)
		ctx, err := di.Resolve[context.Context](expectedContext)

		assert.NoError(t, err)
		assert.Same(t, expectedContext, ctx)
	})

	t.Run("self", func(t *testing.T) {
		expectedDP := di.NewDependencyProvider()
		ctx := di.ContextWithDependencyProvider(
			context.Background(),
			expectedDP,
		)
		dp, err := di.Resolve[*di.DependencyProvider](ctx)

		assert.NoError(t, err)
		assert.Same(t, expectedDP, dp)
	})

	t.Run("error", func(t *testing.T) {
		type Struct struct{ V int }
		ctx := di.TestDependencyProviderContext()
		resolveErr := fmt.Errorf("resolve error")
		di.Register(ctx, func(ctx context.Context, tag string) (*Struct, error) {
			return nil, resolveErr
		})

		v, err := di.Resolve[*Struct](ctx)

		assert.Same(t, resolveErr, err)
		assert.Zero(t, v)
	})

	t.Run("error in fill", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable struct {
			S *Struct `inject:""`
		}
		ctx := di.TestDependencyProviderContext()
		resolveErr := fmt.Errorf("resolve error")
		di.Register(ctx, func(ctx context.Context, tag string) (*Struct, error) {
			return nil, resolveErr
		})

		v, err := di.Resolve[*Fillable](ctx)

		assert.ErrorIs(t, err, resolveErr)
		assert.Zero(t, v)
	})
}

func TestResolveFill(t *testing.T) {
	t.Run("context", func(t *testing.T) {

		expectedContext := di.ContextWithDependencyProvider(
			context.WithValue(context.Background(), "foo", "bar"),
			di.NewDependencyProvider(),
		)
		ctx, err := di.Resolve[context.Context](expectedContext)

		assert.NoError(t, err)
		assert.Same(t, expectedContext, ctx)
	})

	t.Run("fill", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable struct {
			S *Struct `inject:""`
		}
		ctx := di.TestDependencyProviderContext()
		di.Register(ctx, func(ctx context.Context, tag string) (*Struct, error) {
			return &Struct{}, nil
		})

		v, err := di.Resolve[*Fillable](ctx)

		assert.NoError(t, err)
		assert.NotNil(t, v)
	})

	t.Run("error in fill", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable struct {
			S *Struct `inject:""`
		}
		ctx := di.TestDependencyProviderContext()
		resolveErr := fmt.Errorf("resolve error")
		di.Register(ctx, func(ctx context.Context, tag string) (*Struct, error) {
			return nil, resolveErr
		})

		v, err := di.Resolve[*Fillable](ctx)

		assert.ErrorIs(t, err, resolveErr)
		assert.Zero(t, v)
	})

}
