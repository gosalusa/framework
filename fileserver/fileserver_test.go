package fileserver_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/fileserver"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<html>index {{ .Message }}</html>`),
		},
		"style.css": &fstest.MapFile{
			Data: []byte("body { color: red; }"),
		},
		"dir": &fstest.MapFile{
			Mode: fs.ModeDir,
		},
	}
}

func TestWithFallback(t *testing.T) {
	t.Run("serves static file", func(t *testing.T) {
		h := fileserver.WithFallback(testFS(), "", "index.html", nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/style.css", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/css")
		assert.Equal(t, "body { color: red; }", rec.Body.String())
	})
	t.Run("falls back to index when file missing", func(t *testing.T) {
		h := fileserver.WithFallback(testFS(), "", "index.html", map[string]string{"Message": "hello"})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/not-found", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
		assert.Equal(t, `<html>index hello</html>`, rec.Body.String())
	})
	t.Run("falls back to index when directory", func(t *testing.T) {
		h := fileserver.WithFallback(testFS(), "", "index.html", nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `<html>index <no value></html>`, rec.Body.String())
	})
	t.Run("base path is joined", func(t *testing.T) {
		fsys := fstest.MapFS{
			"public/index.html": &fstest.MapFile{
				Data: []byte(`<html>nested</html>`),
			},
		}
		h := fileserver.WithFallback(fsys, "public", "index.html", nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/missing", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `<html>nested</html>`, rec.Body.String())
	})
	t.Run("fallback template missing", func(t *testing.T) {
		fsys := fstest.MapFS{
			"other.html": &fstest.MapFile{
				Data: []byte(`<html>other</html>`),
			},
		}
		h := fileserver.WithFallback(fsys, "", "index.html", nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/missing", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())
	})
	t.Run("fallback execute error", func(t *testing.T) {
		fsys := fstest.MapFS{
			"index.html": &fstest.MapFile{
				Data: []byte(`<html>{{ .Invalid }}</html>`),
			},
		}
		h := fileserver.WithFallback(fsys, "", "index.html", 42)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/missing", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEqual(t, "<html><no value></html>", rec.Body.String())
	})
	t.Run("root open error", func(t *testing.T) {
		fsys := fstest.MapFS{
			"broken": &fstest.MapFile{
				Data: []byte("x"),
			},
			"index.html": &fstest.MapFile{
				Data: []byte(`<html>index</html>`),
			},
		}
		h := fileserver.WithFallback(&openFailFS{FS: fsys, failAfter: 1}, "", "index.html", nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://example.com/broken", nil)

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())
	})
}

type openFailFS struct {
	fs.FS
	failAfter int
	calls     int
}

func (o *openFailFS) Open(name string) (fs.File, error) {
	if name == "broken" {
		o.calls++
		if o.calls > o.failAfter {
			return nil, fs.ErrPermission
		}
	}
	return o.FS.Open(name)
}
