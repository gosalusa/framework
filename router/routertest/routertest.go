package routertest

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"gosalusa.com/router"
)

type TestResolver struct {
	Origin       string
	handlerPaths map[http.Handler]string
}

var _ router.URLResolver = (*TestResolver)(nil)

func NewTestResolver() *TestResolver {
	return &TestResolver{
		Origin:       "https://example.com",
		handlerPaths: map[http.Handler]string{},
	}
}

func (r *TestResolver) Resolve(name string, params ...any) string {
	attrs, err := router.ToAttrs(params)
	if err != nil {
		panic(err)
	}
	v := url.Values{}

	for _, a := range attrs {
		v.Add(a.Key, a.Value)
	}

	return strings.TrimSuffix(r.Origin, "/") + "/" + name + "?" + v.Encode()
}
func (r *TestResolver) ResolveHandler(h http.Handler, params ...any) string {
	p, ok := r.handlerPaths[h]
	if !ok {
		p = uuid.New().String()
		r.handlerPaths[h] = p
	}
	return r.Resolve(p, params...)
}
