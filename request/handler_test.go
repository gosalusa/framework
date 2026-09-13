package request_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/request"
)

func ExampleHandler_input() {
	type ExampleRequest struct {
		A int `query:"a"`
		B int `query:"b"`
	}
	type ExampleResponse struct {
		Sum int `json:"sum"`
	}
	h := request.Handler(func(r *ExampleRequest) (*ExampleResponse, error) {
		return &ExampleResponse{
			Sum: r.A + r.B,
		}, nil
	})

	rw := httptest.NewRecorder()

	h.ServeHTTP(
		rw,
		httptest.NewRequest("GET", "/?a=10&b=5", http.NoBody),
	)

	fmt.Println(rw.Body)
	// Output:
	// {
	//     "sum": 15
	// }
}

func ExampleHandler_error() {
	type ExampleRequest struct {
		A int `query:"a" validate:"min:1"`
	}
	type ExampleResponse struct {
	}

	h := request.Handler(func(r *ExampleRequest) (*ExampleResponse, error) {
		return &ExampleResponse{}, nil
	})

	rw := httptest.NewRecorder()

	h.ServeHTTP(
		rw,
		httptest.NewRequest("GET", "/?a=-1", http.NoBody),
	)

	fmt.Println(rw.Result().StatusCode)
	// Output: 422
}

func ExampleHandler_pathParams() {
	type ExampleRequest struct {
		A string `path:"a"`
		B string `path:"b"`
	}
	r := mux.NewRouter()
	r.Handle("/{a}/{b}", request.Handler(func(r *ExampleRequest) (*ExampleRequest, error) {
		return r, nil
	}))

	rw := httptest.NewRecorder()

	r.ServeHTTP(
		rw,
		httptest.NewRequest("GET", "/path_param_a/path_param_b", http.NoBody),
	)

	fmt.Println(rw.Body)
	// Output:
	// {
	//     "A": "path_param_a",
	//     "B": "path_param_b"
	// }
}

func TestHandler(t *testing.T) {
	t.Run("path params", func(t *testing.T) {
		type ExampleRequest struct {
			A string `path:"a"`
			B string `path:"b"`
		}

		resp := run[ExampleRequest](t, "/{a}/{b}", "GET", "/path_param_a/path_param_b", http.NoBody)

		assert.Equal(t, &ExampleRequest{
			A: "path_param_a",
			B: "path_param_b",
		}, resp)
	})

	t.Run("path params ignore without tag", func(t *testing.T) {
		type ExampleRequest struct {
			A string
			B string `path:"b"`
		}

		resp := run[ExampleRequest](t, "/{A}/{b}", "GET", "/path_param_a/path_param_b", http.NoBody)

		assert.Equal(t, &ExampleRequest{
			A: "",
			B: "path_param_b",
		}, resp)
	})

	t.Run("query params", func(t *testing.T) {
		type ExampleRequest struct {
			A string `query:"a"`
		}

		resp := run[ExampleRequest](t, "/", "GET", "/?a=query_param", http.NoBody)

		assert.Equal(t, &ExampleRequest{
			A: "query_param",
		}, resp)
	})

	t.Run("query params ignore untagged", func(t *testing.T) {
		type ExampleRequest struct {
			A string `query:"a"`
			B string
		}

		resp := run[ExampleRequest](t, "/", "GET", "/?a=query_param_a&B=query_param_b", http.NoBody)

		assert.Equal(t, &ExampleRequest{
			A: "query_param_a",
			B: "",
		}, resp)
	})
}

func run[T any](t *testing.T, path, method, url string, body io.Reader) *T {

	r := mux.NewRouter()
	r.Handle(path, request.Handler(func(r *T) (*T, error) {
		return r, nil
	}))

	rw := httptest.NewRecorder()

	r.ServeHTTP(
		rw,
		httptest.NewRequest(method, url, body),
	)

	var resp T
	err := json.Unmarshal(rw.Body.Bytes(), &resp)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return &resp
}
