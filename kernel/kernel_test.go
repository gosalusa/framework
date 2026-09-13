package kernel

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/di"
	"gosalusa.com/openapidoc"
	"gosalusa.com/router"
	"gosalusa.com/salusaconfig"
)

type testConfig struct {
	port    int
	baseURL string
}

func (c *testConfig) GetHTTPPort() int { return c.port }
func (c *testConfig) GetBaseURL() string {
	if c.baseURL == "" {
		return "https://example.com"
	}
	return c.baseURL
}

var _ salusaconfig.Config = (*testConfig)(nil)

type emptyBaseConfig struct{}

func (c *emptyBaseConfig) GetHTTPPort() int { return 8080 }
func (c *emptyBaseConfig) GetBaseURL() string {
	return ""
}

var _ salusaconfig.Config = (*emptyBaseConfig)(nil)

func okRootHandler() func(ctx context.Context) http.Handler {
	return func(ctx context.Context) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	}
}

type testService struct{}

func (t *testService) Run(ctx context.Context) error { return nil }
func (t *testService) Name() string                  { return "test" }

func TestNewAndOptions(t *testing.T) {
	k := New(
		Bootstrap(func(ctx context.Context) error { return nil }),
		Config(func() *testConfig { return &testConfig{port: 8080} }),
		RootHandler(okRootHandler()),
		Middleware([]router.Middleware{
			router.MiddlewareFunc(func(next http.Handler) http.Handler { return next }),
		}),
		Services(&testService{}),
		APIDocumentation(openapidoc.Info(spec.InfoProps{Title: "title", Version: "1.0", Description: "desc"})),
		FetchAuth(func(ctx context.Context, username string, r *http.Request) error { return nil }),
	)
	assert.NotNil(t, k)
	assert.False(t, k.bootstrapped)
	assert.Len(t, k.services, 1)
	assert.NotNil(t, k.docs)
	assert.NotNil(t, k.fetchAuth)
	assert.Len(t, k.globalMiddleware, 1)
}

func TestInitRoutes(t *testing.T) {
	k := New(InitRoutes(func(r *router.Router) {
		r.Get("/", &testHandler{})
	}))
	ctx := di.TestDependencyProviderContext()
	di.RegisterSingleton(ctx, func() salusaconfig.Config { return &testConfig{} })
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "https://example.com/", http.NoBody)
	w := httptest.NewRecorder()
	k.RootHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBootstrap(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	k := New(Config(func() *testConfig { return &testConfig{} }), RootHandler(okRootHandler()))

	err := k.Bootstrap(ctx)
	assert.NoError(t, err)

	err = k.Bootstrap(ctx)
	assert.ErrorIs(t, err, ErrAlreadyBootstrapped)

	assert.NotNil(t, k.Config())

	assert.Panics(t, func() {
		k2 := New(RootHandler(okRootHandler()))
		k2.RootHandler()
	})
}

func TestAPIDoc(t *testing.T) {
	t.Run("router paths", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		k := New(InitRoutes(func(r *router.Router) {
			r.Get("/things", &testHandler{})
		}), Config(func() *testConfig { return &testConfig{} }))
		di.RegisterSingleton(ctx, func() salusaconfig.Config { return &testConfig{} })
		di.RegisterSingleton(ctx, func() router.URLResolver {
			return router.NewResolver("https://example.com", router.New())
		})
		err := k.Bootstrap(ctx)
		require.NoError(t, err)

		docs, err := k.APIDoc(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "2.0", docs.Swagger)
		assert.Contains(t, docs.Paths.Paths, "/things")
		assert.Equal(t, "https", docs.Schemes[0])
		assert.Equal(t, "example.com", docs.Host)
	})

	t.Run("no url resolver", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		k := New(Config(func() *testConfig { return &testConfig{} }), RootHandler(okRootHandler()))
		err := k.Bootstrap(ctx)
		require.NoError(t, err)

		docs, err := k.APIDoc(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "2.0", docs.Swagger)
	})

	t.Run("invalid url", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		k := New(Config(func() *testConfig { return &testConfig{} }), RootHandler(okRootHandler()))
		di.RegisterSingleton(ctx, func() router.URLResolver {
			return &badResolver{}
		})
		err := k.Bootstrap(ctx)
		require.NoError(t, err)

		_, err = k.APIDoc(ctx)
		assert.Error(t, err)
	})

	t.Run("no host", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		k := New(Config(func() *testConfig { return &testConfig{} }), RootHandler(okRootHandler()))
		di.RegisterSingleton(ctx, func() router.URLResolver {
			return router.NewTestResolver()
		})
		err := k.Bootstrap(ctx)
		require.NoError(t, err)

		_, err = k.APIDoc(ctx)
		assert.Error(t, err)
	})

	t.Run("with docs option", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		k := New(
			Config(func() *testConfig { return &testConfig{} }),
			RootHandler(okRootHandler()),
			APIDocumentation(openapidoc.Info(spec.InfoProps{Title: "t", Version: "1", Description: "d"})),
		)
		di.RegisterSingleton(ctx, func() salusaconfig.Config { return &testConfig{} })
		err := k.Bootstrap(ctx)
		require.NoError(t, err)

		docs, err := k.APIDoc(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "t", docs.Info.Title)
	})
}

type badResolver struct{}

func (b *badResolver) Resolve(name string, params ...any) string {
	return "http://[::1]:namedport"
}
func (b *badResolver) ResolveHandler(h http.Handler, params ...any) string {
	return "/"
}

func TestNewRequest(t *testing.T) {
	t.Run("absolute", func(t *testing.T) {
		r, err := newRequest(context.Background(), "https://example.com/foo?x=1", "get", "")
		assert.NoError(t, err)
		assert.Equal(t, "/foo?x=1", r.URL.String())
		assert.Equal(t, "example.com", r.Host)
		assert.Equal(t, "GET", r.Method)
	})

	t.Run("invalid absolute", func(t *testing.T) {
		_, err := newRequest(context.Background(), "http://[::1]:namedport", "get", "")
		assert.Error(t, err)
	})

	t.Run("relative with config", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() salusaconfig.Config { return &testConfig{baseURL: "https://example.com/base"} })
		r, err := newRequest(ctx, "/foo", "POST", "body")
		assert.NoError(t, err)
		assert.Equal(t, "/base/foo", r.URL.String())
		assert.Equal(t, "example.com", r.Host)
		assert.Equal(t, "POST", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "body", string(b))
	})

	t.Run("relative no config", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		_, err := newRequest(ctx, "/foo", "get", "")
		assert.Error(t, err)
	})

	t.Run("relative empty base", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() salusaconfig.Config { return &emptyBaseConfig{} })
		r, err := newRequest(ctx, "/foo", "get", "")
		assert.NoError(t, err)
		assert.Equal(t, "/foo", r.URL.String())
		assert.Equal(t, "localhost", r.Host)
	})
}

func TestResponseWriter(t *testing.T) {
	var body, headers bytes.Buffer
	w := &ResponseWriter{
		header:  http.Header{},
		body:    &body,
		headers: &headers,
		status:  0,
	}

	w.Header().Set("X-Test", "1")
	_, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Status())
	assert.True(t, w.Ok())
	assert.Equal(t, "hello", body.String())
	assert.Contains(t, headers.String(), "< status: 200")

	w.WriteHeader(500)
	assert.Equal(t, 200, w.Status(), "WriteHeader should be ignored after first write")

	w2 := &ResponseWriter{
		header:  http.Header{},
		body:    &body,
		headers: &headers,
		status:  0,
	}
	w2.Header().Set("X-Test", "1")
	w2.WriteHeader(404)
	assert.Equal(t, 404, w2.Status())
	assert.False(t, w2.Ok())
	assert.Contains(t, headers.String(), "< header: X-Test: 1")
	assert.Contains(t, headers.String(), "< status: 404")
}

func TestRunFetch(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	var authUser string
	k := New(
		Config(func() *testConfig { return &testConfig{} }),
		RootHandler(okRootHandler()),
		FetchAuth(func(ctx context.Context, username string, r *http.Request) error {
			authUser = username
			r.Header.Set("X-Auth", username)
			return nil
		}),
	)
	di.RegisterSingleton(ctx, func() salusaconfig.Config { return &testConfig{} })
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	err = k.runFetch(ctx, "https://example.com/", "get", []string{"X-Custom: value"}, "", "user")
	assert.NoError(t, err)
	assert.Equal(t, "user", authUser)

	err = k.runFetch(ctx, "https://example.com/", "get", []string{"invalid-header"}, "", "user")
	assert.Error(t, err)

	err = k.runFetch(ctx, "https://example.com/", "get", nil, "", "user")
	assert.NoError(t, err)
}

func TestStartServices(t *testing.T) {
	t.Run("service returns nil", func(t *testing.T) {
		done := make(chan struct{})
		k := New(RootHandler(okRootHandler()))
		k.services = []Service{&channelService{done: done, err: nil}}
		k.StartServices(context.Background())

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("service did not run")
		}
	})

	t.Run("service errors and restarts", func(t *testing.T) {
		done := make(chan struct{})
		s := &restartingService{done: done}
		k := New(RootHandler(okRootHandler()))
		k.services = []Service{s}

		k.StartServices(context.Background())
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("service did not finish")
		}
		assert.True(t, s.restarts > 0)
	})
}

type channelService struct {
	done chan struct{}
	err  error
}

func (c *channelService) Run(ctx context.Context) error {
	close(c.done)
	return c.err
}
func (c *channelService) Name() string { return "channel" }

type restartingService struct {
	done     chan struct{}
	runs     int
	restarts int
}

func (r *restartingService) Run(ctx context.Context) error {
	r.runs++
	if r.runs >= 2 {
		close(r.done)
		return nil
	}
	return io.ErrClosedPipe
}
func (r *restartingService) Name() string { return "restarting" }
func (r *restartingService) Restart() {
	r.restarts++
}

func TestClose(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	k := New(Config(func() *testConfig { return &testConfig{} }), RootHandler(okRootHandler()))
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	di.RegisterSingleton(ctx, func() io.Closer {
		return &failCloser{}
	})

	err = k.Close()
	assert.Error(t, err)

	errs := k.closeAll()
	assert.Len(t, errs, 1)
	re, ok := errs[0].(*resourceError)
	assert.True(t, ok)
	assert.Contains(t, re.Error(), "failCloser")
}

type failCloser struct{}

func (f *failCloser) Close() error { return io.ErrClosedPipe }

func TestCloseAndLog(t *testing.T) {
	oldDefault := slog.Default()
	defer slog.SetDefault(oldDefault)
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{} }),
		RootHandler(okRootHandler()),
	)
	err := k.Bootstrap(ctx)
	require.NoError(t, err)
	di.RegisterSingleton(ctx, func() io.Closer {
		return &failCloser{}
	})
	k.closeAndLog(ctx)
}

func TestValidate(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{} }),
		RootHandler(func(ctx context.Context) http.Handler {
			return &validatingHandler{}
		}),
	)
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	err = k.Validate(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestServiceFuncs(t *testing.T) {
	var sf ServiceFunc = func() error { return nil }
	assert.NoError(t, sf.Run(context.Background()))

	var sfr ServiceFuncRestart = func() error { return nil }
	assert.NoError(t, sfr.Run(context.Background()))
}

func TestResourceError(t *testing.T) {
	re := &resourceError{err: io.ErrClosedPipe, resource: "thing"}
	assert.Contains(t, re.Error(), "thing")
	assert.ErrorIs(t, re.Unwrap(), io.ErrClosedPipe)
}

func TestRun(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"test", "--fetch", "https://example.com/", "--method", "get"}

	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{} }),
		RootHandler(okRootHandler()),
	)
	di.RegisterSingleton(ctx, func() salusaconfig.Config { return &testConfig{} })
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	err = k.Run(ctx)
	assert.NoError(t, err)
}

func TestRunHttpServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{port: ln.Addr().(*net.TCPAddr).Port} }),
		RootHandler(okRootHandler()),
	)
	di.RegisterSingleton(ctx, func() salusaconfig.Config {
		return &testConfig{port: ln.Addr().(*net.TCPAddr).Port}
	})
	err = k.Bootstrap(ctx)
	require.NoError(t, err)

	err = k.RunHttpServer(ctx)
	assert.Error(t, err)
}

func TestBootstrapErrors(t *testing.T) {
	t.Run("bootstrap step error", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		stepErr := context.Canceled
		k := New(
			Config(func() *testConfig { return &testConfig{} }),
			RootHandler(okRootHandler()),
			Bootstrap(func(ctx context.Context) error { return stepErr }),
		)
		err := k.Bootstrap(ctx)
		assert.ErrorIs(t, err, stepErr)
	})
}

func TestValidateWithService(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{} }),
		RootHandler(okRootHandler()),
		Services(&validatingService{}),
	)
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	err = k.Validate(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

type validatingService struct{}

func (v *validatingService) Run(ctx context.Context) error { return nil }
func (v *validatingService) Name() string                  { return "validating" }
func (v *validatingService) Validate(ctx context.Context) error {
	return context.Canceled
}

func TestStartServicesFillable(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	di.RegisterSingleton(ctx, func() string { return "hello" })

	done := make(chan struct{})
	s := &fillableService{done: done}
	k := New(RootHandler(okRootHandler()))
	k.services = []Service{s}

	k.StartServices(ctx)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("service did not run")
	}
	assert.Equal(t, "hello", s.Greeting)
}

type fillableService struct {
	Greeting string `inject:""`
	done     chan struct{}
}

func (f *fillableService) Run(ctx context.Context) error {
	close(f.done)
	return nil
}
func (f *fillableService) Name() string { return "fillable" }

func TestHandlerWithMiddleware(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{} }),
		RootHandler(okRootHandler()),
	)
	err := k.Bootstrap(ctx)
	require.NoError(t, err)
	h := k.handlerWithMiddleware()
	req := httptest.NewRequest("GET", "https://example.com/", http.NoBody)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHttpServer(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	k := New(
		Config(func() *testConfig { return &testConfig{port: 9090} }),
		RootHandler(okRootHandler()),
	)
	err := k.Bootstrap(ctx)
	require.NoError(t, err)

	srv := k.HttpServer(ctx)
	assert.Equal(t, ":9090", srv.Addr)

	got, err := di.Resolve[*http.Server](ctx)
	assert.NoError(t, err)
	assert.Equal(t, srv, got)
}

func TestSingles(t *testing.T) {
	k := New(RootHandler(okRootHandler()))
	k.singles(context.Background())
}

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (t *testHandler) Operation(ctx context.Context) (*spec.Operation, error) {
	return spec.NewOperation("test"), nil
}

type validatingHandler struct {
	testHandler
}

func (v *validatingHandler) Validate(ctx context.Context) error {
	return context.Canceled
}
