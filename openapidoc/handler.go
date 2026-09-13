package openapidoc

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gosalusa.com/di"
)

//go:generate sh -c "curl https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js | gzip > redoc.standalone.js.gz"

//go:embed index.html
var redocIndex []byte

//go:embed redoc.standalone.js.gz
var redocSrc []byte

func SwaggerUI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/swagger.json") {
			serveSwagger(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/redoc.standalone.js") {
			serveBytes(w, r, redocSrc, "gz/application/json")
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.RequestURI+"/", http.StatusFound)
			return
		}
		serveBytes(w, r, redocIndex, "text/html")
	})
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	api, err := di.Resolve[APIDocer](r.Context())
	if err != nil {
		panic(err)
	}
	doc, err := api.APIDoc(r.Context())
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json")
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	err = e.Encode(doc)
	if err != nil {
		panic(err)
	}
}

func serveBytes(w http.ResponseWriter, _ *http.Request, b []byte, contentType string) {
	gzPrefix := "gz/"
	if strings.HasPrefix(contentType, gzPrefix) {
		contentType = contentType[len(gzPrefix):]
		w.Header().Add("Content-Encoding", "gzip")
	}
	w.Header().Add("Content-Type", contentType)
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(b)), 10))
	_, err := w.Write(b)
	if err != nil {
		panic(err)
	}
}
