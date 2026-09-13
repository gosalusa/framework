package handlertest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/testing/handlertest"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/created":
		w.WriteHeader(http.StatusCreated)
	case "/redirect":
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
	case "/error":
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"foo":{"bar":1},"list":[1,2]}`)
	}
}
func TestHandlerTest(t *testing.T) {
	responseBody := `{"foo":{"bar":1}}`
	ctx := context.Background()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, responseBody)
		if err != nil {
			panic(err)
		}
	})

	handlertest.New(ctx, t, h).
		Get("/test").
		AssertJSONString(responseBody).
		AssertJSONContains("foo.bar", 1.0)
}

func TestRequestBuilderMethods(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		handlertest.New(context.Background(), t, http.HandlerFunc(testHandler)).
			Get("/x").
			AssertStatus(200).
			AssertStatus2XX().
			AssertJSONString(`{"foo":{"bar":1},"list":[1,2]}`)
	})

	t.Run("get json headers", func(t *testing.T) {
		var seenAccept, seenContentType string
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenAccept = r.Header.Get("Accept")
			seenContentType = r.Header.Get("Content-Type")
			w.WriteHeader(http.StatusNoContent)
		})
		handlertest.New(context.Background(), t, h).GetJSON("/x")
		assert.Equal(t, "application/json", seenAccept)
		assert.Equal(t, "application/json", seenContentType)
	})

	t.Run("post", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).
			Post("/x", nil).
			AssertStatusOK()
	})

	t.Run("post json", func(t *testing.T) {
		var body map[string]any
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).PostJSON("/x", map[string]any{"a": 1})
		assert.Equal(t, map[string]any{"a": float64(1)}, body)
	})

	t.Run("put", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).
			Put("/x", nil).
			AssertStatus2XX()
	})

	t.Run("put json", func(t *testing.T) {
		var body map[string]any
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).PutJSON("/x", map[string]any{"a": 1})
		assert.Equal(t, map[string]any{"a": float64(1)}, body)
	})

	t.Run("delete", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).
			Delete("/x", nil).
			AssertStatus2XX()
	})

	t.Run("delete json", func(t *testing.T) {
		var body map[string]any
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).DeleteJSON("/x", map[string]any{"a": 1})
		assert.Equal(t, map[string]any{"a": float64(1)}, body)
	})

	t.Run("patch", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).
			Patch("/x", nil).
			AssertStatus2XX()
	})

	t.Run("patch json", func(t *testing.T) {
		var body map[string]any
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).PatchJSON("/x", map[string]any{"a": 1})
		assert.Equal(t, map[string]any{"a": float64(1)}, body)
	})
}

func TestAssertions(t *testing.T) {
	t.Run("status ranges", func(t *testing.T) {
		handlertest.New(context.Background(), t, http.HandlerFunc(testHandler)).
			Get("/created").
			AssertStatus(201).
			AssertStatusRange(200, 299).
			AssertStatusOK().
			AssertStatus2XX()

		handlertest.New(context.Background(), t, http.HandlerFunc(testHandler)).
			Get("/redirect").
			AssertStatus3XX()

		handlertest.New(context.Background(), t, http.HandlerFunc(testHandler)).
			Get("/error").
			AssertStatus5XX()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
		handlertest.New(context.Background(), t, h).
			Get("/teapot").
			AssertStatus4XX()
	})

	t.Run("json contains", func(t *testing.T) {
		handlertest.New(context.Background(), t, http.HandlerFunc(testHandler)).
			Get("/x").
			AssertJSONContains("foo.bar", 1.0)
	})

	t.Run("with header", func(t *testing.T) {
		var seen string
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = r.Header.Get("X-Test")
			w.WriteHeader(http.StatusOK)
		})
		handlertest.New(context.Background(), t, h).
			WithHeader("X-Test", "value").
			Get("/x")
		assert.Equal(t, "value", seen)
	})
}
