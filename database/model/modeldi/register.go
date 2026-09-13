package modeldi

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/di"
	"gosalusa.com/request"
)

type modelDeps struct {
	Request *http.Request `inject:""`
	DB      *sqlx.DB      `inject:""`
}

// Register registers a dependency-injection provider for the model type T. The
// provider loads the model by an ID resolved from the current request's query
// parameters, gorilla/mux variables, or path values, matched against the inject
// tag of the field being filled. It returns request.ErrStatusNotFound when no
// matching row exists, and an error when the ID is not present in the request.
func Register[T model.Model](ctx context.Context) {
	di.RegisterWith(ctx, func(ctx context.Context, tag string, deps *modelDeps) (T, error) {
		v, ok := getValue(deps.Request, tag)
		if !ok {
			var zero T
			return zero, fmt.Errorf("could not fetch model: %s not in request", tag)
		}

		u, err := builder.From[T]().WithContext(ctx).Find(deps.DB, v)
		if err != nil {
			var zero T
			return zero, err
		}
		if reflect.ValueOf(u).IsZero() {
			var zero T
			return zero, request.ErrStatusNotFound
		}
		return u, nil
	})
}

func getValue(r *http.Request, tag string) (string, bool) {
	if r.URL.Query().Has(tag) {
		return r.URL.Query().Get(tag), true
	}

	vars := mux.Vars(r)
	v, ok := vars[tag]
	if ok {
		return v, true
	}

	v = r.PathValue(tag)
	if v != "" {
		return v, true
	}

	return "", false
}
