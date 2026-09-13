package view

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"gosalusa.com/di"
	"gosalusa.com/request"
	"gosalusa.com/router"
)

type ViewTemplate struct {
	fsys     fs.FS
	patterns []string
}

type ViewData struct {
	URL      router.URLResolver `inject:""`
	Template *ViewTemplate      `inject:""`
}
type ViewHandler struct {
	file string
	data any
}

var _ request.Responder = (*ViewHandler)(nil)

func (vh *ViewHandler) Execute(ctx context.Context, w io.Writer) error {
	d, err := di.Resolve[*ViewData](ctx)
	if err != nil {
		return fmt.Errorf("ViewHandler.Execute: %w", err)
	}
	return vh.ExecuteData(d, w)
}
func (vh *ViewHandler) ExecuteData(d *ViewData, w io.Writer) error {
	tpl, err := template.New("").
		Funcs(template.FuncMap{
			"route": d.URL.Resolve,
			"dd": func(v any) template.HTML {
				return template.HTML("<pre>" + template.HTMLEscapeString(spew.Sdump(v)) + "</pre>")
			},
		}).
		ParseFS(d.Template.fsys, d.Template.patterns...)
	if err != nil {
		return err
	}
	return tpl.ExecuteTemplate(w, vh.file, vh.data)
}

func (vh *ViewHandler) Respond(w http.ResponseWriter, r *http.Request) error {
	w.Header().Add("Content-Type", "text/html")
	return vh.Execute(r.Context(), w)
}
func (vh *ViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := vh.Respond(w, r)
	if err != nil {
		logger, resolveErr := di.Resolve[*slog.Logger](r.Context())
		if resolveErr != nil {
			logger = slog.Default()
		}
		logger.Error("failed to serve view", "error", err)
	}
}

func (vh *ViewHandler) Bytes(ctx context.Context) ([]byte, error) {
	b := &bytes.Buffer{}
	err := vh.Execute(ctx, b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func (vh *ViewHandler) BytesData(d *ViewData) ([]byte, error) {
	b := &bytes.Buffer{}
	err := vh.ExecuteData(d, b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func View(file string, data any) *ViewHandler {
	return &ViewHandler{
		file: file,
		data: data,
	}

}
func NewViewTemplate(fsys fs.FS, patterns ...string) *ViewTemplate {
	if len(patterns) == 0 {
		patterns = []string{"**/*.html"}
	}
	return &ViewTemplate{
		fsys:     fsys,
		patterns: patterns,
	}
}
func Register(fsys fs.FS, patterns ...string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		di.RegisterSingleton(ctx, func() *ViewTemplate {
			return NewViewTemplate(fsys, patterns...)
		})
		return nil
	}
}
