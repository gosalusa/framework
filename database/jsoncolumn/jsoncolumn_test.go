package jsoncolumn_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/jsoncolumn"
)

func TestScan(t *testing.T) {
	t.Run("from []byte", func(t *testing.T) {
		var v map[string]int
		err := jsoncolumn.Scan(&v, []byte(`{"a":1}`))
		assert.NoError(t, err)
		assert.Equal(t, map[string]int{"a": 1}, v)
	})

	t.Run("from string", func(t *testing.T) {
		var v []string
		err := jsoncolumn.Scan(&v, `["x","y"]`)
		assert.NoError(t, err)
		assert.Equal(t, []string{"x", "y"}, v)
	})

	t.Run("from nil", func(t *testing.T) {
		var v = map[string]int{"a": 1}
		err := jsoncolumn.Scan(&v, nil)
		assert.NoError(t, err)
		assert.Empty(t, v)
	})

	t.Run("invalid type", func(t *testing.T) {
		var v map[string]int
		err := jsoncolumn.Scan(&v, 42)
		assert.EqualError(t, err, "invalid type int")
	})

	t.Run("invalid json", func(t *testing.T) {
		var v map[string]int
		err := jsoncolumn.Scan(&v, []byte(`not-json`))
		assert.Error(t, err)
	})

	t.Run("non-pointer destination", func(t *testing.T) {
		var v map[string]int
		err := jsoncolumn.Scan(v, []byte(`{}`))
		assert.Error(t, err)
		var unmarshalErr *json.InvalidUnmarshalError
		assert.ErrorAs(t, err, &unmarshalErr)
	})
}

func TestValue(t *testing.T) {
	v, err := jsoncolumn.Value(map[string]int{"b": 2})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"b":2}`, string(v.([]byte)))
}

func TestMap(t *testing.T) {
	var m jsoncolumn.Map[string, int]
	err := m.Scan([]byte(`{"a":1}`))
	assert.NoError(t, err)
	assert.Equal(t, 1, m["a"])

	v, err := m.Value()
	assert.NoError(t, err)
	assert.JSONEq(t, `{"a":1}`, string(v.([]byte)))

	var _ driver.Valuer = m
	var _ sql.Scanner = &m
}

func TestSlice(t *testing.T) {
	var s jsoncolumn.Slice[int]
	err := s.Scan([]byte(`[1,2,3]`))
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, []int(s))

	v, err := s.Value()
	assert.NoError(t, err)
	assert.JSONEq(t, `[1,2,3]`, string(v.([]byte)))

	var _ driver.Valuer = s
	var _ sql.Scanner = &s
}
