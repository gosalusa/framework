package kernel

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"gosalusa.com/di"
	"gosalusa.com/salusaconfig"
)

var (
	ErrAlreadyBootstrapped = errors.New("kernel already bootstrapped")
)

func (k *Kernel) Bootstrap(ctx context.Context) error {
	var err error

	if k.bootstrapped {
		return ErrAlreadyBootstrapped
	}
	k.bootstrapped = true
	k.dependencyProvider = di.GetDependencyProvider(ctx)

	k.rootHandler = k.rootHandlerFactory(ctx)

	for _, service := range k.services {
		di.RegisterValue(ctx, reflect.TypeOf(service), func(ctx context.Context, tag string) (reflect.Value, error) {
			return reflect.ValueOf(service), nil
		})
	}

	di.RegisterSingleton(ctx, func() *Kernel {
		return k
	})

	k.registerConfig(ctx)

	for i, b := range k.bootstrap {
		err = b(ctx)
		if err != nil {
			return fmt.Errorf("Kernel.Bootstrap: step %d: %w", i, err)
		}
	}

	return nil
}

func Register[T salusaconfig.Config](fn func(ctx context.Context, c T)) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		cfg, err := di.Resolve[T](ctx)
		if err != nil {
			return err
		}
		fn(ctx, cfg)
		return nil
	}
}
