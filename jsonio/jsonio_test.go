package jsonio_test

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/jsonio"
)

func closeWriter(t *testing.T, w *jsonio.JsonWriter) {
	t.Helper()
	assert.NoError(t, (*io.PipeWriter)(w).Close())
}

type testPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestNewReader(t *testing.T) {
	var _ io.Reader = (*jsonio.JsonReader)(nil)

	r := jsonio.NewReader(testPayload{Name: "bob", Age: 42})
	b, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"bob","age":42}`, string(b))
}

func TestJsonReader_Read(t *testing.T) {
	var _ io.Reader = (*jsonio.JsonReader)(nil)

	r := jsonio.NewReader([]string{"a", "b"})
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Greater(t, n, 0)

	rest, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Contains(t, string(buf[:n])+string(rest), `"a"`)
}

func TestNewWriter(t *testing.T) {
	var _ io.Writer = (*jsonio.JsonWriter)(nil)

	var got testPayload
	w := jsonio.NewWriter(&got)
	_, err := w.Write([]byte(`{"name":"alice","age":30}`))
	assert.NoError(t, err)
	closeWriter(t, w)

	assert.Eventually(t, func() bool {
		return got.Name == "alice" && got.Age == 30
	}, time.Second, 10*time.Millisecond)
}

func TestJsonWriter_Write(t *testing.T) {
	var _ io.Writer = (*jsonio.JsonWriter)(nil)

	var got testPayload
	w := jsonio.NewWriter(&got)
	n, err := w.Write([]byte(`{"name":"carol","age":20}`))
	assert.NoError(t, err)
	assert.Equal(t, 25, n)
	closeWriter(t, w)

	assert.Eventually(t, func() bool {
		return got.Name == "carol" && got.Age == 20
	}, time.Second, 10*time.Millisecond)
}
