package router_test

import (
	"gosalusa.com/request"
	"gosalusa.com/router"
)

func ExampleRouter() {
	r := router.New()

	r.Group("/test", func(r *router.Router) {
		r.Get("/", request.Handler(func(r *any) (any, error) {
			return nil, nil
		}))
	})
}
