package request

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gosalusa.com/internal/helpers"
)

const (
	// Decimal

	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	// Binary

	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

var (
	textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
	fsFileType          = reflect.TypeFor[fs.File]()
)

type File struct {
	file   multipart.File
	handle *FileInfo
}

var _ fs.FileInfo = (*FileInfo)(nil)

// Close implements fs.File.
func (f *File) Close() error {
	return f.file.Close()
}

// Read implements fs.File.
func (f *File) Read(b []byte) (int, error) {
	return f.file.Read(b)
}

// Stat implements fs.File.
func (f *File) Stat() (fs.FileInfo, error) {
	return f.handle, nil
}

type FileInfo struct{ header *multipart.FileHeader }

// IsDir implements fs.FileInfo.
func (f *FileInfo) IsDir() bool {
	return false
}

// ModTime implements fs.FileInfo.
func (f *FileInfo) ModTime() time.Time {
	return time.Now()
}

// Mode implements fs.FileInfo.
func (f *FileInfo) Mode() fs.FileMode {
	return 0644
}

// Name implements fs.FileInfo.
func (f *FileInfo) Name() string {
	return f.header.Filename
}

// Size implements fs.FileInfo.
func (f *FileInfo) Size() int64 {
	return f.header.Size
}

// Sys implements fs.FileInfo.
func (f *FileInfo) Sys() any {
	return nil
}

func Run(requestHttp *http.Request, requestStruct any) error {
	urlArgs := map[string]map[string][]string{
		"query": requestHttp.URL.Query(),
		"path":  pathArgs(requestHttp),
	}

	var jsonBody map[string]json.RawMessage

	if requestHttp.Body != http.NoBody {
		defer requestHttp.Body.Close()
		body, err := io.ReadAll(requestHttp.Body)
		if err != nil {
			return err
		}

		requestHttp.Body = io.NopCloser(bytes.NewBuffer(body))

		contentType := strings.Split(requestHttp.Header.Get("Content-Type"), ";")[0]
		switch contentType {
		case "application/x-www-form-urlencoded":
			bodyQuery, err := url.ParseQuery(string(body))
			if err != nil {
				return err
			}
			urlArgs["json"] = bodyQuery
		case "multipart/form-data":
			err := requestHttp.ParseMultipartForm(100 * MB)
			if err != nil {
				return err
			}
			urlArgs["json"] = requestHttp.MultipartForm.Value
		default:
			jsonBody = map[string]json.RawMessage{}
			err := json.Unmarshal(body, &jsonBody)
			if err != nil {
				return fmt.Errorf("could not decode body: %w", err)
			}
		}
	}

	verr := ValidationError{}
	err := helpers.EachField(reflect.ValueOf(requestStruct), func(sf reflect.StructField, fv reflect.Value) error {
		for tag, args := range urlArgs {
			err := setQuery(sf, fv, tag, args)
			if err != nil {
				verr.AddError(sf.Tag.Get(tag), err.Error())
			}
		}

		err := setFile(sf, fv, requestHttp)
		if err != nil {
			verr.AddError(sf.Tag.Get("json"), err.Error())
		}

		if jsonBody == nil {
			return nil
		}

		err = setJSON(sf, fv, jsonBody)
		if err != nil {
			verr.AddError(sf.Tag.Get("json"), err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}

	if verr.HasErrors() {
		return verr
	}

	err = Validate(requestHttp, requestStruct)
	if err != nil {
		return err
	}

	return nil
}

func pathArgs(r *http.Request) map[string][]string {
	args := map[string][]string{}
	for k, v := range mux.Vars(r) {
		args[k] = []string{v}
	}
	return args
}

func setQuery(sf reflect.StructField, fv reflect.Value, tag string, args map[string][]string) error {
	tagValue, ok := sf.Tag.Lookup(tag)
	if !ok {
		return nil
	}
	arg, ok := args[tagValue]
	if !ok {
		return nil
	}

	rv, err := decode(sf.Type, arg)
	if err != nil {
		return err
	}
	fv.Set(rv)

	return nil
}

func setFile(sf reflect.StructField, fv reflect.Value, requestHttp *http.Request) error {
	tagValue, ok := sf.Tag.Lookup("json")
	if !ok {
		return nil
	}

	if sf.Type != fsFileType {
		return nil
	}

	file, handle, err := requestHttp.FormFile(tagValue)
	if err != nil {
		return err
	}

	fv.Set(reflect.ValueOf(&File{
		file: file,
		handle: &FileInfo{
			header: handle,
		},
	}))
	return nil
}

func setJSON(sf reflect.StructField, fv reflect.Value, jsonBody map[string]json.RawMessage) error {
	tagValue, ok := sf.Tag.Lookup("json")
	if !ok {
		return nil
	}
	b, ok := jsonBody[tagValue]
	if !ok {
		return nil
	}
	var v any
	isPtr := sf.Type.Kind() == reflect.Pointer
	if isPtr {
		if bytes.Equal(b, []byte("null")) {
			fv.Set(reflect.Zero(sf.Type))
			return nil
		}
		v = reflect.New(sf.Type.Elem()).Interface()
	} else {
		v = reflect.New(sf.Type).Interface()
	}
	err := json.Unmarshal(b, v)
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return fmt.Errorf("should be of type %s", sf.Type.Kind())
	} else if err != nil {
		return err
	}
	if !isPtr {
		fv.Set(reflect.ValueOf(v).Elem())
	} else {
		fv.Set(reflect.ValueOf(v))
	}
	return nil
}

func decode(t reflect.Type, values []string) (reflect.Value, error) {
	if t.Implements(textUnmarshalerType) {
		v := helpers.Create(t).Interface()
		err := v.(encoding.TextUnmarshaler).UnmarshalText([]byte(values[0]))
		return reflect.ValueOf(v), err
	}

	if reflect.PointerTo(t).Implements(textUnmarshalerType) {
		v := helpers.Create(reflect.PointerTo(t)).Interface()
		err := v.(encoding.TextUnmarshaler).UnmarshalText([]byte(values[0]))
		return reflect.ValueOf(v).Elem(), err
	}

	switch t.Kind() {
	case reflect.Slice:
		return decodeSlice(t, values)

	case reflect.Pointer:
		return decodePointer(t, values)

	default:
		return decodeBuiltin(t, values)
	}
}

func decodeSlice(t reflect.Type, values []string) (reflect.Value, error) {
	sliceT := t.Elem()
	slice := reflect.MakeSlice(t, 0, len(values))
	for _, part := range values {
		result, err := decode(sliceT, []string{part})
		if err != nil {
			return invalidValue, err
		}
		slice = reflect.Append(slice, result)
	}
	return slice, nil
}

func decodePointer(t reflect.Type, values []string) (reflect.Value, error) {
	if values[0] == "" {
		return reflect.New(t).Elem(), nil
	}
	v, err := decode(t.Elem(), values)
	if err != nil {
		return reflect.Value{}, err
	}
	// result := v.Interface()
	vPtr := reflect.New(t.Elem())
	vPtr.Elem().Set(v)
	return vPtr, nil
}

func decodeBuiltin(t reflect.Type, values []string) (reflect.Value, error) {
	conv, ok := builtinConverters[t.Kind()]
	if !ok {
		return invalidValue, fmt.Errorf("no converter for %s", t.Kind())
	}
	result := conv(values[0])
	if result == invalidValue {
		return invalidValue, fmt.Errorf(`should be of type %s: value "%s"`, t.Kind(), values[0])
	}

	if t != result.Type() {
		return result.Convert(t), nil
	}
	return result, nil
}
