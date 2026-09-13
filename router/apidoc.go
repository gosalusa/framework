package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-openapi/spec"
	"gosalusa.com/openapidoc"
)

var _ openapidoc.Pathser = (*Router)(nil)

func (r *Router) Paths(ctx context.Context, basePath string) (*spec.Paths, error) {
	var err error
	paths := &spec.Paths{
		Paths: map[string]spec.PathItem{},
	}
	for _, route := range r.routes.Routes {
		if basePath != "" && !strings.HasPrefix(route.Path, basePath) {
			continue
		}
		path := strings.TrimPrefix(route.Path, basePath)
		pathItem, ok := paths.Paths[path]
		if !ok {
			pathItem = spec.PathItem{}
		}

		var op *spec.Operation
		oper, ok := route.handler.(openapidoc.Operationer)
		if !ok {
			continue
		}
		op, err = oper.Operation(ctx)
		if err != nil {
			return nil, err
		}

		middleware := route.GetMiddleware()
		for _, m := range middleware {
			if opm, ok := m.(openapidoc.OperationMiddleware); ok {
				op = opm.OperationMiddleware(op)
			}
		}

		if op.ID == "" {
			op.ID = route.name
		}

		switch route.Method {
		case http.MethodGet:
			pathItem.Get = op
		case http.MethodHead:
			pathItem.Head = op
		case http.MethodPost:
			pathItem.Post = op
		case http.MethodPut:
			pathItem.Put = op
		case http.MethodPatch:
			pathItem.Patch = op
		case http.MethodDelete:
			pathItem.Delete = op
		case http.MethodOptions:
			pathItem.Options = op
		default:
			continue
			// return nil, fmt.Errorf("unsupported method %s", route.Method)
		}
		paths.Paths[path] = pathItem
	}
	return paths, nil
}
