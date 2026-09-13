package request

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"gosalusa.com/clog"
)

type HTMLError interface {
	HTMLError() string
}

//go:embed error.html
var errorTemplate string

type errResponse struct {
	Error      template.HTML   `json:"error"`
	Status     int             `json:"status"`
	StatusText string          `json:"-"`
	StackTrace *StackTrace     `json:"stack,omitempty"`
	Fields     ValidationError `json:"fields,omitempty"`
}

type StackTrace struct {
	GoRoutine string        `json:"go_routine"`
	Stack     []*StackFrame `json:"stack"`
}

type StackFrame struct {
	Call  string `json:"call"`
	File  string `json:"file"`
	Line  int    `json:"line"`
	Extra int    `json:"-"`
}

type HTTPError struct {
	err    error
	status int
	stack  []byte
}

type StatusError int

func (e StatusError) Error() string {
	return fmt.Sprintf("http %d: %s", e, http.StatusText(int(e)))
}

func (e StatusError) Respond(w http.ResponseWriter, r *http.Request) error {
	return NewHTTPError(e, int(e)).Respond(w, r)
}

const (
	ErrStatusBadRequest                   = StatusError(http.StatusBadRequest)
	ErrStatusUnauthorized                 = StatusError(http.StatusUnauthorized)
	ErrStatusPaymentRequired              = StatusError(http.StatusPaymentRequired)
	ErrStatusForbidden                    = StatusError(http.StatusForbidden)
	ErrStatusNotFound                     = StatusError(http.StatusNotFound)
	ErrStatusMethodNotAllowed             = StatusError(http.StatusMethodNotAllowed)
	ErrStatusNotAcceptable                = StatusError(http.StatusNotAcceptable)
	ErrStatusProxyAuthRequired            = StatusError(http.StatusProxyAuthRequired)
	ErrStatusRequestTimeout               = StatusError(http.StatusRequestTimeout)
	ErrStatusConflict                     = StatusError(http.StatusConflict)
	ErrStatusGone                         = StatusError(http.StatusGone)
	ErrStatusLengthRequired               = StatusError(http.StatusLengthRequired)
	ErrStatusPreconditionFailed           = StatusError(http.StatusPreconditionFailed)
	ErrStatusRequestEntityTooLarge        = StatusError(http.StatusRequestEntityTooLarge)
	ErrStatusRequestURITooLong            = StatusError(http.StatusRequestURITooLong)
	ErrStatusUnsupportedMediaType         = StatusError(http.StatusUnsupportedMediaType)
	ErrStatusRequestedRangeNotSatisfiable = StatusError(http.StatusRequestedRangeNotSatisfiable)
	ErrStatusExpectationFailed            = StatusError(http.StatusExpectationFailed)
	ErrStatusTeapot                       = StatusError(http.StatusTeapot)
	ErrStatusMisdirectedRequest           = StatusError(http.StatusMisdirectedRequest)
	ErrStatusUnprocessableEntity          = StatusError(http.StatusUnprocessableEntity)
	ErrStatusLocked                       = StatusError(http.StatusLocked)
	ErrStatusFailedDependency             = StatusError(http.StatusFailedDependency)
	ErrStatusTooEarly                     = StatusError(http.StatusTooEarly)
	ErrStatusUpgradeRequired              = StatusError(http.StatusUpgradeRequired)
	ErrStatusPreconditionRequired         = StatusError(http.StatusPreconditionRequired)
	ErrStatusTooManyRequests              = StatusError(http.StatusTooManyRequests)
	ErrStatusRequestHeaderFieldsTooLarge  = StatusError(http.StatusRequestHeaderFieldsTooLarge)
	ErrStatusUnavailableForLegalReasons   = StatusError(http.StatusUnavailableForLegalReasons)
)

var _ error = &HTTPError{}
var _ Responder = &HTTPError{}

func NewHTTPError(err error, status int) *HTTPError {
	return &HTTPError{
		err:    err,
		status: status,
	}
}
func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d: %v", e.status, e.err)
}
func (e *HTTPError) Unwrap() error {
	return e.err
}

func (e *HTTPError) WithStack() {
	e.stack = debug.Stack()
}

func (e *HTTPError) Status() int {
	return e.status
}

func (e *HTTPError) Respond(w http.ResponseWriter, r *http.Request) error {

	response := errResponse{
		Error:      template.HTML(e.err.Error()),
		Status:     e.status,
		StatusText: http.StatusText(e.status),
	}
	if validationErr, ok := e.err.(ValidationError); ok {
		response.Fields = validationErr
	}

	if e.status == 500 && e.stack != nil {
		response.StackTrace = parseStack(e.stack)
	}

	if strings.HasPrefix(r.Header.Get("Accept"), "application/json") {
		return NewJSONResponse(response).SetStatus(e.status).Respond(w, r)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	t := template.New("error").Funcs(template.FuncMap{
		"jsonEncode": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return err.Error()
			}
			return string(b)
		},
		"isLocal": func(s string) bool {
			return strings.HasPrefix(s, cwd)
		},
		"getSrc": func(s *StackFrame) string {
			b, err := os.ReadFile(s.File)
			if err != nil {
				return "Error: " + err.Error()
			}
			lines := bytes.Split(b, []byte("\n"))

			out := []byte{}

			around := 4

			for i := max(s.Line-around-1, 0); i < min(len(lines), s.Line+around); i++ {
				out = fmt.Appendf(out, "%3d %s\n", i+1, lines[i])
			}

			return string(out)
		},
	})
	t, err = t.Parse(errorTemplate)
	if err != nil {
		return err
	}

	if err, ok := e.err.(HTMLError); ok {
		response.Error = template.HTML(err.HTMLError())
	} else {
		response.Error = template.HTML("<h2>" + e.err.Error() + "</h2>")
	}

	reader, writer := io.Pipe()
	go func() {
		err := t.Execute(writer, response)
		err = writer.CloseWithError(err)
		if err != nil {
			clog.Use(r.Context()).Warn("failed to close error response", "err", err)
		}
	}()
	return NewResponse(reader).SetStatus(e.status).AddHeader("Content-Type", "text/html").Respond(w, r)
}

func parseStack(stack []byte) *StackTrace {
	lines := bytes.Split(stack, []byte("\n"))
	goRoutine := string(lines[0])

	frames := []*StackFrame{}

	hasPanicked := false
	for i := 1; i < len(lines); i += 2 {
		if len(lines) <= i+1 {
			break
		}
		call := string(lines[i])
		pathLine := string(lines[i+1])
		i := strings.LastIndex(pathLine, ":")
		// pathLine := strings.SplitN(string(lines[i+1]), ":", 2)
		file := strings.TrimSpace(pathLine[:i])
		pathLineEnd := strings.SplitN(pathLine[i+1:], " ", 2)

		line, err := strconv.Atoi(pathLineEnd[0])
		if err != nil {
			line = -1
		}
		extra := -1
		if len(pathLineEnd) > 1 {
			c, err := parsInt(pathLineEnd[1])
			if err == nil {
				extra = c
			}
		}
		if hasPanicked {
			frames = append(frames, &StackFrame{
				Call:  call,
				File:  file,
				Line:  line,
				Extra: extra,
			})
		} else if strings.HasPrefix(call, "panic(") {
			hasPanicked = true
		}
	}
	return &StackTrace{
		GoRoutine: goRoutine,
		Stack:     frames,
	}
}

func ErrorHandler(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder, ok := getResponder(err)
		if !ok {
			responder = NewHTTPError(err, http.StatusInternalServerError)
		}

		err = responder.Respond(w, r)
		if err != nil {
			clog.Use(r.Context()).Warn("failed to respond to request", "err", err)
		}
	})
}

func parsInt(s string) (int, error) {
	sign := ""
	base := 10

	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		s = s[1:]
		sign = "-"
	}

	s, ok := strings.CutPrefix(s, "0x")
	if ok {
		base = 16
	}

	i, err := strconv.ParseInt(sign+s, base, strconv.IntSize)
	return int(i), err
}
