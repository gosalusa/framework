package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/go-openapi/spec"
	"gosalusa.com/di"
	"gosalusa.com/internal/helpers"
	"gosalusa.com/validate"
)

type RequestHandler[TRequest, TResponse any] struct {
	handler   func(r *TRequest) (TResponse, error)
	operation *spec.OperationProps
}

func (h *RequestHandler[TRequest, TResponse]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.serveHTTP(w, r)
	if err != nil {
		RespondError(w, r, err)
		return
	}
}
func (h *RequestHandler[TRequest, TResponse]) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	var req TRequest
	err := Run(r, &req)
	if validationErr, ok := err.(ValidationError); ok {
		return NewHTTPError(validationErr, http.StatusUnprocessableEntity)
	} else if err != nil {
		return err
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, requestKey, r)
	ctx = context.WithValue(ctx, responseKey, w)

	if di.IsFillable(&req) {
		err = di.Fill(ctx, &req)
		if err != nil {
			return fmt.Errorf("failed to di.Fill request: %w", err)
		}
	}

	resp, err := h.handler(&req)
	if validationErr, ok := err.(ValidationError); ok {
		return NewHTTPError(validationErr, http.StatusUnprocessableEntity)
	} else if err != nil {
		return err
	}

	return respond(w, r, resp)
}

func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	addError(r, err)
	if !hasHandleErrors(r) {
		ErrorHandler(err).ServeHTTP(w, r)
	}
}

func Respond(w http.ResponseWriter, r *http.Request, resp any) {
	err := respond(w, r, resp)
	if err != nil {
		RespondError(w, r, err)
		return
	}
}
func respond(w http.ResponseWriter, r *http.Request, resp any) error {
	switch resp := resp.(type) {
	case *http.Response:
		return serveResponse(w, resp)
	case Responder:
		return resp.Respond(w, r)
	case http.Handler:
		resp.ServeHTTP(w, r)
		return nil
	default:
		return NewJSONResponse(resp).Respond(w, r)
	}
}

func (h *RequestHandler[TRequest, TResponse]) Run(r *TRequest) (TResponse, error) {
	return h.handler(r)
}

func Handler[TRequest, TResponse any](callback func(r *TRequest) (TResponse, error)) *RequestHandler[TRequest, TResponse] {
	return &RequestHandler[TRequest, TResponse]{
		handler: callback,
	}
}

func (r *RequestHandler[TRequest, TResponse]) Validate(ctx context.Context) error {
	t := reflect.TypeFor[TRequest]()
	errs := []error{}
	for _, sf := range helpers.GetFields(t) {
		if isMissingTags(sf) {
			errs = append(errs, fmt.Errorf("%s.%s missing tags", t, sf.Name))
		}
	}
	return validate.Append(ctx, errors.Join(errs...), di.Validator(ctx, t))
}

func isMissingTags(sf reflect.StructField) bool {
	if _, ok := sf.Tag.Lookup("json"); ok {
		return false
	}
	if _, ok := sf.Tag.Lookup("query"); ok {
		return false
	}
	if _, ok := sf.Tag.Lookup("path"); ok {
		return false
	}
	if _, ok := sf.Tag.Lookup("inject"); ok {
		return false
	}
	return true
}
func serveResponse(w http.ResponseWriter, r *http.Response) error {
	if r == nil {
		return nil
	}
	defer r.Body.Close()

	for k, vs := range r.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if r.StatusCode != 0 {
		w.WriteHeader(r.StatusCode)
	}

	_, err := io.Copy(w, r.Body)
	return err
}
