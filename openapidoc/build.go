package openapidoc

import (
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-openapi/spec"
	"github.com/google/uuid"
	"gosalusa.com/internal/helpers"
)

var (
	// typeByteSlice             = reflect.TypeFor[[]byte]()
	typeTime                  = reflect.TypeFor[time.Time]()
	typeRune                  = reflect.TypeFor[rune]()
	typeEncodingTextMarshaler = reflect.TypeFor[encoding.TextMarshaler]()
	typeJSONMarshaler         = reflect.TypeFor[json.Marshaler]()
	typeHttpResponse          = reflect.TypeFor[*http.Response]()
)

var formatMap = map[reflect.Type]string{
	typeTime: "date-time",
}

func init() {
	RegisterFormat[uuid.UUID]("uuid")
}

func RegisterFormat[T any](format string) {
	formatMap[reflect.TypeFor[T]()] = format
}

func Param(t reflect.Type, name, in string) (*spec.Parameter, error) {
	var err error
	param := &spec.Parameter{}
	param.Schema, err = Schema(t, false)
	if err != nil {
		return nil, fmt.Errorf("param %s: %w", name, err)
	}
	param.Type = Type(t)
	param.Format = Format(t)
	param.Name = name
	param.In = in
	return param, nil
}
func Type(t reflect.Type) string {
	if t.Implements(typeEncodingTextMarshaler) {
		return "string"
	}
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	case reflect.String:
		return "string"
	case reflect.Pointer:
		return Type(t.Elem())
	default:
		return ""
	}
}

func Format(t reflect.Type) string {
	format, ok := formatMap[t]
	if ok {
		return format
	}
	if t.Implements(typeEncodingTextMarshaler) {
		return ""
	}

	switch t.Kind() {
	case reflect.Pointer:
		return Format(t.Elem())
	case reflect.Bool, reflect.Array, reflect.Slice, reflect.Map, reflect.Struct, reflect.String, reflect.Int:
		return ""
	default:
		return t.Kind().String()
	}
}

var contentTypeMap = map[reflect.Type]string{}

func RegisterContentType[T any](contentType string) {
	contentTypeMap[reflect.TypeFor[T]()] = contentType
}

func GetContentType[T any]() (string, bool) {
	ct, ok := contentTypeMap[reflect.TypeFor[T]()]
	return ct, ok
}

var responseMap = map[reflect.Type]*spec.Response{
	typeHttpResponse: spec.NewResponse(),
}

func RegisterResponse[T any](response *spec.Response) {
	responseMap[reflect.TypeFor[T]()] = response
}
func Response(t reflect.Type) (*spec.Response, error) {
	r, ok := responseMap[t]
	if ok {
		return r, nil
	}

	resp := spec.NewResponse()
	s, err := Schema(t, false)
	if err != nil {
		slog.Warn("Could not build schema for api response", "type", t, "err", err)
	} else {
		resp.Schema = s
	}

	return resp, nil
}

var schemaMap = map[reflect.Type]*spec.Schema{
	typeTime: spec.DateTimeProperty(),
	typeRune: spec.CharProperty(),
}

func RegisterSchema[T any](schema *spec.Schema) {
	schemaMap[reflect.TypeFor[T]()] = schema
}

func Schema(t reflect.Type, requireTag bool) (*spec.Schema, error) {
	s, ok := schemaMap[t]
	if ok {
		return s, nil
	}

	if t.Implements(typeJSONMarshaler) {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"unknown"}}}, nil
	}
	if t.Implements(typeEncodingTextMarshaler) {
		return spec.StringProperty(), nil
	}
	// fmt.Printf("%v %v\n", t, t.Kind())
	switch t.Kind() {
	case reflect.Bool:
		return spec.BoolProperty(), nil
	case reflect.Int8:
		return spec.Int8Property(), nil
	case reflect.Int16, reflect.Uint8:
		return spec.Int16Property(), nil
	case reflect.Int32, reflect.Uint16:
		return spec.Int32Property(), nil
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return spec.Int64Property(), nil
	case reflect.Float32:
		return spec.Float32Property(), nil
	case reflect.Float64:
		return spec.Float64Property(), nil
	case reflect.Array, reflect.Slice:
		item, err := Schema(t.Elem(), requireTag)
		if err != nil {
			return nil, err
		}
		return spec.ArrayProperty(item), nil
	case reflect.Map:
		item, err := Schema(t.Elem(), requireTag)
		if err != nil {
			return nil, err
		}
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map keys expected strings found %s", t.Key().Kind())
		}
		return spec.MapProperty(item), nil
	case reflect.Pointer:
		item, err := Schema(t.Elem(), requireTag)
		if err != nil {
			return nil, err
		}
		return item.AsNullable(), nil
	case reflect.String:
		return spec.StringProperty(), nil
	case reflect.Struct:

		schema := &spec.Schema{}
		schema.AddType("object", "")

		for _, field := range helpers.GetFields(t) {
			name, ok := field.Tag.Lookup("json")
			if ok {
				name = strings.Split(name, ",")[0]
				if name == "-" {
					continue
				}
			} else {
				if requireTag {
					continue
				}
				name = field.Name
			}
			item, err := Schema(field.Type, requireTag)
			if err != nil {
				return nil, err
			}
			schema.SetProperty(name, *item)
		}

		return schema, nil
	default:
		return nil, fmt.Errorf("unsupported type %v %v", t, t.Kind())
	}
}
