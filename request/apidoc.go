package request

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-openapi/spec"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/openapidoc"
)

var typeModel = reflect.TypeOf((*model.Model)(nil)).Elem()

var _ openapidoc.Operationer = (*RequestHandler[any, any])(nil)

func (h *RequestHandler[TRequest, TResponse]) Docs(op *spec.OperationProps) *RequestHandler[TRequest, TResponse] {
	h.operation = op
	return h
}

func (h *RequestHandler[TRequest, TResponse]) Operation(ctx context.Context) (*spec.Operation, error) {
	var err error
	var op *spec.Operation
	if h.operation != nil {
		op = &spec.Operation{OperationProps: *h.operation}
	} else {
		op = spec.NewOperation("")
	}

	op.Parameters, err = newAPIRequest[TRequest]()
	if err != nil {
		return nil, err
	}

	if op.Responses == nil {
		op.Responses = &spec.Responses{}
	}
	if op.Responses.Default == nil {
		op.Responses.Default, err = openapidoc.Response(reflect.TypeFor[TResponse]())
		if err != nil {
			return nil, err
		}
	}

	if op.Produces == nil {
		ct, ok := openapidoc.GetContentType[TResponse]()
		if ok {
			op.Produces = []string{ct}
		} else if op.Responses.Default != nil {
			op.Produces = []string{"application/json"}
		}
	}

	return op, nil
}

func newAPIRequest[T any]() ([]spec.Parameter, error) {
	params := []spec.Parameter{}
	var emptyRequest *T
	t := reflect.TypeOf(emptyRequest).Elem()

	schema, err := openapidoc.Schema(t, true)
	if err != nil {
		return nil, fmt.Errorf("api request %s: %w", reflect.TypeFor[T](), err)
	}
	if len(schema.Properties) > 0 {
		params = append(params, *spec.BodyParam(t.Name(), schema))
	}
	for _, field := range helpers.GetFields(t) {
		if name, ok := field.Tag.Lookup("query"); ok {
			param, err := openapidoc.Param(field.Type, name, "query")
			if err != nil {
				return nil, err
			}
			params = append(params, *param)
		}

		if name, ok := field.Tag.Lookup("path"); ok {
			param, err := openapidoc.Param(field.Type, name, "path")
			if err != nil {
				return nil, err
			}
			params = append(params, *param)
		}

		if field.Type.Implements(typeModel) {
			if name, ok := field.Tag.Lookup("inject"); ok && name != "" {

				var pKey reflect.Type
				modelType := field.Type

				if modelType.Kind() == reflect.Pointer {
					modelType = modelType.Elem()
				}
				if modelType.Kind() == reflect.Struct {
					keys := helpers.RPrimaryKeyFields(modelType)
					if len(keys) == 1 {
						pKey = keys[0].Type
					}
				}
				param, err := openapidoc.Param(pKey, name, "path")
				if err != nil {
					return nil, err
				}
				params = append(params, *param)
			}
		}
	}

	return params, nil
}
