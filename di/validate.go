package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/dominikbraun/graph"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/validate"
)

var (
	// ErrDependancyCycle is returned when a set of factories depends on
	// itself, directly or transitively.
	ErrDependancyCycle = errors.New("dependancy cycle")
	// ErrMissingDependancy is returned when a factory or inject field requires
	// a dependency that is not registered.
	ErrMissingDependancy = errors.New("missing dependancy")
)

// Uses embeds into a struct to declare and validate dependencies of type T.
// When embedded, it allows the struct to be validated to ensure all dependencies
// that T requires via `inject` fields are registered in the DI container.
//
// Example:
//
//	type Foo struct {
//		Logger *slog.Logger `inject:""`
//	}
//	type Bar struct {
//		di.Uses[*Foo]
//	}
type Uses[T any] struct{}

func (v *Uses[T]) Validate(ctx context.Context) error {
	return ValidatorFor[T](ctx).Validate(ctx)
}

// DIValidator checks that every inject field on a struct has a registered
// factory. It implements validate.Validator.
type DIValidator struct {
	dp  *DependencyProvider
	typ reflect.Type
}

var _ validate.Validator = (*DIValidator)(nil)

// Validate returns an error joining ErrMissingDependancy for every inject
// field of the validator's struct whose type has no registered factory.
func (v *DIValidator) Validate(ctx context.Context) error {
	errs := []error{}
	t := v.typ
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	for _, sf := range helpers.GetFields(t) {
		if !sf.IsExported() {
			continue
		}

		_, ok := sf.Tag.Lookup("inject")
		if !ok {
			continue
		}

		switch sf.Type {
		case contextType, dependencyProviderType:
			continue
		}

		_, ok = v.dp.factories.Get(sf.Type)
		if !ok {
			errs = append(errs, fmt.Errorf("%w %s on %s.%s", ErrMissingDependancy, sf.Type, v.typ, sf.Name))
		}
	}

	return errors.Join(errs...)
}

// Validator returns a DIValidator for rootType using the dependency provider
// carried by ctx. Use GetDependencyProvider(ctx) to retrieve the provider.
func Validator(ctx context.Context, rootType reflect.Type) *DIValidator {
	return GetDependencyProvider(ctx).Validator(rootType)
}

// ValidatorFor returns a DIValidator for type T using the dependency provider
// carried by ctx. Use GetDependencyProvider(ctx) to retrieve the provider.
func ValidatorFor[T any](ctx context.Context) *DIValidator {
	return Validator(ctx, reflect.TypeFor[T]())
}

// Validator returns a DIValidator for rootType validated against this
// DependencyProvider.
func (dp *DependencyProvider) Validator(rootType reflect.Type) *DIValidator {
	return &DIValidator{
		dp:  dp,
		typ: rootType,
	}
}

// Validate checks the registered factories for dependency cycles. A missing
// dependency in a Dependant factory is reported as an error wrapping
// ErrMissingDependancy, and a cycle as an error wrapping ErrDependancyCycle.
func (dp *DependencyProvider) Validate(ctx context.Context) error {
	errs := []error{}

	if err := dp.validateCycles(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func typeHash(t reflect.Type) reflect.Type {
	return t
}
func (dp *DependencyProvider) validateCycles() error {
	g := graph.New(typeHash, graph.Directed(), graph.PreventCycles())

	var err error
	for typ := range dp.factories.All() {
		err = g.AddVertex(typ)
		if err != nil {
			panic(err)
		}
	}
	for typ, factory := range dp.factories.All() {
		depends, ok := factory.(Dependant)
		if !ok {
			continue
		}
		deps := depends.DependsOn()
		for _, dep := range deps {
			err = g.AddEdge(typ, dep)
			if errors.Is(err, graph.ErrEdgeCreatesCycle) {
				return newCycleError(g, dep, typ)
			} else if errors.Is(err, graph.ErrVertexNotFound) {
				return fmt.Errorf("%w %s in %s factory", ErrMissingDependancy, dep, typ)
			} else if err != nil {
				panic(err)
			}
		}
	}
	return nil
}

func newCycleError[K comparable, T any](g graph.Graph[K, T], source, target K) error {
	path, err := graph.ShortestPath(g, source, target)
	if err != nil {
		panic(err)
	}

	strPath := []byte{}
	for _, t := range path {
		strPath = fmt.Appendf(strPath, "%v -> ", t)
	}
	strPath = fmt.Appendf(strPath, "%v", path[0])

	return fmt.Errorf("%w: %s", ErrDependancyCycle, strPath)

}
