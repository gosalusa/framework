package handlertest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/testing/match"
)

type HttpResult struct {
	t                 *testing.T
	response          *http.Response
	body              []byte
	isBodyUnmarshaled bool
	unmarshaledBody   any
}

func (r *HttpResult) Body() []byte {
	if r.body != nil {
		return r.body
	}
	defer r.response.Body.Close()
	b, err := io.ReadAll(r.response.Body)
	if err != nil {
		panic(err)
	}
	r.body = b
	return r.body
}

func (r *HttpResult) getUnmarshaledBody() (any, bool) {
	r.t.Helper()
	if r.isBodyUnmarshaled {
		return r.unmarshaledBody, true
	}

	unmarshaledBody := [1]any{}

	body := []byte("[")
	body = append(body, r.Body()...)
	body = append(body, ']')
	err := json.Unmarshal(body, &unmarshaledBody)
	if err != nil {
		assert.Fail(r.t, "body is not json", "%s\n=====\n%s\n=====", err.Error(), r.Body())
		return nil, false
	}
	r.isBodyUnmarshaled = true
	r.unmarshaledBody = unmarshaledBody[0]
	return r.unmarshaledBody, true
}

func (r *HttpResult) AssertStatus(status int) *HttpResult {
	r.t.Helper()
	assert.Equal(r.t, status, r.response.StatusCode, "Statuses do not match")
	return r
}
func (r *HttpResult) AssertStatusRange(min, max int) *HttpResult {
	r.t.Helper()
	msg := fmt.Sprintf("Statuses must be between %d and %d", min, max)
	assert.GreaterOrEqual(r.t, r.response.StatusCode, min, msg)
	assert.LessOrEqual(r.t, r.response.StatusCode, max, msg)
	return r
}
func (r *HttpResult) AssertStatusOK() *HttpResult {
	r.t.Helper()
	return r.AssertStatusRange(200, 399)
}
func (r *HttpResult) AssertStatus2XX() *HttpResult {
	r.t.Helper()
	return r.AssertStatusRange(200, 299)
}
func (r *HttpResult) AssertStatus3XX() *HttpResult {
	r.t.Helper()
	return r.AssertStatusRange(300, 399)
}
func (r *HttpResult) AssertStatus4XX() *HttpResult {
	r.t.Helper()
	return r.AssertStatusRange(400, 499)
}
func (r *HttpResult) AssertStatus5XX() *HttpResult {
	r.t.Helper()
	return r.AssertStatusRange(500, 599)
}

func (r *HttpResult) AssertJSONString(jsonBody string) *HttpResult {
	r.t.Helper()
	var expected any
	err := json.Unmarshal([]byte(jsonBody), &expected)
	if err != nil {
		panic(err)
	}
	return r.AssertJSON(expected)
}
func (r *HttpResult) AssertJSON[T any](expected T) *HttpResult {
	r.t.Helper()
	acctual, ok := r.getUnmarshaledBody()
	if !ok {
		return r
	}
	assert.Equal(r.t, expected, acctual, "Response body does not match")
	return r
}

func (r *HttpResult) AssertJSONContains(path string, expected match.Matcher) *HttpResult {
	r.t.Helper()
	body, ok := r.getUnmarshaledBody()
	if !ok {
		return r
	}

	acctual, ok := lookup(body, path)
	if !ok {
		assert.Fail(r.t, "Response body is missing the expected property", "Property: %s", path)
		return r
	}

	res := expected.Matches(acctual)
	if !res.Matches {
		assert.Fail(r.t, "JSON value does not match", res.Description)
	}

	return r
}

func lookup(v any, path string) (any, bool) {
	if path == "" {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var result any
	var ok bool
	switch v := v.(type) {
	case map[string]any:
		result, ok = v[parts[0]]
	default:
		return nil, false
	}
	if !ok {
		return nil, false
	}
	if len(parts) > 1 {
		return lookup(result, strings.Join(parts[1:], "."))
	}
	return result, true
	// rv := rLookup(reflect.ValueOf(v), path)
	// if (rv == reflect.Value{}) {
	// 	return nil, false
	// }
	// return rv.Interface(), true
}

// func rLookup(v reflect.Value, path string) reflect.Value {
// 	if path == "" {
// 		return reflect.Value{}
// 	}

// 	switch v.Kind() {
// 	case reflect.Map:
// 		return mapLookup(v, path)
// 	case reflect.Pointer:
// 		return rLookup(v.Elem(), path)
// 	default:
// 		panic("should not get here")
// 	}
// }
// func mapLookup(v reflect.Value, path string) reflect.Value {
// 	keyType :=v.Type().Key()
// 	if keyType.ConvertibleTo(reflect.TypeFor[string]()){
// 		reflect.ValueOf(path)
// 		v.MapIndex()
// 		return
// 	}
// }
