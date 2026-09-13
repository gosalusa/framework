package di_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
)

func TestFill(t *testing.T) {
	t.Run("fill", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable struct {
			WithTag *Struct `inject:""`
			NoTag   *Struct
		}

		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{}
		})

		f := &Fillable{}
		err := di.Fill(ctx, f)
		assert.NoError(t, err)
		assert.NotNil(t, f.WithTag)
		assert.Nil(t, f.NoTag)
	})

	t.Run("deep", func(t *testing.T) {
		type Struct struct{ V int }
		type FillableB struct {
			Struct *Struct `inject:""`
		}
		type FillableA struct {
			B *FillableB `inject:""`
		}
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{}
		})

		f := &FillableA{}
		err := di.Fill(ctx, f)
		assert.NoError(t, err)
		assert.NotNil(t, f.B)
		assert.NotNil(t, f.B.Struct)
	})

	t.Run("tag", func(t *testing.T) {
		type Struct struct {
			Value string
		}
		type Fillable struct {
			WithTag *Struct `inject:"with tag"`
			Empty   *Struct `inject:""`
		}
		ctx := di.TestDependencyProviderContext()
		di.Register(ctx, func(ctx context.Context, tag string) (*Struct, error) {
			return &Struct{
				Value: tag,
			}, nil
		})

		f := &Fillable{}
		err := di.Fill(ctx, f)
		assert.NoError(t, err)
		assert.Equal(t, "with tag", f.WithTag.Value)
		assert.NotNil(t, "", f.Empty.Value)
	})

	t.Run("not registered", func(t *testing.T) {
		type Fillable struct {
			Miss *int `inject:""`
		}

		ctx := di.TestDependencyProviderContext()
		f := &Fillable{}
		err := di.Fill(ctx, f)
		assert.ErrorIs(t, err, di.ErrNotRegistered)
	})

	t.Run("non pointer", func(t *testing.T) {
		type Fillable struct {
		}

		ctx := di.TestDependencyProviderContext()
		f := Fillable{}
		err := di.Fill(ctx, f)
		assert.ErrorIs(t, err, di.ErrFillParameters)
	})

	t.Run("not struct", func(t *testing.T) {
		type Fillable int

		ctx := di.TestDependencyProviderContext()
		di.Register(ctx, func(ctx context.Context, tag string) (Fillable, error) {
			return 7, nil
		})
		var f Fillable
		err := di.Fill(ctx, &f)
		assert.NoError(t, err)
		assert.Equal(t, Fillable(7), f)
	})

	t.Run("resolve error", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable struct {
			S *Struct `inject:""`
		}

		ctx := di.TestDependencyProviderContext()
		resolveErr := fmt.Errorf("error in register")
		di.Register(ctx, func(ctx context.Context, tag string) (*Struct, error) {
			return nil, resolveErr
		})

		f := &Fillable{}
		err := di.Fill(ctx, f)

		assert.Error(t, err)
		assert.ErrorIs(t, err, resolveErr)
	})

	t.Run("not fill unfillable type", func(t *testing.T) {
		type Struct struct{ V int }
		type IsFillable struct {
			S *Struct `inject:""`
		}
		type NotFillable struct {
			S *Struct
		}
		type FillableRoot struct {
			IsFillable  *IsFillable  `inject:""`
			NotFillable *NotFillable `inject:""`
		}

		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{}
		})
		f := &FillableRoot{}
		err := di.Fill(ctx, f)
		assert.ErrorIs(t, err, di.ErrNotRegistered)
	})

	t.Run("deep", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable2 struct {
			S *Struct `inject:""`
		}
		type Fillable1 struct {
			F *Fillable2 `inject:""`
		}
		type FillableRoot struct {
			F *Fillable1 `inject:""`
		}

		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{}
		})
		f := &FillableRoot{}
		err := di.Fill(ctx, f)
		assert.NoError(t, err)
		assert.NotNil(t, f.F.F.S)
	})

	t.Run("deep tag", func(t *testing.T) {
		type Struct struct{ V int }
		type Fillable2 struct {
			S *Struct `inject:""`
		}
		type Fillable1 struct {
			F *Fillable2 `inject:""`
		}
		type FillableRoot struct {
			F *Fillable1 `inject:""`
		}

		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *Struct {
			return &Struct{}
		})
		f := &FillableRoot{}
		err := di.Fill(ctx, f)
		assert.NoError(t, err)
		assert.NotNil(t, f.F.F.S)
	})

	t.Run("optional", func(t *testing.T) {
		type Struct struct{ V int }
		type FillableRoot struct {
			S *Struct `inject:",optional"`
		}

		ctx := di.TestDependencyProviderContext()
		f := &FillableRoot{}
		err := di.Fill(ctx, f)
		assert.NoError(t, err)
		assert.Nil(t, f.S)
	})

	t.Run("resolve", func(t *testing.T) {
		type Struct struct{ V int }

		ctx := di.TestDependencyProviderContext()

		expected := &Struct{}
		di.RegisterSingleton(ctx, func() *Struct {
			return expected
		})

		var s *Struct
		err := di.Fill(ctx, &s)
		assert.NoError(t, err)
		assert.Equal(t, expected, s)
	})

	t.Run("interface direct", func(t *testing.T) {
		type Interface interface{ Foo() }

		ctx := di.TestDependencyProviderContext()

		var i Interface
		err := di.Fill(ctx, i)
		assert.Error(t, err)
	})

	t.Run("context", func(t *testing.T) {
		expectedCtx := di.TestDependencyProviderContext()

		var ctx context.Context
		err := di.Fill(expectedCtx, &ctx)
		assert.NoError(t, err)
		assert.Same(t, expectedCtx, ctx)
	})

	t.Run("context no provider", func(t *testing.T) {
		dp := di.NewDependencyProvider()

		backgroundCtx := context.Background()

		var ctx context.Context
		err := dp.Fill(backgroundCtx, &ctx)
		assert.NoError(t, err)

		assert.NotSame(t, &backgroundCtx, &ctx)

		ctxDP := di.GetDependencyProvider(ctx)
		assert.Same(t, dp, ctxDP)
	})

	t.Run("context with provider", func(t *testing.T) {
		dp := di.NewDependencyProvider()
		baseCtx := di.ContextWithDependencyProvider(context.Background(), dp)

		var ctx context.Context
		err := dp.Fill(baseCtx, &ctx)
		assert.NoError(t, err)

		assert.Same(t, baseCtx, ctx)

		ctxDP := di.GetDependencyProvider(ctx)
		assert.Same(t, dp, ctxDP)
	})

}
