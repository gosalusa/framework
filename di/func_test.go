package di_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
)

func TestFunc(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type Struct struct{ V int }
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{V: 10}
		})

		fn := di.PrepareFuncCtx[func(ctx context.Context, i int) int](func(ctx context.Context, i int, s *Struct) int {
			return i + s.V
		})

		result := fn(ctx, 7)

		assert.Equal(t, 17, result)
	})

	t.Run("with error return, dependency missing", func(t *testing.T) {
		type Struct struct{ V int }
		ctx := di.TestDependencyProviderContext()

		fn := di.PrepareFuncCtx[func(ctx context.Context, i int) (int, error)](func(ctx context.Context, i int, s *Struct) (int, error) {
			return i + s.V, nil
		})

		result, err := fn(ctx, 7)

		assert.Error(t, err)
		assert.Zero(t, result)
	})

	t.Run("without error return, dependency missing panics", func(t *testing.T) {
		type Struct struct{ V int }
		ctx := di.TestDependencyProviderContext()

		fn := di.PrepareFuncCtx[func(ctx context.Context, i int) int](func(ctx context.Context, i int, s *Struct) int {
			return i + s.V
		})

		assert.Panics(t, func() {
			fn(ctx, 7)
		})
	})

	t.Run("non function type parameter panics", func(t *testing.T) {
		assert.Panics(t, func() {
			di.PrepareFuncCtx[int](func() {})
		})
	})

	t.Run("non function argument panics", func(t *testing.T) {
		assert.Panics(t, func() {
			di.PrepareFuncCtx[func(ctx context.Context)](42)
		})
	})

	t.Run("return counts do not match panics", func(t *testing.T) {
		assert.Panics(t, func() {
			di.PrepareFuncCtx[func(ctx context.Context) int](func(ctx context.Context) (int, error) {
				return 0, nil
			})
		})
	})

	t.Run("return types do not match panics", func(t *testing.T) {
		assert.Panics(t, func() {
			di.PrepareFuncCtx[func(ctx context.Context) int](func(ctx context.Context) string {
				return ""
			})
		})
	})

	t.Run("more parameters on result function panics", func(t *testing.T) {
		assert.Panics(t, func() {
			di.PrepareFuncCtx[func(ctx context.Context, i int, s string) int](func(ctx context.Context) int {
				return 0
			})
		})
	})

	t.Run("parameters do not match panics", func(t *testing.T) {
		assert.Panics(t, func() {
			di.PrepareFuncCtx[func(ctx context.Context, i int) int](func(ctx context.Context, s string) int {
				return 0
			})
		})
	})

	t.Run("PrepareFunc with explicit provider", func(t *testing.T) {
		type Struct struct{ V int }
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{V: 5}
		})
		dp := di.GetDependencyProvider(ctx)

		fn := di.PrepareFunc[func(ctx context.Context, i int) int](dp, func(ctx context.Context, i int, s *Struct) int {
			return i + s.V
		})

		result := fn(ctx, 1)

		assert.Equal(t, 6, result)
	})
}
