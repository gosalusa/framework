package router_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/openapidoc"
	"gosalusa.com/router"
	"gosalusa.com/salusaconfig"
)

type testHandler struct {
	op      *spec.Operation
	callErr error
}

func (t *testHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (t *testHandler) Operation(context.Context) (*spec.Operation, error) {
	if t.callErr != nil {
		return nil, t.callErr
	}
	if t.op != nil {
		return t.op, nil
	}
	return spec.NewOperation("test"), nil
}
func (t *testHandler) Validate(context.Context) error { return t.callErr }

var _ openapidoc.Operationer = (*testHandler)(nil)
var _ interface{ Validate(context.Context) error } = (*testHandler)(nil)

type opMiddleware struct{}

func (opMiddleware) Middleware(next http.Handler) http.Handler { return next }

func (opMiddleware) OperationMiddleware(op *spec.Operation) *spec.Operation {
	op.Summary = "modified"
	return op
}

type config struct {
	baseURL string
	port    int
}

func (c *config) GetHTTPPort() int { return c.port }
func (c *config) GetBaseURL() string {
	return c.baseURL
}

var _ salusaconfig.Config = (*config)(nil)

func TestRouterHTTPMethods(t *testing.T) {
	r := router.New()

	h := &testHandler{}
	r.Get("/get", h)
	r.Post("/post", h)
	r.Put("/put", h)
	r.Patch("/patch", h)
	r.Delete("/delete", h)
	r.GetFunc("/get-func", func(w http.ResponseWriter, req *http.Request) {})
	r.PostFunc("/post-func", func(w http.ResponseWriter, req *http.Request) {})
	r.PutFunc("/put-func", func(w http.ResponseWriter, req *http.Request) {})
	r.PatchFunc("/patch-func", func(w http.ResponseWriter, req *http.Request) {})
	r.DeleteFunc("/delete-func", func(w http.ResponseWriter, req *http.Request) {})
	r.Handle("/handle", h)

	assert.Len(t, r.Routes(), 11)

	req := httptest.NewRequest("GET", "https://example.com/get", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("POST", "https://example.com/post", http.NoBody)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouteNameMiddleware(t *testing.T) {
	r := router.New()
	mw := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return next
	})
	h := &testHandler{}
	route := r.Get("/get", h).Name("named")
	assert.Equal(t, "/get", route.Path)
	assert.Len(t, route.GetMiddleware(), 0)

	route.Middleware(mw)
	assert.Len(t, route.GetMiddleware(), 0)

	r.Use(mw)
	assert.Len(t, route.GetMiddleware(), 1)
}

func TestInlineMiddlewareFunc(t *testing.T) {
	var called bool
	mw := router.InlineMiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		called = true
		next.ServeHTTP(w, r)
	})

	r := router.New()
	r.Use(mw)
	r.Get("/", &testHandler{})

	req := httptest.NewRequest("GET", "https://example.com/", http.NoBody)
	r.ServeHTTP(httptest.NewRecorder(), req)
	assert.True(t, called)
}

func TestRouterGroup(t *testing.T) {
	r := router.New()
	var mwCalled bool
	mw := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mwCalled = true
			next.ServeHTTP(w, r)
		})
	})
	r.Use(mw)
	r.Group("/api", func(r *router.Router) {
		r.Get("/users", &testHandler{})
	})
	r.Get("/home", &testHandler{})

	assert.Len(t, r.Routes(), 2)
	assert.Equal(t, "/api/users", r.Routes()[0].Path)

	req := httptest.NewRequest("GET", "https://example.com/api/users", http.NoBody)
	r.ServeHTTP(httptest.NewRecorder(), req)
	assert.True(t, mwCalled)
}

func TestRouterPrintRoutes(t *testing.T) {
	r := router.New()
	r.Get("/x", &testHandler{}).Name("x")
	r.PrintRoutes()
}

func TestRouterPaths(t *testing.T) {
	t.Run("all methods", func(t *testing.T) {
		r := router.New()
		r.Get("/get", &testHandler{}).Name("get")
		r.Post("/post", &testHandler{})
		r.Put("/put", &testHandler{})
		r.Patch("/patch", &testHandler{})
		r.Delete("/delete", &testHandler{})
		r.Handle("/all", &testHandler{})

		paths, err := r.Paths(context.Background(), "")
		assert.NoError(t, err)
		assert.Contains(t, paths.Paths, "/get")
		assert.NotNil(t, paths.Paths["/get"].Get)
		assert.NotNil(t, paths.Paths["/post"].Post)
		assert.NotNil(t, paths.Paths["/put"].Put)
		assert.NotNil(t, paths.Paths["/patch"].Patch)
		assert.NotNil(t, paths.Paths["/delete"].Delete)
	})

	t.Run("base path", func(t *testing.T) {
		r := router.New()
		r.Get("/api/get", &testHandler{})
		r.Get("/other", &testHandler{})

		paths, err := r.Paths(context.Background(), "/api")
		assert.NoError(t, err)
		_, ok := paths.Paths["/get"]
		assert.True(t, ok)
		_, ok = paths.Paths["/other"]
		assert.False(t, ok)
	})

	t.Run("operation middleware", func(t *testing.T) {
		r := router.New()
		r.Use(opMiddleware{})
		r.Get("/get", &testHandler{})
		r.Get("/plain", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		paths, err := r.Paths(context.Background(), "")
		assert.NoError(t, err)
		assert.Equal(t, "modified", paths.Paths["/get"].Get.Summary)
	})

	t.Run("operation error", func(t *testing.T) {
		r := router.New()
		r.Get("/get", &testHandler{callErr: context.Canceled})

		_, err := r.Paths(context.Background(), "")
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestRouterValidate(t *testing.T) {
	r := router.New()
	r.Get("/ok", &testHandler{})
	r.Get("/err", &testHandler{callErr: context.Canceled})

	err := r.Validate(context.Background())
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRouterRegister(t *testing.T) {
	t.Run("base url from config", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() salusaconfig.Config {
			return &config{baseURL: "https://example.com"}
		})
		r := router.New()
		r.Register(ctx)

		resolver, err := di.Resolve[router.URLResolver](ctx)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/test", resolver.Resolve("test"))
	})

	t.Run("no config url, no request", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() salusaconfig.Config {
			return &config{}
		})
		r := router.New()
		r.Register(ctx)

		resolver, err := di.Resolve[router.URLResolver](ctx)
		assert.NoError(t, err)
		assert.Equal(t, "/test", resolver.Resolve("test"))
	})

	t.Run("origin header", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() salusaconfig.Config {
			return &config{}
		})
		req := httptest.NewRequest("GET", "https://internal", http.NoBody)
		req.Header.Set("Origin", "https://origin.example.com")
		di.RegisterSingleton(ctx, func() *http.Request { return req })
		r := router.New()
		r.Register(ctx)

		resolver, err := di.Resolve[router.URLResolver](ctx)
		assert.NoError(t, err)
		assert.Equal(t, "https://origin.example.com/test", resolver.Resolve("test"))
	})

	t.Run("host fallback", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() salusaconfig.Config {
			return &config{}
		})
		req := httptest.NewRequest("GET", "https://internal.example.com", http.NoBody)
		di.RegisterSingleton(ctx, func() *http.Request { return req })
		r := router.New()
		r.Register(ctx)

		resolver, err := di.Resolve[router.URLResolver](ctx)
		assert.NoError(t, err)
		assert.Equal(t, "http://internal.example.com/test", resolver.Resolve("test"))
	})
}

func TestSalusaResolverResolveHandler(t *testing.T) {
	h1 := &testHandler{}
	h2 := &testHandler{}
	r := router.New()
	r.Get("/named", h1).Name("named")

	resolver := router.NewResolver("https://example.com", r)

	u := resolver.Resolve("named", "id", 5)
	assert.Equal(t, "https://example.com/named?id=5", u)

	u = resolver.Resolve("does-not-exist", "id", 5)
	assert.Equal(t, "https://example.com/does-not-exist?id=5", u)

	u = resolver.ResolveHandler(h1, "id", 5)
	assert.Equal(t, "https://example.com/named?id=5", u)

	u = resolver.ResolveHandler(h2)
	assert.Equal(t, "https://example.com/", u)
}

func TestRouterHandleAll(t *testing.T) {
	r := router.New()
	r.Handle("/anything", &testHandler{})
	req := httptest.NewRequest("GET", "https://example.com/anything", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
