package router

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

type Middleware interface {
	Middleware(next http.Handler) http.Handler
}
type MiddlewareFunc func(next http.Handler) http.Handler

func (f MiddlewareFunc) Middleware(next http.Handler) http.Handler {
	return f(next)
}

type InlineMiddlewareFunc func(w http.ResponseWriter, r *http.Request, next http.Handler)

func (f InlineMiddlewareFunc) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f(w, r, next)
	})
}

type Route struct {
	Path       string
	Method     string
	name       string
	handler    http.Handler
	router     *Router
	route      *mux.Route
	middleware []Middleware
}

func (r *Route) GetMiddleware() []Middleware {
	return r.router.middleware
}

func (r *Route) Name(name string) *Route {
	r.name = name
	r.route.Name(name)
	return r
}

func (r *Route) Middleware(middleware Middleware) *Route {
	r.handler = middleware.Middleware(r.handler)
	r.route.Handler(r.handler)
	r.middleware = append(r.middleware, middleware)
	return r
}

type routeList struct {
	Routes []*Route
}

type Router struct {
	prefix     string
	router     *mux.Router
	routes     *routeList
	middleware []Middleware
}

func New() *Router {
	return &Router{
		prefix:     "",
		router:     mux.NewRouter(),
		routes:     &routeList{Routes: []*Route{}},
		middleware: []Middleware{},
	}
}

func (r *Router) Get(path string, handler http.Handler) *Route {
	return r.handleMethod(http.MethodGet, path, handler)
}
func (r *Router) Post(path string, handler http.Handler) *Route {
	return r.handleMethod(http.MethodPost, path, handler)
}
func (r *Router) Put(path string, handler http.Handler) *Route {
	return r.handleMethod(http.MethodPut, path, handler)
}
func (r *Router) Patch(path string, handler http.Handler) *Route {
	return r.handleMethod(http.MethodPatch, path, handler)
}
func (r *Router) Delete(path string, handler http.Handler) *Route {
	return r.handleMethod(http.MethodDelete, path, handler)
}

func (r *Router) GetFunc(path string, handler http.HandlerFunc) *Route {
	return r.Get(path, handler)
}
func (r *Router) PostFunc(path string, handler http.HandlerFunc) *Route {
	return r.Post(path, handler)
}
func (r *Router) PutFunc(path string, handler http.HandlerFunc) *Route {
	return r.Put(path, handler)
}
func (r *Router) PatchFunc(path string, handler http.HandlerFunc) *Route {
	return r.Patch(path, handler)
}
func (r *Router) DeleteFunc(path string, handler http.HandlerFunc) *Route {
	return r.Delete(path, handler)
}

func (r *Router) handleMethod(method, path string, handler http.Handler) *Route {
	return r.addRoute(r.router.Handle(path, handler).Methods(method), handler, path, method)
}
func (r *Router) Handle(p string, handler http.Handler) *Route {
	return r.addRoute(r.router.PathPrefix(p).Handler(handler), handler, path.Join(p, "*"), "ALL")
}

func (r *Router) UseFunc(middleware func(http.Handler) http.Handler) {
	m := MiddlewareFunc(middleware)
	r.router.Use(m.Middleware)
	r.middleware = append(r.middleware, m)
}
func (r *Router) Use(middleware Middleware) {
	r.router.Use(middleware.Middleware)
	r.middleware = append(r.middleware, middleware)
}

func (r *Router) Group(prefix string, cb func(r *Router)) {
	middleware := make([]Middleware, len(r.middleware))
	copy(middleware, r.middleware)
	cb(&Router{
		prefix:     path.Join(r.prefix, prefix),
		router:     r.router.PathPrefix(prefix).Subrouter(),
		routes:     r.routes,
		middleware: middleware,
	})
}

func (r *Router) addRoute(muxRoute *mux.Route, handler http.Handler, pathName, method string) *Route {
	route := &Route{
		Path:       path.Join(r.prefix, pathName),
		Method:     method,
		handler:    handler,
		router:     r,
		route:      muxRoute,
		middleware: []Middleware{},
	}
	r.routes.Routes = append(r.routes.Routes, route)
	return route
}

func (r *Router) Routes() []*Route {
	return r.routes.Routes
}
func (r *Router) PrintRoutes() {
	for _, route := range r.routes.Routes {
		fmt.Printf("%-40s %s\n", route.Path, route.Method)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
