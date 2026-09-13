package openapidocdi

import (
	"context"

	"gosalusa.com/di"
	"gosalusa.com/kernel"
	"gosalusa.com/openapidoc"
)

type apiDocerOpts struct {
	Kernel *kernel.Kernel `inject:""`
}

func Register(ctx context.Context) {
	di.RegisterLazySingletonWith(ctx, func(opts *apiDocerOpts) (openapidoc.APIDocer, error) {
		return opts.Kernel, nil
	})
}
