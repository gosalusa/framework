package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"gosalusa.com/extra/maps"
)

type contextKey uint8

const (
	dpKey contextKey = iota
)

// DependencyProvider is a registry of the factories that build and cache
// dependencies. It is attached to a context.Context and used to resolve and
// fill dependencies from it.
type DependencyProvider struct {
	factories maps.Map[reflect.Type, Factory]
}

var (
	// ErrNotRegistered is returned when a dependency has no registered factory.
	ErrNotRegistered = errors.New("dependency not registered")
	// ErrFillParameters is returned when Fill is called with a value that is
	// not a non-nil pointer.
	ErrFillParameters = errors.New("invalid fill parameters")
	// ErrDependencyProviderNotInContext is returned when a dependency that
	// requires a DependencyProvider is resolved from a context that does not
	// carry one.
	ErrDependencyProviderNotInContext = errors.New("DependencyProvider not in context")
)

var (
	contextType            = reflect.TypeFor[context.Context]()
	dependencyProviderType = reflect.TypeFor[*DependencyProvider]()
)

func errNotRegistered(t reflect.Type) error {
	return fmt.Errorf("%w: %s", ErrNotRegistered, t)
}

var defaultProvider = NewDependencyProvider()

// NewDependencyProvider returns a new DependencyProvider that can resolve
// itself and the surrounding context.Context. Any dependencies registered on
// the returned provider are isolated from every other provider.
func NewDependencyProvider() *DependencyProvider {
	dp := &DependencyProvider{
		factories: &maps.Sync[reflect.Type, Factory]{},
	}
	dp.Register(NewFactoryFunc(func(ctx context.Context, tag string) (context.Context, error) {
		return ctx, nil
	}))
	dp.Register(NewFactoryFunc(func(ctx context.Context, tag string) (*DependencyProvider, error) {
		return dp, nil
	}))
	return dp
}

// ContextWithDependencyProvider returns a copy of ctx that carries dp as its
// dependency provider.
func ContextWithDependencyProvider(ctx context.Context, dp *DependencyProvider) context.Context {
	return context.WithValue(ctx, dpKey, dp)
}

// GetDependencyProvider returns the DependencyProvider stored on ctx, or the
// shared default provider when ctx does not carry one.
func GetDependencyProvider(ctx context.Context) *DependencyProvider {
	v := ctx.Value(dpKey)
	if v == nil {
		return defaultProvider
	}
	dp, ok := v.(*DependencyProvider)
	if !ok {
		return defaultProvider
	}
	return dp
}

// TestDependencyProviderContext returns a background context carrying a fresh
// DependencyProvider for use in tests.
func TestDependencyProviderContext() context.Context {
	return ContextWithDependencyProvider(
		context.Background(),
		NewDependencyProvider(),
	)
}

// Singletons returns the registered singleton factories in no particular
// order.
func (dp *DependencyProvider) Singletons() []Singleton {
	singletons := []Singleton{}

	for _, f := range dp.factories.All() {
		if s, ok := f.(Singleton); ok {
			singletons = append(singletons, s)
		}
	}
	return singletons
}

// Singletons returns the singleton factories registered on the provider
// carried by ctx, in no particular order.
func Singletons(ctx context.Context) []Singleton {
	dp := GetDependencyProvider(ctx)
	return dp.Singletons()
}
