package request

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"gosalusa.com/router"
)

type errorContextType uint8

const errorContextKey = errorContextType(iota)

type errorsContainer struct {
	errors []error
}

func (e *errorsContainer) add(err error) {
	e.errors = append(e.errors, err)
}
func (w *errorsContainer) joinedError() error {
	if len(w.errors) == 0 {
		return nil
	}
	if len(w.errors) == 1 {
		return w.errors[0]
	}
	return errors.Join(w.errors...)
}

func addError(r *http.Request, err error) {
	e := r.Context().Value(errorContextKey)
	if e == nil {
		return
	}
	e.(*errorsContainer).add(err)
}

func hasHandleErrors(r *http.Request) bool {
	e := r.Context().Value(errorContextKey)
	return e != nil
}

func HandleErrors(handlers ...func(ctx context.Context, err error) http.Handler) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errContainer := &errorsContainer{errors: []error{}}

			defer func() {
				err := toError(recover())
				if err != nil {
					var httpErr *HTTPError
					if !errors.As(err, &httpErr) {
						httpErr = NewHTTPError(err, http.StatusInternalServerError)
					}
					httpErr.WithStack()

					errContainer.add(httpErr)
				}

				err = errContainer.joinedError()
				if err == nil {
					return
				}

				var errorResponse http.Handler

				for _, handler := range handlers {
					h := handler(r.Context(), err)
					if h != nil {
						errorResponse = h
					}
				}
				if errorResponse == nil {
					errorResponse = ErrorHandler(err)
				}

				errorResponse.ServeHTTP(w, r)
			}()

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), errorContextKey, errContainer)))
		})
	}
}

func toError(e any) error {
	if e == nil {
		return nil
	}
	switch e := e.(type) {
	case error:
		return e
	case string:
		return errors.New(e)
	default:
		return errors.New(fmt.Sprint(e))
	}
}
