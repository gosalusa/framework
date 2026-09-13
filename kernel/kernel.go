package kernel

import (
	"context"
	"net/http"

	"github.com/go-openapi/spec"
	"gosalusa.com/di"
	"gosalusa.com/request"
	"gosalusa.com/router"
	"gosalusa.com/salusaconfig"
)

type Kernel struct {
	bootstrap          []func(context.Context) error
	registerConfig     func(context.Context)
	rootHandlerFactory func(ctx context.Context) http.Handler
	rootHandler        http.Handler
	services           []Service
	dependencyProvider *di.DependencyProvider

	globalMiddleware []router.Middleware

	docs      *spec.Swagger
	fetchAuth func(ctx context.Context, username string, r *http.Request) error

	cfg salusaconfig.Config

	bootstrapped bool
}

func New(options ...KernelOption) *Kernel {
	k := &Kernel{
		bootstrap:      []func(context.Context) error{},
		registerConfig: func(context.Context) {},
		rootHandler:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		globalMiddleware: []router.Middleware{
			request.DIMiddleware(),
		},
		services:     []Service{},
		bootstrapped: false,
	}

	for _, o := range options {
		k = o(k)
	}

	return k
}

func (k *Kernel) RootHandler() http.Handler {
	if !k.bootstrapped {
		panic("cannot access root handler before the kernel is bootstrapped")
	}
	return k.rootHandler
}

func (k *Kernel) Config() salusaconfig.Config {
	return k.cfg
}
