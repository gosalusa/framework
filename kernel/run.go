package kernel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"

	"github.com/spf13/pflag"
	"gosalusa.com/clog"
	"gosalusa.com/di"
)

func (k *Kernel) Run(ctx context.Context) error {
	k.singles(ctx)
	defer func() {
		k.closeAndLog(ctx)
	}()

	go func() {
		<-ctx.Done()
		k.closeAndLog(ctx)
	}()

	validate := pflag.BoolP("validate", "v", false, "validate di")
	fetch := pflag.String("fetch", "", "run a single request and print the result to stdout")
	method := pflag.StringP("method", "m", "get", "method for fetch")
	headers := pflag.StringArrayP("header", "h", []string{}, "header")
	body := pflag.StringP("body", "b", "", "body")
	username := pflag.StringP("user", "u", "", "uername")

	pflag.Parse()

	err := k.Validate(ctx)
	if err != nil {
		clog.Use(ctx).Warn("validation errors", "err", err)
	}
	if *validate {
		if err != nil {
			os.Exit(1)
		}
		return nil
	}

	if *fetch != "" {
		return k.runFetch(ctx, *fetch, *method, *headers, *body, *username)
	}

	k.StartServices(ctx)

	return k.RunHttpServer(ctx)
}
func (k *Kernel) handlerWithMiddleware() http.Handler {
	h := k.rootHandler

	for _, m := range k.globalMiddleware {
		h = m.Middleware(h)
	}
	return h
}

func (k *Kernel) HttpServer(ctx context.Context) *http.Server {
	clog.Use(ctx).Info(fmt.Sprintf("listening at http://localhost:%d", k.cfg.GetHTTPPort()))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", k.cfg.GetHTTPPort()),
		Handler: k.handlerWithMiddleware(),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// k.addCloser(server)
	di.RegisterSingleton(ctx, func() *http.Server {
		return server
	})

	return server
}

func (k *Kernel) RunHttpServer(ctx context.Context) error {
	return k.HttpServer(ctx).ListenAndServe()
}

func (k *Kernel) StartServices(ctx context.Context) {
	for _, s := range k.services {
		ctx := clog.With(ctx, slog.String("service", s.Name()))
		go func(ctx context.Context, s Service) {
			for {
				if di.IsFillable(s) {
					err := di.Fill(ctx, s)
					if err != nil {
						clog.Use(ctx).Error("service dependency injection failed", slog.Any("error", err))
						return
					}
				}
				err := s.Run(ctx)
				if err == nil {
					return
				}
				clog.Use(ctx).Error("service failed", slog.Any("error", err))
				r, ok := s.(Restarter)
				if !ok {
					return
				}
				r.Restart()
			}
		}(ctx, s)
	}
}

func (k *Kernel) singles(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		logger := clog.Use(ctx)
		tries := 0
		for range c {
			if tries == 0 {
				go func() {
					logger.Info("Gracefully shutting down server Ctrl+C again to force")
					k.closeAndLog(ctx)
					os.Exit(1)
				}()
			} else {
				logger.Warn("Force shutting down server")
				os.Exit(1)
			}
			tries++
		}
	}()
}

func (k *Kernel) Close() error {
	return errors.Join(k.closeAll()...)
}

type resourceError struct {
	err      error
	resource string
}

func (e *resourceError) Error() string {
	return fmt.Sprintf("resource %s: %v", e.resource, e.err)
}

func (e *resourceError) Unwrap() error {
	return e.err
}

func (k *Kernel) closeAndLog(ctx context.Context) {
	logger := clog.Use(ctx)
	errs := k.closeAll()
	for _, err := range errs {
		if resErr, ok := err.(*resourceError); ok {
			logger.Error("failed to close",
				"err", resErr.err,
				"resource", resErr.resource,
			)
		} else {
			logger.Error("failed to close",
				"err", err,
			)
		}
	}
}
func (k *Kernel) closeAll() []error {
	errs := []error{}

	for _, s := range k.dependencyProvider.Singletons() {
		v, err, ready := s.Peek()
		if err != nil || !ready || v == k {
			continue
		}
		if closer, ok := v.(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				errs = append(errs, &resourceError{
					err:      err,
					resource: reflect.TypeOf(closer).String(),
				})
			}
		}
	}
	return errs
}
