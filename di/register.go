package di

import (
	"context"
	"reflect"
)

// RegisterFactory registers a factory with the dependency provider carried by
// ctx.
func RegisterFactory(ctx context.Context, factory Factory) {
	GetDependencyProvider(ctx).RegisterFactory(factory)
}

// Register registers a factory for T with the dependency provider carried by
// ctx. The factory is called on every resolve and receives the name of the
// dependency being resolved.
func Register[T any](ctx context.Context, factory func(ctx context.Context, tag string) (T, error)) {
	GetDependencyProvider(ctx).Register(factory)
}

// RegisterWith registers a factory for T that is filled with its dependencies
// before each build. The factory receives a W whose inject fields have been
// populated from the dependency provider carried by ctx.
func RegisterWith[T, W any](ctx context.Context, factory func(ctx context.Context, tag string, with W) (T, error)) {
	GetDependencyProvider(ctx).RegisterWith(factory)
}

// RegisterSingleton registers a value of type T that is built once at
// registration time and returned on every resolve.
func RegisterSingleton[T any](ctx context.Context, factory func() T) {
	GetDependencyProvider(ctx).RegisterSingleton(factory)
}

// RegisterLazySingleton registers a factory for T that is called at most once,
// on the first resolve. Subsequent resolves return the same value.
func RegisterLazySingleton[T any](ctx context.Context, factory func() (T, error)) {
	GetDependencyProvider(ctx).RegisterLazySingleton(factory)
}

// RegisterLazySingletonWith registers a factory for T that is filled with its
// dependencies and called at most once, on the first resolve.
func RegisterLazySingletonWith[T, W any](ctx context.Context, factory func(with W) (T, error)) {
	GetDependencyProvider(ctx).RegisterLazySingletonWith(factory)
}

// RegisterValue registers a factory for the type t that builds a
// reflect.Value rather than a concrete Go type. It is primarily useful when
// the dependency type is only known dynamically.
func RegisterValue(ctx context.Context, t reflect.Type, factory func(ctx context.Context, tag string) (reflect.Value, error)) {
	GetDependencyProvider(ctx).RegisterValue(t, factory)
}

// RegisterFactory registers a factory with the provider, replacing any existing
// factory for the same type.
func (d *DependencyProvider) RegisterFactory(factory Factory) {
	d.factories.Set(factory.Type(), factory)
}

// Register registers a factory for T  the provider. The factory is called on
// every resolve and receives the name of the dependency being resolved.
func (d *DependencyProvider) Register[T any](factory func(ctx context.Context, tag string) (T, error)) {
	d.RegisterFactory(NewFactoryFunc(factory))
}

// RegisterWith registers a factory for T that is filled with its dependencies
// before each build. The factory receives a W whose inject fields have been
// populated from the dependency provider.
func (d *DependencyProvider) RegisterWith[T, W any](factory func(ctx context.Context, tag string, with W) (T, error)) {
	d.RegisterFactory(NewFactoryFuncWith(factory))
}

// RegisterSingleton registers a value of type T that is built once at
// registration time and returned on every resolve.
func (d *DependencyProvider) RegisterSingleton[T any](factory func() T) {
	d.RegisterFactory(NewSingletonFactory(factory()))
}

// RegisterLazySingleton registers a factory for T that is called at most once,
// on the first resolve. Subsequent resolves return the same value.
func (d *DependencyProvider) RegisterLazySingleton[T any](factory func() (T, error)) {
	d.RegisterFactory(NewLazySingletonFactory(factory))
}

// RegisterLazySingletonWith registers a factory for T that is filled with its
// dependencies and called at most once, on the first resolve.
func (d *DependencyProvider) RegisterLazySingletonWith[T, W any](factory func(with W) (T, error)) {
	d.RegisterFactory(NewLazySingletonWithFactory(factory))
}

// RegisterValue registers a factory for the type t that builds a
// reflect.Value rather than a concrete Go type. It is primarily useful when
// the dependency type is only known dynamically.
func (d *DependencyProvider) RegisterValue(t reflect.Type, factory func(ctx context.Context, tag string) (reflect.Value, error)) {
	d.RegisterFactory(NewValueFactory(t, factory))
}
