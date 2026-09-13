package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"gosalusa.com/internal/helpers"
)

// Factory builds a value for a single dependency type each time the resolver
// needs it.
type Factory interface {
	// Build returns the dependency, constructing a fresh value or returning a
	// cached one depending on the factory's lifecycle.
	Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error)
	// Type returns the dependency type the factory builds.
	Type() reflect.Type
}

// Singleton is implemented by factories that cache their built value and can
// expose it without building it again.
type Singleton interface {
	// Peek returns the cached value, the error recorded while building it, and
	// whether it has been built yet.
	Peek() (any, error, bool)
}

// Dependant is implemented by factories that declare a value of type W whose
// inject fields must be resolved before the factory runs.
type Dependant interface {
	// DependsOn returns the types the factory requires to be present in the
	// provider.
	DependsOn() []reflect.Type
}

func dependancies(t reflect.Type) []reflect.Type {
	if !isFillable(t) {
		return []reflect.Type{t}
	}
	deps := []reflect.Type{}
	for _, sf := range helpers.GetFields(t.Elem()) {
		if _, ok := sf.Tag.Lookup("inject"); ok {
			deps = append(deps, sf.Type)
		}
	}
	return deps
}

// ===============================
// ||                           ||
// ||        FactoryFunc        ||
// ||                           ||
// ===============================

// FactoryFunc is a Factory backed by an ordinary function. Each call to Build
// invokes the function and returns a fresh value.
type FactoryFunc[T any] func(ctx context.Context, tag string) (T, error)

// NewFactoryFunc wraps f as a FactoryFunc that builds values of type T.
func NewFactoryFunc[T any](f func(ctx context.Context, tag string) (T, error)) FactoryFunc[T] {
	return FactoryFunc[T](f)
}

var _ Factory = (FactoryFunc[any])(nil)

// Build calls the wrapped function and returns its result.
func (f FactoryFunc[T]) Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error) {
	return f(ctx, tag)
}

// Type returns the type of T this factory builds.
func (f FactoryFunc[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

// ===============================
// ||                           ||
// ||      FactoryFuncWith      ||
// ||                           ||
// ===============================

// FactoryFuncWith is a Factory backed by a function that receives a value of
// type W whose inject fields are filled from the provider before each build.
type FactoryFuncWith[T, W any] func(ctx context.Context, tag string, with W) (T, error)

// NewFactoryFuncWith wraps f as a FactoryFuncWith that builds values of type T
// with a filled dependency value W.
func NewFactoryFuncWith[T, W any](f func(ctx context.Context, tag string, with W) (T, error)) FactoryFuncWith[T, W] {
	return FactoryFuncWith[T, W](f)
}

var _ Factory = (FactoryFuncWith[any, any])(nil)

// Build fills a W with the provided dependency values and passes it to the
// wrapped function.
func (f FactoryFuncWith[T, W]) Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error) {
	var with W
	err := dp.Fill(ctx, &with)
	if err != nil {
		var zero T
		return zero, err
	}
	return f(ctx, tag, with)
}

// Type returns the type of T this factory builds.
func (f FactoryFuncWith[T, W]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

// DependsOn returns the inject field types of W.
func (f FactoryFuncWith[T, W]) DependsOn() []reflect.Type {
	return dependancies(reflect.TypeFor[W]())
}

// ================================
// ||                            ||
// ||        ValueFactory        ||
// ||                            ||
// ================================

// ValueFactory is a Factory for a dependency type given as a reflect.Type. It
// is primarily used when the dependency type is only known dynamically.
type ValueFactory struct {
	factory func(ctx context.Context, tag string) (reflect.Value, error)
	typ     reflect.Type
}

// NewValueFactory wraps factory as a ValueFactory for the type typ.
func NewValueFactory(typ reflect.Type, factory func(ctx context.Context, tag string) (reflect.Value, error)) *ValueFactory {
	return &ValueFactory{
		factory: factory,
		typ:     typ,
	}
}

var _ Factory = (*ValueFactory)(nil)

// Build calls the wrapped factory and verifies the returned value's type
// matches the declared dependency type.
func (f *ValueFactory) Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error) {
	v, err := f.factory(ctx, tag)
	if err == nil && v.Type() != f.typ {
		return reflect.Zero(f.typ), fmt.Errorf("invalid type %v expected %v", v.Type(), f.typ)
	}
	return v.Interface(), err
}

// Type returns the dependency type this factory builds.
func (f *ValueFactory) Type() reflect.Type {
	return f.typ
}

// ================================
// ||                            ||
// ||      SingletonFactory      ||
// ||                            ||
// ================================

// SingletonFactory is a Factory that returns a fixed value on every build. It
// implements Singleton.
type SingletonFactory[T any] struct {
	value T
}

var _ Singleton = (*SingletonFactory[any])(nil)

// NewSingletonFactory wraps value as a SingletonFactory for the type of T.
func NewSingletonFactory[T any](value T) *SingletonFactory[T] {
	return &SingletonFactory[T]{value: value}
}

// Build returns the fixed value.
func (f *SingletonFactory[T]) Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error) {
	return f.value, nil
}

// Type returns the type of T this factory builds.
func (f *SingletonFactory[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

// Peek returns the fixed value.
func (f *SingletonFactory[T]) Peek() (any, error, bool) {
	return f.value, nil, true
}

// ================================
// ||                            ||
// ||    LazySingletonFactory    ||
// ||                            ||
// ================================

// LazySingletonFactory is a Factory that builds its value at most once, on the
// first build. It implements Singleton.
type LazySingletonFactory[T any] struct {
	once    sync.Once
	ready   bool
	factory func() (T, error)

	value T
	err   error
}

var _ Singleton = (*LazySingletonFactory[any])(nil)

// NewLazySingletonFactory wraps factory as a LazySingletonFactory for the type
// of T.
func NewLazySingletonFactory[T any](factory func() (T, error)) *LazySingletonFactory[T] {
	return &LazySingletonFactory[T]{
		once:    sync.Once{},
		factory: factory,
	}
}

// Build calls the wrapped factory once and returns the cached result on
// subsequent calls.
func (f *LazySingletonFactory[T]) Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error) {
	f.once.Do(f.load)
	return f.value, f.err
}

// Type returns the type of T this factory builds.
func (f *LazySingletonFactory[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (f *LazySingletonFactory[T]) load() {
	f.value, f.err = f.factory()
	f.ready = true
}

// Peek returns the cached value, the error recorded while building it, and
// whether it has been built yet.
func (f *LazySingletonFactory[T]) Peek() (any, error, bool) {
	return f.value, f.err, f.ready
}

// ================================
// ||                            ||
// ||  LazySingletonWithFactory  ||
// ||                            ||
// ================================

// LazySingletonWithFactory is a Factory that fills a W with dependencies and
// builds its value at most once, on the first build. It implements Singleton
// and Dependant.
type LazySingletonWithFactory[T, W any] struct {
	done    atomic.Uint32
	m       sync.Mutex
	factory func(with W) (T, error)

	value T
	err   error
}

var _ Singleton = (*LazySingletonWithFactory[any, any])(nil)

// NewLazySingletonWithFactory wraps factory as a LazySingletonWithFactory for
// the type of T.
func NewLazySingletonWithFactory[T, W any](factory func(with W) (T, error)) *LazySingletonWithFactory[T, W] {
	return &LazySingletonWithFactory[T, W]{
		factory: factory,
	}
}

// Build fills a W with dependencies, then calls the wrapped factory; the
// result is cached and returned on subsequent calls.
func (f *LazySingletonWithFactory[T, W]) Build(ctx context.Context, dp *DependencyProvider, tag string) (any, error) {
	// implanted matching sync.Once to allow for inlining the fast path
	if f.done.Load() == 0 {
		f.load(ctx, dp)
	}
	return f.value, f.err
}

// Type returns the type of T this factory builds.
func (f *LazySingletonWithFactory[T, W]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (f *LazySingletonWithFactory[T, W]) load(ctx context.Context, dp *DependencyProvider) {
	f.m.Lock()
	defer f.m.Unlock()
	if f.done.Load() == 0 {
		defer f.done.Store(1)

		var with W
		err := dp.Fill(ctx, &with)
		if err != nil {
			f.err = err
			return
		}
		f.value, f.err = f.factory(with)
	}
}

// Peek returns the cached value, the error recorded while building it, and
// whether it has been built yet.
func (f *LazySingletonWithFactory[T, W]) Peek() (any, error, bool) {
	return f.value, f.err, f.done.Load() != 0
}

// DependsOn returns the inject field types of W.
func (f *LazySingletonWithFactory[T, W]) DependsOn() []reflect.Type {
	return dependancies(reflect.TypeFor[W]())
}
