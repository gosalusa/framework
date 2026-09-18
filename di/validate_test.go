package di_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
)

func TestDependencyProvider_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		type WithFloat struct {
			Float float64 `inject:""`
		}
		type WithInt struct {
			Int int `inject:""`
		}
		type WithString struct {
			String string `inject:""`
		}
		dp := di.NewDependencyProvider()
		dp.RegisterFactory(&di.SingletonFactory[float64]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[string, *WithFloat]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[int, *WithString]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[uint, *WithInt]{})

		ctx := context.Background()

		err := dp.Validate(ctx)
		assert.NoError(t, err)
	})

	t.Run("missing dependencies", func(t *testing.T) {
		type WithFloat struct {
			Float float64 `inject:""`
		}
		dp := di.NewDependencyProvider()
		dp.RegisterFactory(&di.LazySingletonWithFactory[string, *WithFloat]{})

		ctx := context.Background()

		err := dp.Validate(ctx)
		assert.ErrorIs(t, err, di.ErrMissingDependancy)
	})

	t.Run("cycle", func(t *testing.T) {
		dp := di.NewDependencyProvider()
		dp.RegisterFactory(&di.LazySingletonWithFactory[int, float64]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[float64, string]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[string, int]{})

		ctx := context.Background()

		err := dp.Validate(ctx)
		assert.ErrorIs(t, err, di.ErrDependancyCycle)
	})

	t.Run("cycle fill", func(t *testing.T) {
		type WithFloat struct {
			Float float64 `inject:""`
		}
		type WithInt struct {
			Int int `inject:""`
		}
		type WithString struct {
			String string `inject:""`
		}
		dp := di.NewDependencyProvider()
		dp.RegisterFactory(&di.LazySingletonWithFactory[int, *WithFloat]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[float64, *WithString]{})
		dp.RegisterFactory(&di.LazySingletonWithFactory[string, *WithInt]{})

		ctx := context.Background()

		err := dp.Validate(ctx)
		assert.ErrorIs(t, err, di.ErrDependancyCycle)
	})
}

func TestValidator(t *testing.T) {
	type Dep struct{ V int }
	type Fillable struct {
		D       *Dep            `inject:""`
		Context context.Context `inject:""`
	}

	t.Run("valid", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Dep {
			return &Dep{}
		})

		v := di.Validator(ctx, reflect.TypeOf(&Fillable{}))
		err := v.Validate(ctx)
		assert.NoError(t, err)
	})

	t.Run("missing dependency", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()

		v := di.Validator(ctx, reflect.TypeOf(&Fillable{}))
		err := v.Validate(ctx)
		assert.ErrorIs(t, err, di.ErrMissingDependancy)
	})

	t.Run("non struct", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		type NotStruct int

		v := di.Validator(ctx, reflect.TypeOf(NotStruct(0)))
		err := v.Validate(ctx)
		assert.NoError(t, err)
	})

	t.Run("pointer to non struct", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		type NotStruct int

		v := di.Validator(ctx, reflect.TypeOf(new(NotStruct)))
		err := v.Validate(ctx)
		assert.NoError(t, err)
	})
}
