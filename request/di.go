package request

import (
	"context"
	"fmt"
	"net/http"

	"gosalusa.com/di"
	"gosalusa.com/router"
)

type contextKey uint8

const (
	requestKey contextKey = iota
	responseKey
)

func Register(ctx context.Context) {
	di.Register(ctx, func(ctx context.Context, tag string) (*http.Request, error) {
		req, ok := ctx.Value(requestKey).(*http.Request)
		if !ok {
			return nil, fmt.Errorf("request not in context")
		}
		return req, nil
	})
	di.Register(ctx, func(ctx context.Context, tag string) (http.ResponseWriter, error) {
		resp, ok := ctx.Value(responseKey).(http.ResponseWriter)
		if !ok {
			return nil, fmt.Errorf("response not in context")
		}
		return resp, nil
	})
}

func DIMiddleware() router.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, requestKey, r)
			ctx = context.WithValue(ctx, responseKey, w)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
