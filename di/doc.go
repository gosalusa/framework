// Package di provides a dependency injection container. Dependencies are
// registered on a DependencyProvider and resolved either directly through the
// Resolve function or implicitly by filling a struct whose fields carry an
// inject struct tag.
//
// # Registering dependencies
//
// Dependencies are registered with one of the Register functions on a
// context.Context. The context determines which DependencyProvider receives
// the registration:
//
//	ctx := di.TestDependencyProviderContext()
//	di.RegisterSingleton(ctx, func() *Config {
//		return loadConfig()
//	})
//
// # Resolving dependencies
//
// A dependency can be resolved explicitly with Resolve:
//
//	cfg, err := di.Resolve[*Config](ctx)
//
// or implicitly by filling a struct with inject tags:
//
//	type handler struct {
//		Config *Config   `inject:""`
//		DB     *sql.DB   `inject:"db,optional"`
//		Cache  *redis.Client
//	}
//
//	filled, err := di.Fill(ctx, &handler{})
//
// The tag's first value names the dependency and is passed to the factory,
// which lets a single factory serve dependencies of the same type under
// different names. The remaining values are flags; optional allows a field to
// be left as its zero value when the dependency is not registered.
//
// # Lifecycles
//
// Register adds a factory that runs on every resolve. RegisterSingleton and
// RegisterLazySingleton register a value that is created once and reused; the
// lazy variants defer construction until the first resolve.
package di
