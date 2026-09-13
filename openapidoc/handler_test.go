package openapidoc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
)

type testAPIDocer struct {
	doc *spec.Swagger
	err error
}

func (d *testAPIDocer) APIDoc(context.Context) (*spec.Swagger, error) {
	if d.err != nil {
		return nil, d.err
	}
	if d.doc == nil {
		d.doc = &spec.Swagger{SwaggerProps: spec.SwaggerProps{Swagger: "2.0"}}
	}
	return d.doc, nil
}

func TestSwaggerUI(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	di.Register(ctx, func(ctx context.Context, tag string) (APIDocer, error) {
		return &testAPIDocer{}, nil
	})

	h := SwaggerUI()

	t.Run("swagger.json", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/swagger.json", nil).WithContext(ctx)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Contains(t, rw.Header().Get("Content-Type"), "application/json")
		assert.Contains(t, rw.Body.String(), `"swagger": "2.0"`)
	})

	t.Run("redoc script", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/redoc.standalone.js", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "gzip", rw.Header().Get("Content-Encoding"))
		assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
		assert.Equal(t, strconv.Itoa(len(redocSrc)), rw.Header().Get("Content-Length"))
		assert.NotEmpty(t, rw.Body.Bytes())
	})

	t.Run("index", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/html", rw.Header().Get("Content-Type"))
		assert.Equal(t, string(redocIndex), rw.Body.String())
	})

	t.Run("redirect to trailing slash", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assert.Equal(t, http.StatusFound, rw.Code)
		assert.Equal(t, "/api/", rw.Header().Get("Location"))
	})
}

func TestServeSwaggerPanicsOnResolve(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	req := httptest.NewRequest("GET", "/swagger.json", nil).WithContext(ctx)
	rw := httptest.NewRecorder()
	assert.Panics(t, func() {
		SwaggerUI().ServeHTTP(rw, req)
	})
}

func TestServeSwaggerPanicsOnAPIDocError(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	di.Register(ctx, func(ctx context.Context, tag string) (APIDocer, error) {
		return &testAPIDocer{err: errors.New("doc failed")}, nil
	})
	req := httptest.NewRequest("GET", "/swagger.json", nil).WithContext(ctx)
	rw := httptest.NewRecorder()
	assert.Panics(t, func() {
		SwaggerUI().ServeHTTP(rw, req)
	})
}
