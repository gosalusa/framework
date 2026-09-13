package request_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/request"
)

func TestHandlerRun(t *testing.T) {
	type runReq struct {
		Foo string
	}
	h := request.Handler(func(r *runReq) (*runReq, error) {
		r.Foo = "set"
		return r, nil
	})
	got, err := h.Run(&runReq{})
	assert.NoError(t, err)
	assert.Equal(t, &runReq{Foo: "set"}, got)
}

func TestHandlerValidate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		type req struct {
			Foo string `json:"foo"`
		}
		h := request.Handler(func(r *req) (any, error) { return nil, nil })
		assert.NoError(t, h.Validate(context.Background()))
	})

	t.Run("missing tags", func(t *testing.T) {
		type req struct {
			Foo string
		}
		h := request.Handler(func(r *req) (any, error) { return nil, nil })
		err := h.Validate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing tags")
	})

	t.Run("unregistered dependency", func(t *testing.T) {
		type req struct {
			Foo int `inject:""`
		}
		h := request.Handler(func(r *req) (any, error) { return nil, nil })
		err := h.Validate(context.Background())
		assert.Error(t, err)
	})

	t.Run("query and path tags", func(t *testing.T) {
		type req struct {
			A int `json:"a"`
			B int `query:"b"`
			C int `path:"c"`
		}
		h := request.Handler(func(r *req) (any, error) { return nil, nil })
		assert.NoError(t, h.Validate(context.Background()))
	})
}

func TestHandlerErrorResponses(t *testing.T) {
	t.Run("handler returns error", func(t *testing.T) {
		type req struct {
			A int `json:"a"`
		}
		h := request.Handler(func(r *req) (any, error) {
			return nil, errors.New("handler failed")
		})
		rw := httptest.NewRecorder()
		httpReq := httptest.NewRequest("GET", "/", nil)
		httpReq.Header.Set("Accept", "application/json")
		h.ServeHTTP(rw, httpReq)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		assert.Contains(t, rw.Body.String(), "handler failed")
	})

	t.Run("handler returns validation error", func(t *testing.T) {
		type req struct {
			A int `json:"a"`
		}
		h := request.Handler(func(r *req) (any, error) {
			return nil, request.ValidationError{"a": []string{"bad value"}}
		})
		rw := httptest.NewRecorder()
		httpReq := httptest.NewRequest("GET", "/", nil)
		httpReq.Header.Set("Accept", "application/json")
		h.ServeHTTP(rw, httpReq)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
		assert.Contains(t, rw.Body.String(), "bad value")
	})

	t.Run("run produces validation error", func(t *testing.T) {
		type req struct {
			A int `json:"a" validate:"min:5"`
		}
		h := request.Handler(func(r *req) (any, error) {
			return r, nil
		})
		rw := httptest.NewRecorder()
		httpReq := httptest.NewRequest("POST", "/", strings.NewReader(`{"a": 1}`))
		httpReq.Header.Set("Accept", "application/json")
		h.ServeHTTP(rw, httpReq)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	t.Run("run produces non-validation error", func(t *testing.T) {
		type req struct {
			A int `json:"a"`
		}
		h := request.Handler(func(r *req) (any, error) {
			return r, nil
		})
		rw := httptest.NewRecorder()
		httpReq := httptest.NewRequest("POST", "/", strings.NewReader(`not json`))
		httpReq.Header.Set("Accept", "application/json")
		h.ServeHTTP(rw, httpReq)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	t.Run("di fill error", func(t *testing.T) {
		type req struct {
			Foo int `inject:""`
		}
		h := request.Handler(func(r *req) (any, error) {
			return r, nil
		})
		rw := httptest.NewRecorder()
		httpReq := httptest.NewRequest("GET", "/", nil)
		httpReq.Header.Set("Accept", "application/json")
		h.ServeHTTP(rw, httpReq)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

type responderStub struct {
	status int
	body   string
	err    error
}

func (r *responderStub) Respond(w http.ResponseWriter, _ *http.Request) error {
	if r.err != nil {
		return r.err
	}
	w.WriteHeader(r.status)
	w.Write([]byte(r.body))
	return nil
}

func TestHandlerRespond(t *testing.T) {
	t.Run("http.Response", func(t *testing.T) {
		type req struct{}
		h := request.Handler(func(r *req) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     http.Header{"X-Test": []string{"yes"}},
				Body:       io.NopCloser(strings.NewReader("response body")),
			}, nil
		})
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, http.StatusCreated, rw.Code)
		assert.Equal(t, "yes", rw.Header().Get("X-Test"))
		assert.Equal(t, "response body", rw.Body.String())
	})

	t.Run("responder", func(t *testing.T) {
		type req struct{}
		h := request.Handler(func(r *req) (request.Responder, error) {
			return &responderStub{status: http.StatusAccepted, body: "responder body"}, nil
		})
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, http.StatusAccepted, rw.Code)
		assert.Equal(t, "responder body", rw.Body.String())
	})

	t.Run("http.Handler", func(t *testing.T) {
		type req struct{}
		h := request.Handler(func(r *req) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handler body"))
			}), nil
		})
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, "handler body", rw.Body.String())
	})

	t.Run("nil http.Response", func(t *testing.T) {
		type req struct{}
		h := request.Handler(func(r *req) (*http.Response, error) {
			return nil, nil
		})
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, http.StatusOK, rw.Code)
	})
}

func TestRespond(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rw := httptest.NewRecorder()
		request.Respond(rw, httptest.NewRequest("GET", "/", nil), &responderStub{status: http.StatusOK, body: "ok"})
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "ok", rw.Body.String())
	})

	t.Run("failing responder", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "application/json")
		request.Respond(rw, req, &responderStub{err: errors.New("respond failed")})
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		assert.Contains(t, rw.Body.String(), "respond failed")
	})
}

func TestRespondError(t *testing.T) {
	t.Run("nil error is a no-op", func(t *testing.T) {
		rw := httptest.NewRecorder()
		request.RespondError(rw, httptest.NewRequest("GET", "/", nil), nil)
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	t.Run("responds with 500", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "application/json")
		request.RespondError(rw, req, errors.New("boom"))
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		assert.Contains(t, rw.Body.String(), "boom")
	})
}
