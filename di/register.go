package di

import (
	"context"
	"reflect"
)

// RegisterFactory registers a factory with the dependency provider carried by
// ctx.
func RegisterFactory(ctx context.Context, factory Factory) {
	dp := GetDependencyProvider(ctx)
	dp.Register(factory)
}

// Register registers a factory for T with the dependency provider carried by
// ctx. The factory is called on every resolve and receives the name of the
// dependency being resolved.
func Register[T any](ctx context.Context, factory func(ctx context.Context, tag string) (T, error)) {
	RegisterFactory(ctx, NewFactoryFunc(factory))
}

// RegisterWith registers a factory for T that is filled with its dependencies
// before each build. The factory receives a W whose inject fields have been
// populated from the dependency provider carried by ctx.
func RegisterWith[T, W any](ctx context.Context, factory func(ctx context.Context, tag string, with W) (T, error)) {
	RegisterFactory(ctx, NewFactoryFuncWith(factory))
}

// RegisterSingleton registers a value of type T that is built once at
// registration time and returned on every resolve.
func RegisterSingleton[T any](ctx context.Context, factory func() T) {
	RegisterFactory(ctx, NewSingletonFactory(factory()))
}

// RegisterLazySingleton registers a factory for T that is called at most once,
// on the first resolve. Subsequent resolves return the same value.
func RegisterLazySingleton[T any](ctx context.Context, factory func() (T, error)) {
	RegisterFactory(ctx, NewLazySingletonFactory(func() (T, error) {
		return factory()
	}))
}

// RegisterLazySingletonWith registers a factory for T that is filled with its
// dependencies and called at most once, on the first resolve.
func RegisterLazySingletonWith[T, W any](ctx context.Context, factory func(with W) (T, error)) {
	RegisterFactory(ctx, NewLazySingletonWithFactory(factory))
}

// RegisterValue registers a factory for the type t that builds a
// reflect.Value rather than a concrete Go type. It is primarily useful when
// the dependency type is only known dynamically.
func RegisterValue(ctx context.Context, t reflect.Type, factory func(ctx context.Context, tag string) (reflect.Value, error)) {
	RegisterFactory(ctx, NewValueFactory(t, factory))
}

// Register registers a factory with the provider, replacing any existing
// factory for the same type.
func (d *DependencyProvider) Register(factory Factory) {
	d.factories.Set(factory.Type(), factory)
}
