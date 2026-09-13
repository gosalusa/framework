package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/request"
)

func TestDIMiddleware(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	request.Register(ctx)

	type Request struct {
		Request *http.Request `inject:""`
	}

	httpRequest := httptest.NewRequest("GET", "http://0.0.0.0/", http.NoBody).WithContext(ctx)

	var gotReq *http.Request
	middleware := request.DIMiddleware()
	h := middleware(request.Handler(func(r *Request) (any, error) {
		gotReq = r.Request
		return nil, nil
	}))

	h.ServeHTTP(httptest.NewRecorder(), httpRequest)
	assert.NotNil(t, gotReq)
	assert.Equal(t, httpRequest.URL.String(), gotReq.URL.String())
}

func TestRegisterError(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	request.Register(ctx)

	_, err := di.Resolve[*http.Request](ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request not in context")

	_, err = di.Resolve[http.ResponseWriter](ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "response not in context")
}

func TestInjectRequest(t *testing.T) {
	type Request struct {
		Request *http.Request `inject:""`
	}
	ctx := di.TestDependencyProviderContext()
	request.Register(ctx)

	httpRequest := httptest.
		NewRequest("GET", "http://0.0.0.0/", http.NoBody).
		WithContext(ctx)

	h := request.Handler(func(r *Request) (any, error) {
		assert.Same(t, httpRequest, r.Request)
		return nil, nil
	})

	h.ServeHTTP(
		httptest.NewRecorder(),
		httpRequest,
	)
}

func TestInjectResponseWriter(t *testing.T) {
	type Request struct {
		ResponseWriter http.ResponseWriter `inject:""`
	}

	ctx := di.TestDependencyProviderContext()
	request.Register(ctx)

	rw := httptest.NewRecorder()

	h := request.Handler(func(r *Request) (any, error) {
		assert.Same(t, rw, r.ResponseWriter)
		return nil, nil
	})

	h.ServeHTTP(
		rw,
		httptest.
			NewRequest("GET", "http://0.0.0.0/", http.NoBody).
			WithContext(ctx),
	)
}
