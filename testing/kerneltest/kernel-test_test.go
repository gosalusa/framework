package kerneltest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"gosalusa.com/kernel"
	"gosalusa.com/testing/kerneltest"
)

type testConfig struct{}

func (testConfig) GetHTTPPort() int   { return 8080 }
func (testConfig) GetBaseURL() string { return "https://example.com" }

func newKernel() *kernel.Kernel {
	return kernel.New(
		kernel.RootHandler(func(ctx context.Context) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
		}),
	)
}

func TestTestKernel(t *testing.T) {
	factory := kerneltest.NewTestKernelFactory(newKernel(), testConfig{})
	tk := factory(t)

	t.Run("Get", func(t *testing.T) {
		tk.Get("/").AssertStatus(http.StatusNoContent)
	})
	t.Run("GetJSON", func(t *testing.T) {
		tk.GetJSON("/").AssertStatus(http.StatusNoContent)
	})
	t.Run("Post", func(t *testing.T) {
		tk.Post("/", strings.NewReader("body")).AssertStatus(http.StatusNoContent)
	})
	t.Run("PostJSON", func(t *testing.T) {
		tk.PostJSON("/", map[string]any{"a": 1}).AssertStatus(http.StatusNoContent)
	})
	t.Run("Put", func(t *testing.T) {
		tk.Put("/", strings.NewReader("body")).AssertStatus(http.StatusNoContent)
	})
	t.Run("PutJSON", func(t *testing.T) {
		tk.PutJSON("/", map[string]any{"a": 1}).AssertStatus(http.StatusNoContent)
	})
	t.Run("Patch", func(t *testing.T) {
		tk.Patch("/", strings.NewReader("body")).AssertStatus(http.StatusNoContent)
	})
	t.Run("PatchJSON", func(t *testing.T) {
		tk.PatchJSON("/", map[string]any{"a": 1}).AssertStatus(http.StatusNoContent)
	})
	t.Run("Delete", func(t *testing.T) {
		tk.Delete("/", strings.NewReader("body")).AssertStatus(http.StatusNoContent)
	})
	t.Run("DeleteJSON", func(t *testing.T) {
		tk.DeleteJSON("/", map[string]any{"a": 1}).AssertStatus(http.StatusNoContent)
	})
}
