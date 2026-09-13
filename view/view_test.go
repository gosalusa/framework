package view_test

import (
	"bytes"
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/di"
	"gosalusa.com/router"
	"gosalusa.com/view"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte(`<html>{{ route "home" }}</html>`),
		},
		"dist/user.html": &fstest.MapFile{
			Data: []byte(`<html>hello {{ .Name }} {{ route "home" }}</html>`),
		},
		"dist/dump.html": &fstest.MapFile{
			Data: []byte(`<html>{{ dd . }}</html>`),
		},
		"dist/missing-template.html": &fstest.MapFile{},
	}
}

func newTestContext(t *testing.T, fsys fs.FS) context.Context {
	t.Helper()
	ctx := di.TestDependencyProviderContext()
	di.Register(ctx, func(ctx context.Context, tag string) (router.URLResolver, error) {
		return router.NewTestResolver(), nil
	})
	di.RegisterSingleton(ctx, func() *view.ViewTemplate {
		return view.NewViewTemplate(fsys)
	})
	return ctx
}

func TestNewViewTemplate(t *testing.T) {
	t.Run("default pattern", func(t *testing.T) {
		tpl := view.NewViewTemplate(testFS())
		require.NotNil(t, tpl)
	})
	t.Run("with patterns", func(t *testing.T) {
		tpl := view.NewViewTemplate(testFS(), "*.html")
		require.NotNil(t, tpl)
	})
}

func TestView(t *testing.T) {
	vh := view.View("index.html", nil)
	require.NotNil(t, vh)
}

func TestRegister(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	register := view.Register(testFS())
	require.NotNil(t, register)
	err := register(ctx)
	require.NoError(t, err)

	tpl, err := di.Resolve[*view.ViewTemplate](ctx)
	require.NoError(t, err)
	require.NotNil(t, tpl)
}

func TestViewHandler_Execute(t *testing.T) {
	t.Run("renders template", func(t *testing.T) {
		ctx := newTestContext(t, testFS())
		vh := view.View("index.html", nil)

		buf := &bytes.Buffer{}
		err := vh.Execute(ctx, buf)
		require.NoError(t, err)
		assert.Equal(t, `<html>home?</html>`, buf.String())
	})
	t.Run("view data not registered", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		vh := view.View("index.html", nil)

		buf := &bytes.Buffer{}
		err := vh.Execute(ctx, buf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ViewHandler.Execute")
	})
}

func TestViewHandler_ExecuteData(t *testing.T) {
	ctx := newTestContext(t, testFS())
	d, err := di.Resolve[*view.ViewData](ctx)
	require.NoError(t, err)

	t.Run("renders with data", func(t *testing.T) {
		vh := view.View("user.html", map[string]string{"Name": "world"})

		buf := &bytes.Buffer{}
		err := vh.ExecuteData(d, buf)
		require.NoError(t, err)
		assert.Equal(t, `<html>hello world home?</html>`, buf.String())
	})
	t.Run("renders with dd function", func(t *testing.T) {
		vh := view.View("dump.html", map[string]string{"Name": "world"})

		buf := &bytes.Buffer{}
		err := vh.ExecuteData(d, buf)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "<pre>")
	})
	t.Run("template missing", func(t *testing.T) {
		vh := view.View("does-not-exist.html", nil)

		buf := &bytes.Buffer{}
		err := vh.ExecuteData(d, buf)
		require.Error(t, err)
	})
}

func TestViewHandler_Bytes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := newTestContext(t, testFS())
		vh := view.View("index.html", nil)

		b, err := vh.Bytes(ctx)
		require.NoError(t, err)
		assert.Equal(t, `<html>home?</html>`, string(b))
	})
	t.Run("error", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		vh := view.View("index.html", nil)

		b, err := vh.Bytes(ctx)
		require.Error(t, err)
		assert.Nil(t, b)
	})
}

func TestViewHandler_BytesData(t *testing.T) {
	ctx := newTestContext(t, testFS())
	d, err := di.Resolve[*view.ViewData](ctx)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		vh := view.View("index.html", nil)

		b, err := vh.BytesData(d)
		require.NoError(t, err)
		assert.Equal(t, `<html>home?</html>`, string(b))
	})
	t.Run("error", func(t *testing.T) {
		vh := view.View("missing.html", nil)

		b, err := vh.BytesData(d)
		require.Error(t, err)
		assert.Nil(t, b)
	})
}

func TestViewHandler_Respond(t *testing.T) {
	ctx := newTestContext(t, testFS())
	req := httptest.NewRequest("GET", "http://example.com/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	vh := view.View("index.html", nil)
	err := vh.Respond(rec, req)
	require.NoError(t, err)

	assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
	assert.Equal(t, `<html>home?</html>`, rec.Body.String())
}

func TestViewHandler_ServeHTTP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := newTestContext(t, testFS())
		req := httptest.NewRequest("GET", "http://example.com/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		vh := view.View("index.html", nil)
		vh.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `<html>home?</html>`, rec.Body.String())
	})
	t.Run("rendering error logs", func(t *testing.T) {
		ctx := newTestContext(t, testFS())
		req := httptest.NewRequest("GET", "http://example.com/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		vh := view.View("missing.html", nil)
		vh.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("rendering error uses default logger", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		req := httptest.NewRequest("GET", "http://example.com/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		vh := view.View("missing.html", nil)
		vh.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("rendering error uses registered logger", func(t *testing.T) {
		ctx := di.TestDependencyProviderContext()
		di.RegisterSingleton(ctx, func() *slog.Logger {
			return slog.Default()
		})
		req := httptest.NewRequest("GET", "http://example.com/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		vh := view.View("missing.html", nil)
		vh.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
