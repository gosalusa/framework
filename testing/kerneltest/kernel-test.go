package kerneltest

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/kernel"
	"gosalusa.com/salusaconfig"
	"gosalusa.com/testing/handlertest"
)

type TestKernel[T salusaconfig.Config] struct {
	kernel *kernel.Kernel
	config T
	ctx    context.Context
	t      *testing.T
}

func NewTestKernelFactory[T salusaconfig.Config](k *kernel.Kernel, cfg T) func(t *testing.T) *TestKernel[T] {
	return func(t *testing.T) *TestKernel[T] {
		ctx := di.TestDependencyProviderContext()
		k := kernel.Config(func() T { return cfg })(k)
		err := k.Bootstrap(ctx)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return &TestKernel[T]{
			kernel: k,
			config: cfg,
			ctx:    ctx,
			t:      t,
		}
	}
}

func (k *TestKernel[T]) Request() *handlertest.RequestBuilder {
	return handlertest.New(k.ctx, k.t, k.kernel.RootHandler())
}

func (k *TestKernel[T]) Get(target string) *handlertest.HttpResult {
	return k.Request().Get(target)
}
func (k *TestKernel[T]) GetJSON(target string) *handlertest.HttpResult {
	return k.Request().GetJSON(target)
}
func (k *TestKernel[T]) Post(target string, body io.Reader) *handlertest.HttpResult {
	return k.Request().Post(target, body)
}
func (k *TestKernel[T]) PostJSON(target string, body any) *handlertest.HttpResult {
	return k.Request().PostJSON(target, body)
}
func (k *TestKernel[T]) Put(target string, body io.Reader) *handlertest.HttpResult {
	return k.Request().Put(target, body)
}
func (k *TestKernel[T]) PutJSON(target string, body any) *handlertest.HttpResult {
	return k.Request().PutJSON(target, body)
}
func (k *TestKernel[T]) Patch(target string, body io.Reader) *handlertest.HttpResult {
	return k.Request().Patch(target, body)
}
func (k *TestKernel[T]) PatchJSON(target string, body any) *handlertest.HttpResult {
	return k.Request().PatchJSON(target, body)
}
func (k *TestKernel[T]) Delete(target string, body io.Reader) *handlertest.HttpResult {
	return k.Request().Delete(target, body)
}
func (k *TestKernel[T]) DeleteJSON(target string, body any) *handlertest.HttpResult {
	return k.Request().DeleteJSON(target, body)
}
