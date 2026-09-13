package routertest_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/router"
	"gosalusa.com/router/routertest"
)

func TestNewTestResolver(t *testing.T) {
	var _ router.URLResolver = (*routertest.TestResolver)(nil)

	r := routertest.NewTestResolver()
	assert.Equal(t, "https://example.com", r.Origin)
}

func TestResolve(t *testing.T) {
	r := routertest.NewTestResolver()

	assert.Equal(t, "https://example.com/path?", r.Resolve("path"))

	assert.Equal(t, "https://example.com/path?key=value", r.Resolve("path", "key", "value"))

	assert.Equal(t, "https://example.com/path?a=1&b=2", r.Resolve("path", "a", 1, "b", 2))

	assert.Equal(t, "https://example.com/path?attr=val", r.Resolve("path", &router.Attr{Key: "attr", Value: "val"}))

	r.Origin = "https://example.com/"
	assert.Equal(t, "https://example.com/path?key=value", r.Resolve("path", "key", "value"))
}

func TestResolvePanicsOnInvalidParams(t *testing.T) {
	r := routertest.NewTestResolver()

	assert.Panics(t, func() {
		r.Resolve("path", 123, "value")
	})

	assert.Panics(t, func() {
		r.Resolve("path", "key", true)
	})
}

func TestResolveHandler(t *testing.T) {
	r := routertest.NewTestResolver()
	h := &concreteHandler{}

	url1 := r.ResolveHandler(h, "key", "value")
	assert.Contains(t, url1, "https://example.com/")
	assert.Contains(t, url1, "key=value")

	url2 := r.ResolveHandler(h, "key", "value")
	assert.Equal(t, url1, url2, "same handler should resolve to the same path")
}

type concreteHandler struct{}

func (c *concreteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
