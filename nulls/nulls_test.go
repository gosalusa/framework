package nulls_test

import (
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gosalusa.com/nulls"
)

var errBadText = errors.New("cannot decode from text")

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// upper is a named string type with its own text encoding, so it exercises the
// TextMarshaler and TextUnmarshaler branches of Null.
type upper string

func (u upper) MarshalText() ([]byte, error) {
	return []byte(strings.ToUpper(string(u))), nil
}

func (u *upper) UnmarshalText(b []byte) error {
	*u = upper(strings.ToLower(string(b)))
	return nil
}

// undecodable always fails to decode from text.
type undecodable string

func (u *undecodable) UnmarshalText([]byte) error {
	return errBadText
}

func TestNew(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, nulls.Null[int]{V: 42, Valid: true}, nulls.New(42))
	})

	t.Run("string", func(t *testing.T) {
		assert.Equal(t, nulls.Null[string]{V: "hello", Valid: true}, nulls.New("hello"))
	})

	t.Run("struct", func(t *testing.T) {
		want := person{Name: "Salusa", Age: 3}
		assert.Equal(t, nulls.Null[person]{V: want, Valid: true}, nulls.New(want))
	})

	t.Run("zero value is still valid", func(t *testing.T) {
		assert.True(t, nulls.New(0).Valid)
	})
}

func TestNull_MarshalJSON(t *testing.T) {
	i := 5

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"valid int", nulls.New(123), `123`},
		{"invalid int", nulls.Null[int]{}, `null`},
		{"valid string", nulls.New("hello world"), `"hello world"`},
		{"invalid string", nulls.Null[string]{}, `null`},
		{"valid true", nulls.New(true), `true`},
		{"valid false", nulls.New(false), `false`},
		{"invalid bool", nulls.Null[bool]{}, `null`},
		{"valid float", nulls.New(3.14), `3.14`},
		{"invalid float", nulls.Null[float64]{}, `null`},
		{"valid struct", nulls.New(person{Name: "Bob", Age: 30}), `{"name":"Bob","age":30}`},
		{"invalid struct", nulls.Null[person]{}, `null`},
		{"valid time", nulls.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), `"2024-01-15T10:30:00Z"`},
		{"invalid time", nulls.Null[time.Time]{}, `null`},
		{"valid slice", nulls.New([]int{1, 2, 3}), `[1,2,3]`},
		{"invalid slice", nulls.Null[[]int]{}, `null`},
		{"valid pointer", nulls.New(&i), `5`},
		{"invalid pointer", nulls.Null[*int]{}, `null`},
		{"valid nil pointer", nulls.New[*int](nil), `null`},
		{"valid struct field", struct {
			Name nulls.Null[string] `json:"name"`
		}{nulls.New("Salusa")}, `{"name":"Salusa"}`},
		{"invalid struct field", struct {
			Name nulls.Null[string] `json:"name"`
		}{}, `{"name":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}

	t.Run("unsupported value", func(t *testing.T) {
		_, err := json.Marshal(nulls.New(make(chan int)))
		assert.Error(t, err)
	})
}

func TestNull_UnmarshalJSON(t *testing.T) {
	i := 10

	tests := []struct {
		name string
		in   string
		want any
	}{
		{"valid int", `456`, nulls.New(456)},
		{"null int", `null`, nulls.Null[int]{}},
		{"valid string", `"world"`, nulls.New("world")},
		{"null string", `null`, nulls.Null[string]{}},
		{"empty string", `""`, nulls.New("")},
		{"valid true", `true`, nulls.New(true)},
		{"valid false", `false`, nulls.New(false)},
		{"null bool", `null`, nulls.Null[bool]{}},
		{"valid float", `9.81`, nulls.New(9.81)},
		{"null float", `null`, nulls.Null[float64]{}},
		{"valid struct", `{"name":"Alice","age":25}`, nulls.New(person{Name: "Alice", Age: 25})},
		{"null struct", `null`, nulls.Null[person]{}},
		{"valid time", `"2024-01-15T10:30:00Z"`, nulls.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))},
		{"null time", `null`, nulls.Null[time.Time]{}},
		{"valid slice", `[4,5,6]`, nulls.New([]int{4, 5, 6})},
		{"null slice", `null`, nulls.Null[[]int]{}},
		{"valid pointer", `10`, nulls.New(&i)},
		{"null pointer", `null`, nulls.Null[*int]{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := reflect.New(reflect.TypeOf(tt.want))
			require.NoError(t, json.Unmarshal([]byte(tt.in), target.Interface()))
			assert.Equal(t, tt.want, target.Elem().Interface())
		})
	}

	errs := []struct {
		name string
		in   string
		want any
	}{
		{"string into int", `"abc"`, new(nulls.Null[int])},
		{"int into string", `123`, new(nulls.Null[string])},
		{"string into bool", `"true"`, new(nulls.Null[bool])},
		{"invalid struct field", `{"name":"Alice","age":"twenty-five"}`, new(nulls.Null[person])},
		{"invalid time", `"not-a-time"`, new(nulls.Null[time.Time])},
		{"truncated json", `{`, new(nulls.Null[person])},
	}

	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tt.in), tt.want))
		})
	}
}

func TestNull_MarshalText(t *testing.T) {
	tests := []struct {
		name string
		in   encoding.TextMarshaler
		want string
	}{
		{"invalid int", nulls.Null[int]{}, ``},
		{"invalid string", nulls.Null[string]{}, ``},
		{"invalid time", nulls.Null[time.Time]{}, ``},
		{"string", nulls.New("salusa"), `salusa`},
		{"empty string", nulls.New(""), ``},
		{"time", nulls.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), `2024-01-15T10:30:00Z`},
		{"text marshaler", nulls.New(upper("salusa")), `SALUSA`},
		{"int", nulls.New(123), `123`},
		{"false", nulls.New(false), `false`},
		{"struct", nulls.New(person{Name: "Bob", Age: 30}), `{"name":"Bob","age":30}`},
		{"slice", nulls.New([]int{1, 2}), `[1,2]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}

	t.Run("unsupported value", func(t *testing.T) {
		_, err := nulls.New(make(chan int)).MarshalText()
		assert.Error(t, err)
	})
}

func TestNull_UnmarshalText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want any
	}{
		{"empty int", ``, nulls.Null[int]{}},
		{"empty string", ``, nulls.Null[string]{}},
		{"empty time", ``, nulls.Null[time.Time]{}},
		{"int", `456`, nulls.New(456)},
		{"string", `world`, nulls.New("world")},
		{"true", `true`, nulls.New(true)},
		{"time", `2024-01-15T10:30:00Z`, nulls.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))},
		{"text unmarshaler", `SaLuSa`, nulls.New(upper("salusa"))},
		{"struct", `{"name":"Bob","age":30}`, nulls.New(person{Name: "Bob", Age: 30})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := reflect.New(reflect.TypeOf(tt.want))
			targeter, ok := target.Interface().(encoding.TextUnmarshaler)
			require.True(t, ok)

			require.NoError(t, targeter.UnmarshalText([]byte(tt.in)))
			assert.Equal(t, tt.want, target.Elem().Interface())
		})
	}

	t.Run("text unmarshaler error invalidates", func(t *testing.T) {
		var n nulls.Null[undecodable]
		require.ErrorIs(t, n.UnmarshalText([]byte("nope")), errBadText)
		assert.Equal(t, nulls.Null[undecodable]{}, n)
	})

	t.Run("json fallback error", func(t *testing.T) {
		var n nulls.Null[int]
		assert.Error(t, n.UnmarshalText([]byte(`"abc"`)))
	})
}

func TestNull_Scan(t *testing.T) {
	ts := time.Date(2025, 4, 22, 10, 0, 0, 0, time.UTC)

	t.Run("int from int64", func(t *testing.T) {
		var n nulls.Null[int]
		require.NoError(t, n.Scan(int64(987)))
		assert.Equal(t, nulls.New(987), n)
	})

	t.Run("string from string", func(t *testing.T) {
		var n nulls.Null[string]
		require.NoError(t, n.Scan("db string"))
		assert.Equal(t, nulls.New("db string"), n)
	})

	t.Run("string from bytes", func(t *testing.T) {
		var n nulls.Null[string]
		require.NoError(t, n.Scan([]byte("db bytes")))
		assert.Equal(t, nulls.New("db bytes"), n)
	})

	t.Run("bool from bool", func(t *testing.T) {
		var n nulls.Null[bool]
		require.NoError(t, n.Scan(true))
		assert.Equal(t, nulls.New(true), n)
	})

	t.Run("time from time", func(t *testing.T) {
		var n nulls.Null[time.Time]
		require.NoError(t, n.Scan(ts))
		assert.Equal(t, nulls.New(ts), n)
	})

	t.Run("nil invalidates", func(t *testing.T) {
		n := nulls.New(100)
		require.NoError(t, n.Scan(nil))
		assert.Equal(t, nulls.Null[int]{}, n)

		s := nulls.New("initial")
		require.NoError(t, s.Scan(nil))
		assert.Equal(t, nulls.Null[string]{}, s)
	})

	t.Run("unsupported value", func(t *testing.T) {
		var n nulls.Null[int]
		assert.Error(t, n.Scan("not an int"))
	})
}

func TestNull_Value(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   driver.Valuer
		want driver.Value
	}{
		{"valid int", nulls.New(55), int64(55)},
		{"invalid int", nulls.Null[int]{}, nil},
		{"valid string", nulls.New("sql value"), "sql value"},
		{"invalid string", nulls.Null[string]{}, nil},
		{"valid bool", nulls.New(true), true},
		{"valid time", nulls.New(ts), ts},
		{"invalid time", nulls.Null[time.Time]{}, nil},
		{"invalid struct", nulls.Null[person]{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("valid struct has no database representation", func(t *testing.T) {
		_, err := nulls.New(person{Name: "Bob"}).Value()
		assert.Error(t, err)
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("Null int", func(t *testing.T) {
		got, ok := nulls.Unwrap(reflect.TypeFor[nulls.Null[int]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[int](), got)
	})

	t.Run("Null struct", func(t *testing.T) {
		got, ok := nulls.Unwrap(reflect.TypeFor[nulls.Null[person]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[person](), got)
	})

	t.Run("nested Null", func(t *testing.T) {
		got, ok := nulls.Unwrap(reflect.TypeFor[nulls.Null[nulls.Null[string]]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[nulls.Null[string]](), got)
	})

	t.Run("non Null struct", func(t *testing.T) {
		got, ok := nulls.Unwrap(reflect.TypeFor[person]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("non struct", func(t *testing.T) {
		got, ok := nulls.Unwrap(reflect.TypeFor[string]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("pointer to Null", func(t *testing.T) {
		got, ok := nulls.Unwrap(reflect.TypeFor[*nulls.Null[int]]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("nil type", func(t *testing.T) {
		got, ok := nulls.Unwrap(nil)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestNull_OrElse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert.Equal(t, 42, nulls.New(42).OrElse(7))
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, 7, nulls.Null[int]{}.OrElse(7))
	})

	t.Run("valid zero value is not the fallback", func(t *testing.T) {
		assert.Equal(t, "", nulls.New("").OrElse("fallback"))
		assert.Equal(t, 0, nulls.New(0).OrElse(7))
	})
}

func TestNull_Map(t *testing.T) {
	double := func(i int) int { return i * 2 }
	describe := func(i int) string { return fmt.Sprintf("%d", i) }

	t.Run("valid", func(t *testing.T) {
		assert.Equal(t, nulls.New(42), nulls.New(21).Map(double))
	})

	t.Run("changes type", func(t *testing.T) {
		assert.Equal(t, nulls.New("21"), nulls.New(21).Map(describe))
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, nulls.Null[string]{}, nulls.Null[int]{}.Map(describe))
	})

	t.Run("fn is not called when invalid", func(t *testing.T) {
		var called bool
		got := nulls.Null[int]{}.Map(func(i int) string {
			called = true
			return "mapped"
		})
		assert.False(t, called)
		assert.Equal(t, nulls.Null[string]{}, got)
	})
}
