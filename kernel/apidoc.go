package kernel

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-openapi/spec"
	"gosalusa.com/di"
	"gosalusa.com/openapidoc"
	"gosalusa.com/router"
)

// var _ openapidoc.APIDocer = (*Kernel)(nil)

// Operation implements openapidoc.Operationer.
func (k *Kernel) APIDoc(ctx context.Context) (*spec.Swagger, error) {
	var docs spec.Swagger

	if k.docs != nil {
		docs = *k.docs
	}
	h := k.RootHandler()
	if paths, ok := h.(openapidoc.Pathser); ok {
		var err error
		docs.Paths, err = paths.Paths(ctx, docs.BasePath)
		if err != nil {
			return nil, err
		}
	}

	err := k.addUrlInfo(ctx, &docs)
	if err != nil {
		return nil, err
	}

	docs.Swagger = "2.0"

	return &docs, nil
}

func (k *Kernel) addUrlInfo(ctx context.Context, docs *spec.Swagger) error {
	urlResolver, err := di.Resolve[router.URLResolver](ctx)
	if errors.Is(err, di.ErrNotRegistered) {
		return nil
	} else if err != nil {
		return err
	}

	u, err := url.Parse(urlResolver.Resolve("/"))
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	if u.Host == "" {
		return fmt.Errorf("no host in url %#v", u)
	}

	if docs.Schemes == nil {
		docs.Schemes = []string{u.Scheme}
	}
	if docs.Host == "" {
		docs.Host = u.Host
	}

	return nil
}
