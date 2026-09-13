package kernel

import (
	"context"
	"net/http"

	"github.com/go-openapi/spec"
	"gosalusa.com/di"
	"gosalusa.com/openapidoc"
	"gosalusa.com/router"
	"gosalusa.com/salusaconfig"
)

type KernelOption func(*Kernel) *Kernel

func Bootstrap(bootstrap ...func(context.Context) error) KernelOption {
	return func(k *Kernel) *Kernel {
		k.bootstrap = bootstrap
		return k
	}
}

func Config[T salusaconfig.Config](cb func() T) KernelOption {
	return func(k *Kernel) *Kernel {
		cfg := cb()
		k.cfg = cfg
		k.registerConfig = func(ctx context.Context) {
			di.RegisterSingleton(ctx, func() T {
				return cfg
			})
			di.RegisterSingleton(ctx, func() salusaconfig.Config {
				return cfg
			})
		}
		return k
	}
}

func RootHandler(rootHandler func(ctx context.Context) http.Handler) KernelOption {
	return func(k *Kernel) *Kernel {
		k.rootHandlerFactory = rootHandler
		return k
	}
}
func InitRoutes(cb func(r *router.Router)) KernelOption {
	return RootHandler(func(ctx context.Context) http.Handler {
		r := router.New()
		cb(r)
		r.Register(ctx)
		return r
	})
}
func Middleware(globalMiddleware []router.Middleware) KernelOption {
	return func(k *Kernel) *Kernel {
		k.globalMiddleware = globalMiddleware
		return k
	}
}

func Services(services ...Service) KernelOption {
	return func(k *Kernel) *Kernel {
		k.services = services
		return k
	}
}

func APIDocumentation(options ...openapidoc.SwaggerOption) KernelOption {
	return func(k *Kernel) *Kernel {
		k.docs = &spec.Swagger{}
		for _, opt := range options {
			k.docs = opt(k.docs)
		}
		return k
	}
}
