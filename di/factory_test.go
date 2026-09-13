package di_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
)

func TestFactories(t *testing.T) {
	type Struct struct{ V int }
	type Dependant struct{ DependsOn *Struct }

	structType := reflect.TypeFor[*Struct]()

	type FactoryFactory struct {
		Error   func() di.Factory
		Default func() di.Factory
		Eger    bool
	}

	ctx := context.Background()
	dp := di.NewDependencyProvider()

	expectedError := fmt.Errorf("expected error")

	registeredStruct := &Struct{}
	dp.Register(di.NewSingletonFactory(registeredStruct))

	factories := []FactoryFactory{
		{
			Default: func() di.Factory {
				return di.NewFactoryFunc(func(ctx context.Context, tag string) (*Struct, error) {
					return &Struct{}, nil
				})
			},
			Error: func() di.Factory {
				return di.NewFactoryFunc(func(ctx context.Context, tag string) (*Struct, error) {
					return nil, expectedError
				})
			},
		},
		{
			Default: func() di.Factory {
				return di.NewFactoryFuncWith(func(ctx context.Context, tag string, with *Struct) (*Dependant, error) {
					return &Dependant{DependsOn: with}, nil
				})
			},
			Error: func() di.Factory {
				return di.NewFactoryFuncWith(func(ctx context.Context, tag string, with *Struct) (*Dependant, error) {
					return nil, expectedError
				})
			},
		},
		{
			Default: func() di.Factory {
				return di.NewValueFactory(structType, func(ctx context.Context, tag string) (reflect.Value, error) {
					return reflect.ValueOf(&Struct{}), nil
				})
			},
			Error: func() di.Factory {
				return di.NewValueFactory(structType, func(ctx context.Context, tag string) (reflect.Value, error) {
					return reflect.Zero(structType), expectedError
				})
			},
		},
		{
			Default: func() di.Factory {
				return di.NewSingletonFactory(&Struct{})
			},
			Eger: true,
		},
		{
			Default: func() di.Factory {
				return di.NewLazySingletonFactory(func() (*Struct, error) {
					return &Struct{}, nil
				})
			},
			Error: func() di.Factory {
				return di.NewLazySingletonFactory(func() (*Struct, error) {
					return nil, expectedError
				})
			},
		},
		{
			Default: func() di.Factory {
				return di.NewLazySingletonWithFactory(func(with *Struct) (*Dependant, error) {
					return &Dependant{DependsOn: with}, nil
				})
			},
			Error: func() di.Factory {
				return di.NewLazySingletonWithFactory(func(with *Struct) (*Dependant, error) {
					return nil, expectedError
				})
			},
		},
	}

	for _, ff := range factories {
		t.Run(reflect.TypeOf(ff.Default).Name(), func(t *testing.T) {

			t.Run("success", func(t *testing.T) {
				factory := ff.Default()
				c, err := factory.Build(ctx, dp, "")
				assert.NoError(t, err)
				assert.NotZero(t, c)
			})

			t.Run("type", func(t *testing.T) {
				factory := ff.Default()
				c, err := factory.Build(ctx, dp, "")
				assert.NoError(t, err)
				assert.Same(t, factory.Type(), reflect.TypeOf(c))
			})

			if ff.Error != nil {
				t.Run("error", func(t *testing.T) {
					factory := ff.Error()
					c, err := factory.Build(ctx, dp, "")
					assert.Same(t, expectedError, err)
					assert.Zero(t, c)
				})
			}

			if _, ok := ff.Default().(di.Singleton); ok {
				t.Run("is singleton", func(t *testing.T) {
					factory := ff.Default()
					c1, err := factory.Build(ctx, dp, "")
					assert.NoError(t, err)

					c2, err := factory.Build(ctx, dp, "")
					assert.NoError(t, err)

					assert.Same(t, c1, c2)
				})
				t.Run("peek", func(t *testing.T) {
					factory := ff.Default()
					singleton := factory.(di.Singleton)

					c1, err, ready := singleton.Peek()
					if ff.Eger {
						assert.NotZero(t, c1)
						assert.NoError(t, err)
						assert.True(t, ready)
					} else {
						assert.Zero(t, c1)
						assert.NoError(t, err)
						assert.False(t, ready)
					}

					c2, err := factory.Build(ctx, dp, "")
					assert.NoError(t, err)
					assert.NotZero(t, c2)

					c3, err, ready := singleton.Peek()
					assert.NotZero(t, c3)
					assert.NoError(t, err)
					assert.True(t, ready)

					assert.Same(t, c2, c3)
				})
			}

			if _, ok := ff.Default().(di.Dependant); ok {
				t.Run("with", func(t *testing.T) {
					factory := ff.Default()
					c, err := factory.Build(ctx, dp, "")
					assert.NoError(t, err)
					assert.Equal(t, registeredStruct, c.(*Dependant).DependsOn)
				})
				t.Run("missing with", func(t *testing.T) {
					factory := ff.Default()
					ctx := context.Background()
					dp := di.NewDependencyProvider()

					c, err := factory.Build(ctx, dp, "")
					assert.Error(t, err)
					assert.Zero(t, c)
				})
				t.Run("depends on", func(t *testing.T) {
					factory := ff.Default().(di.Dependant)
					assert.Equal(t, []reflect.Type{structType}, factory.DependsOn())
				})
			}
		})
	}
}
