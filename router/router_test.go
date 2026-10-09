package router_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/request"
	"gosalusa.com/router"
	"gosalusa.com/validate"
)

func ExampleRouter() {
	r := router.New()

	r.Group("/test", func(r *router.Router) {
		r.Get("/", request.Handler(func(r *any) (any, error) {
			return nil, nil
		}))
	})
}

func TestNewRouter(t *testing.T) {
	r := router.New()
	require.NotNil(t, r)
	assert.Empty(t, r.Routes())
}

func TestRouter_Methods(t *testing.T) {
	methods := []struct {
		name     string
		method   string
		register func(r *router.Router, path string, h http.Handler) *router.Route
	}{
		{name: "GET", method: http.MethodGet, register: func(r *router.Router, p string, h http.Handler) *router.Route { return r.Get(p, h) }},
		{name: "POST", method: http.MethodPost, register: func(r *router.Router, p string, h http.Handler) *router.Route { return r.Post(p, h) }},
		{name: "PUT", method: http.MethodPut, register: func(r *router.Router, p string, h http.Handler) *router.Route { return r.Put(p, h) }},
		{name: "PATCH", method: http.MethodPatch, register: func(r *router.Router, p string, h http.Handler) *router.Route { return r.Patch(p, h) }},
		{name: "DELETE", method: http.MethodDelete, register: func(r *router.Router, p string, h http.Handler) *router.Route { return r.Delete(p, h) }},
	}

	for _, tc := range methods {
		t.Run(tc.name, func(t *testing.T) {
			r := router.New()
			var called bool
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.name + " response"))
			})

			rt := tc.register(r, "/test-path", h)
			assert.Equal(t, "/test-path", rt.Path)
			assert.Equal(t, tc.method, rt.Method)

			// Matching request
			req := httptest.NewRequest(tc.method, "/test-path", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.True(t, called)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tc.name+" response", w.Body.String())

			// Wrong method request
			wrongMethod := http.MethodPost
			if tc.method == http.MethodPost {
				wrongMethod = http.MethodGet
			}
			called = false
			reqWrong := httptest.NewRequest(wrongMethod, "/test-path", nil)
			wWrong := httptest.NewRecorder()
			r.ServeHTTP(wWrong, reqWrong)
			assert.False(t, called)
			assert.Equal(t, http.StatusMethodNotAllowed, wWrong.Code)

			// Unknown path request
			reqUnknown := httptest.NewRequest(tc.method, "/non-existent", nil)
			wUnknown := httptest.NewRecorder()
			r.ServeHTTP(wUnknown, reqUnknown)
			assert.Equal(t, http.StatusNotFound, wUnknown.Code)
		})
	}
}

func TestRouter_FuncMethods(t *testing.T) {
	methods := []struct {
		name     string
		method   string
		register func(r *router.Router, path string, h http.HandlerFunc) *router.Route
	}{
		{name: "GET", method: http.MethodGet, register: func(r *router.Router, p string, h http.HandlerFunc) *router.Route { return r.GetFunc(p, h) }},
		{name: "POST", method: http.MethodPost, register: func(r *router.Router, p string, h http.HandlerFunc) *router.Route { return r.PostFunc(p, h) }},
		{name: "PUT", method: http.MethodPut, register: func(r *router.Router, p string, h http.HandlerFunc) *router.Route { return r.PutFunc(p, h) }},
		{name: "PATCH", method: http.MethodPatch, register: func(r *router.Router, p string, h http.HandlerFunc) *router.Route { return r.PatchFunc(p, h) }},
		{name: "DELETE", method: http.MethodDelete, register: func(r *router.Router, p string, h http.HandlerFunc) *router.Route { return r.DeleteFunc(p, h) }},
	}

	for _, tc := range methods {
		t.Run(tc.name, func(t *testing.T) {
			r := router.New()
			var called bool
			fn := func(w http.ResponseWriter, req *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.name + " func response"))
			}

			rt := tc.register(r, "/func-path", fn)
			assert.Equal(t, "/func-path", rt.Path)
			assert.Equal(t, tc.method, rt.Method)

			req := httptest.NewRequest(tc.method, "/func-path", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.True(t, called)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tc.name+" func response", w.Body.String())
		})
	}
}

func TestRouter_Handle(t *testing.T) {
	r := router.New()
	var calledPaths []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calledPaths = append(calledPaths, req.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	rt := r.Handle("/static", handler)
	assert.Equal(t, "/static/*", rt.Path)
	assert.Equal(t, "ALL", rt.Method)

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, m := range methods {
		req := httptest.NewRequest(m, "/static/sub/file.css", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	assert.Len(t, calledPaths, len(methods))

	reqOther := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	wOther := httptest.NewRecorder()
	r.ServeHTTP(wOther, reqOther)
	assert.Equal(t, http.StatusNotFound, wOther.Code)
}

func TestRoute_Name(t *testing.T) {
	r := router.New()
	rt := r.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})).Name("users.show")
	require.NotNil(t, rt)

	resolver := router.NewResolver("https://api.example.com", r)
	resolved := resolver.Resolve("users.show", "id", 123)
	assert.Equal(t, "https://api.example.com/users/123", resolved)
}

func TestRoute_Middleware_ExecutionAndIsolation(t *testing.T) {
	r := router.New()

	var mwCalled bool
	mw := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			mwCalled = true
			w.Header().Set("X-Custom-Header", "route-mw")
			next.ServeHTTP(w, req)
		})
	})

	r.Get("/route-a", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).Middleware(mw)

	r.Get("/route-b", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request to route-a should invoke mw
	reqA := httptest.NewRequest(http.MethodGet, "/route-a", nil)
	recA := httptest.NewRecorder()
	r.ServeHTTP(recA, reqA)
	assert.True(t, mwCalled)
	assert.Equal(t, "route-mw", recA.Header().Get("X-Custom-Header"))
	assert.Equal(t, http.StatusOK, recA.Code)

	// Request to route-b should NOT invoke mw
	mwCalled = false
	reqB := httptest.NewRequest(http.MethodGet, "/route-b", nil)
	recB := httptest.NewRecorder()
	r.ServeHTTP(recB, reqB)
	assert.False(t, mwCalled)
	assert.Empty(t, recB.Header().Get("X-Custom-Header"))
	assert.Equal(t, http.StatusOK, recB.Code)
}

func TestRoute_Middleware_ChainingOrder(t *testing.T) {
	r := router.New()
	var order []string

	mw1 := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, req)
		})
	})
	mw2 := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, req)
		})
	})

	// When chaining .Middleware(mw1).Middleware(mw2):
	// mw1 wraps handler: handler = mw1(H)
	// mw2 wraps that handler: handler = mw2(mw1(H))
	// So when request arrives, mw2 executes first, then mw1, then handler.
	r.Get("/chained", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})).Middleware(mw1).Middleware(mw2)

	req := httptest.NewRequest(http.MethodGet, "/chained", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, []string{"mw2", "mw1", "handler"}, order)
}

func TestRoute_Middleware_ShortCircuit(t *testing.T) {
	r := router.New()
	handlerCalled := false

	guardMw := router.InlineMiddlewareFunc(func(w http.ResponseWriter, req *http.Request, next http.Handler) {
		if req.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("unauthorized"))
			return
		}
		next.ServeHTTP(w, req)
	})

	r.Get("/protected", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})).Middleware(guardMw)

	// Without token
	reqUnauth := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recUnauth := httptest.NewRecorder()
	r.ServeHTTP(recUnauth, reqUnauth)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, recUnauth.Code)
	assert.Equal(t, "unauthorized", recUnauth.Body.String())

	// With token
	reqAuth := httptest.NewRequest(http.MethodGet, "/protected", nil)
	reqAuth.Header.Set("Authorization", "Bearer secret")
	recAuth := httptest.NewRecorder()
	r.ServeHTTP(recAuth, reqAuth)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, recAuth.Code)
}

type routeTestValidatorMiddleware struct {
	err error
}

func (m *routeTestValidatorMiddleware) Middleware(next http.Handler) http.Handler {
	return next
}

func (m *routeTestValidatorMiddleware) Validate(ctx context.Context) error {
	return m.err
}

var _ router.Middleware = (*routeTestValidatorMiddleware)(nil)
var _ validate.Validator = (*routeTestValidatorMiddleware)(nil)

func TestRoute_Middleware_Validator(t *testing.T) {
	t.Run("passes when middleware valid", func(t *testing.T) {
		r := router.New()
		mw := &routeTestValidatorMiddleware{err: nil}
		r.Get("/valid", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})).Middleware(mw)

		err := r.Validate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("fails when middleware validation fails", func(t *testing.T) {
		r := router.New()
		expectedErr := errors.New("route middleware invalid")
		mw := &routeTestValidatorMiddleware{err: expectedErr}
		r.Get("/invalid", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})).Middleware(mw)

		err := r.Validate(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestRoute_GetMiddleware(t *testing.T) {
	r := router.New()
	mw := router.MiddlewareFunc(func(next http.Handler) http.Handler { return next })
	r.Use(mw)

	route := r.Get("/path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	assert.Len(t, route.GetMiddleware(), 1)
}

func TestRouter_Use(t *testing.T) {
	r := router.New()
	var calls []string

	mw1 := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			calls = append(calls, "mw1")
			next.ServeHTTP(w, req)
		})
	})
	mw2 := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			calls = append(calls, "mw2")
			next.ServeHTTP(w, req)
		})
	})

	r.Use(mw1)
	r.Use(mw2)

	r.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls = append(calls, "handler")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, []string{"mw1", "mw2", "handler"}, calls)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_UseFunc(t *testing.T) {
	r := router.New()
	var executed bool

	r.UseFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			executed = true
			w.Header().Set("X-UseFunc", "true")
			next.ServeHTTP(w, req)
		})
	})

	route := r.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	assert.Len(t, route.GetMiddleware(), 1)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.True(t, executed)
	assert.Equal(t, "true", rec.Header().Get("X-UseFunc"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_Group(t *testing.T) {
	t.Run("nested group prefixes", func(t *testing.T) {
		r := router.New()
		var handlerCalled bool

		r.Group("/api", func(api *router.Router) {
			api.Group("/v1", func(v1 *router.Router) {
				v1.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					handlerCalled = true
					w.WriteHeader(http.StatusOK)
				}))
			})
		})

		require.Len(t, r.Routes(), 1)
		assert.Equal(t, "/api/v1/users", r.Routes()[0].Path)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("group middleware isolation", func(t *testing.T) {
		r := router.New()
		var parentMwCalled, groupMwCalled bool

		parentMw := router.MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				parentMwCalled = true
				next.ServeHTTP(w, req)
			})
		})
		groupMw := router.MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				groupMwCalled = true
				next.ServeHTTP(w, req)
			})
		})

		r.Use(parentMw)

		r.Group("/grouped", func(sub *router.Router) {
			sub.Use(groupMw)
			sub.Get("/inner", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
		})

		r.Get("/outer", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Calling grouped route should execute both
		reqInner := httptest.NewRequest(http.MethodGet, "/grouped/inner", nil)
		r.ServeHTTP(httptest.NewRecorder(), reqInner)
		assert.True(t, parentMwCalled)
		assert.True(t, groupMwCalled)

		// Calling outer route should execute ONLY parentMw
		parentMwCalled = false
		groupMwCalled = false
		reqOuter := httptest.NewRequest(http.MethodGet, "/outer", nil)
		r.ServeHTTP(httptest.NewRecorder(), reqOuter)
		assert.True(t, parentMwCalled)
		assert.False(t, groupMwCalled)
	})

	t.Run("subsequent parent middleware does not affect existing group", func(t *testing.T) {
		r := router.New()
		var subrouter *router.Router

		r.Group("/sub", func(sub *router.Router) {
			subrouter = sub
		})

		assert.Empty(t, subrouter.Routes())
		r.Use(router.MiddlewareFunc(func(next http.Handler) http.Handler { return next }))
		route := subrouter.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
		assert.Empty(t, route.GetMiddleware())
	})
}

func TestRouter_Routes_Order(t *testing.T) {
	r := router.New()
	r.Get("/a", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	r.Post("/b", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	r.Group("/c", func(sub *router.Router) {
		sub.Put("/d", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
		sub.Delete("/e", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	})

	routes := r.Routes()
	require.Len(t, routes, 4)
	assert.Equal(t, "/a", routes[0].Path)
	assert.Equal(t, "GET", routes[0].Method)
	assert.Equal(t, "/b", routes[1].Path)
	assert.Equal(t, "POST", routes[1].Method)
	assert.Equal(t, "/c/d", routes[2].Path)
	assert.Equal(t, "PUT", routes[2].Method)
	assert.Equal(t, "/c/e", routes[3].Path)
	assert.Equal(t, "DELETE", routes[3].Method)
}

func TestRouter_PrintRoutes_Output(t *testing.T) {
	r := router.New()
	r.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	r.Post("/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))

	rescueStdout := os.Stdout
	pipeReader, pipeWriter, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = pipeWriter

	r.PrintRoutes()

	_ = pipeWriter.Close()
	out, err := io.ReadAll(pipeReader)
	require.NoError(t, err)
	os.Stdout = rescueStdout

	output := string(out)
	assert.Contains(t, output, fmt.Sprintf("%-40s %s\n", "/users", "GET"))
	assert.Contains(t, output, fmt.Sprintf("%-40s %s\n", "/users", "POST"))
}

func TestRouter_ServeHTTP_Interface(t *testing.T) {
	var _ http.Handler = (*router.Router)(nil)

	r := router.New()
	r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("world"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "world", rec.Body.String())
}

func TestInlineMiddlewareFunc_Methods(t *testing.T) {
	called := false
	imf := router.InlineMiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		called = true
		next.ServeHTTP(w, r)
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := imf.Middleware(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddlewareFunc_Methods(t *testing.T) {
	called := false
	mf := router.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mf.Middleware(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}
