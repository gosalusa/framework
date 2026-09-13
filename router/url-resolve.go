package router

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"gosalusa.com/database/model"
	"gosalusa.com/di"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/salusaconfig"
)

type URLResolver interface {
	Resolve(name string, params ...any) string
	ResolveHandler(h http.Handler, params ...any) string
}

func (r *Router) Register(ctx context.Context) {
	di.RegisterWith(ctx, func(ctx context.Context, tag string, cfg salusaconfig.Config) (URLResolver, error) {
		origin := cfg.GetBaseURL()

		if origin == "" {
			req, err := di.Resolve[*http.Request](ctx)
			if err == nil {
				origin = getOrigin(req)
			} else {
				origin = "/"
			}
		}

		return &SalusaResolver{
			origin: origin,
			router: r,
		}, nil
	})
}

func getOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin != "" {
		return origin
	}

	return "http://" + r.Host
}

type Attr struct {
	Key   string
	Value string
}

func ToAttrs(params []any) ([]*Attr, error) {
	var active *Attr
	attrs := []*Attr{}
	for _, p := range params {
		if active == nil {
			switch p := p.(type) {
			case string:
				active = &Attr{
					Key: p,
				}
			case *Attr:
				attrs = append(attrs, p)
			default:
				return nil, fmt.Errorf("invalid attrs: expected key string or attr received %s", reflect.TypeOf(p))
			}
		} else {
			switch p := p.(type) {
			case string, fmt.Stringer,
				int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64,
				float32, float64:

				active.Value = fmt.Sprint(p)
				attrs = append(attrs, active)
				active = nil
			case model.Model:
				pKeyValues, err := helpers.PrimaryKeyValue(p)
				if err != nil {
					return nil, err
				}
				if len(pKeyValues) != 1 {
					return nil, fmt.Errorf("invalid attrs: model must have only 1 primary key found %d", len(pKeyValues))
				}
				active.Value = fmt.Sprint(pKeyValues[0])
				attrs = append(attrs, active)
				active = nil
			default:
				return nil, fmt.Errorf("invalid attrs: expected value string or number received %s", reflect.TypeOf(p))
			}
		}
	}
	return attrs, nil
}

type SalusaResolver struct {
	origin string
	router *Router
}

func NewResolver(origin string, r *Router) *SalusaResolver {
	return &SalusaResolver{
		origin: origin,
		router: r,
	}
}

func (r *SalusaResolver) Resolve(name string, params ...any) string {
	s, ok := r.namePath(name)
	if !ok {
		s = name
	}
	return r.resolve(s, params...)
}
func (r *SalusaResolver) ResolveHandler(h http.Handler, params ...any) string {
	s, ok := r.handlerPath(h)
	if !ok {
		s = "/"
	}
	return r.resolve(s, params...)
}

func (r *SalusaResolver) namePath(name string) (string, bool) {
	for _, route := range r.router.Routes() {
		if route.name == name {
			return route.Path, true
		}
	}
	return "", false
}
func (r *SalusaResolver) handlerPath(h http.Handler) (string, bool) {
	for _, route := range r.router.Routes() {
		if route.handler == h {
			return route.Path, true
		}
	}
	return "", false
}

func (r *SalusaResolver) resolve(name string, params ...any) string {
	attrs, err := ToAttrs(params)
	if err != nil {
		panic(err)
	}
	v := url.Values{}

	path := name
	for _, a := range attrs {
		k := "{" + a.Key + "}"
		if strings.Contains(path, k) {
			path = strings.ReplaceAll(path, k, a.Value)
		} else {
			v.Add(a.Key, a.Value)
		}
	}

	u := strings.TrimSuffix(r.origin, "/") + "/" + strings.TrimPrefix(path, "/")
	if len(v) > 0 {
		u += "?" + v.Encode()
	}
	return u
}

type TestResolver struct{}

func NewTestResolver() URLResolver {
	return &TestResolver{}
}

var _ URLResolver = (*TestResolver)(nil)

// Resolve implements URLResolver.
func (t TestResolver) Resolve(name string, params ...any) string {
	attrs, err := ToAttrs(params)
	if err != nil {
		panic(err)
	}
	query := url.Values{}

	for _, attr := range attrs {
		query.Add(attr.Key, attr.Value)
	}

	return name + "?" + query.Encode()
}

// ResolveHandler implements URLResolver.
func (t TestResolver) ResolveHandler(h http.Handler, params ...any) string {
	return ""
}
