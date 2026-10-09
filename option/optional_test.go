package option_test

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

	"gosalusa.com/option"
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

func TestSome(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		o := option.Some(42)
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("string", func(t *testing.T) {
		o := option.Some("hello")
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, "hello", val)
	})

	t.Run("struct", func(t *testing.T) {
		want := person{Name: "Salusa", Age: 3}
		o := option.Some(want)
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, want, val)
	})

	t.Run("zero value is still valid", func(t *testing.T) {
		o := option.Some(0)
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestNull_MarshalJSON(t *testing.T) {
	i := 5

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"valid int", option.Some(123), `123`},
		{"invalid int", option.Option[int]{}, `null`},
		{"valid string", option.Some("hello world"), `"hello world"`},
		{"invalid string", option.Option[string]{}, `null`},
		{"valid true", option.Some(true), `true`},
		{"valid false", option.Some(false), `false`},
		{"invalid bool", option.Option[bool]{}, `null`},
		{"valid float", option.Some(3.14), `3.14`},
		{"invalid float", option.Option[float64]{}, `null`},
		{"valid struct", option.Some(person{Name: "Bob", Age: 30}), `{"name":"Bob","age":30}`},
		{"invalid struct", option.Option[person]{}, `null`},
		{"valid time", option.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), `"2024-01-15T10:30:00Z"`},
		{"invalid time", option.Option[time.Time]{}, `null`},
		{"valid slice", option.Some([]int{1, 2, 3}), `[1,2,3]`},
		{"invalid slice", option.Option[[]int]{}, `null`},
		{"valid pointer", option.Some(&i), `5`},
		{"invalid pointer", option.Option[*int]{}, `null`},
		{"valid nil pointer", option.Some[*int](nil), `null`},
		{"valid struct field", struct {
			Name option.Option[string] `json:"name"`
		}{option.Some("Salusa")}, `{"name":"Salusa"}`},
		{"invalid struct field", struct {
			Name option.Option[string] `json:"name"`
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
		_, err := json.Marshal(option.Some(make(chan int)))
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
		{"valid int", `456`, option.Some(456)},
		{"null int", `null`, option.Option[int]{}},
		{"valid string", `"world"`, option.Some("world")},
		{"null string", `null`, option.Option[string]{}},
		{"empty string", `""`, option.Some("")},
		{"valid true", `true`, option.Some(true)},
		{"valid false", `false`, option.Some(false)},
		{"null bool", `null`, option.Option[bool]{}},
		{"valid float", `9.81`, option.Some(9.81)},
		{"null float", `null`, option.Option[float64]{}},
		{"valid struct", `{"name":"Alice","age":25}`, option.Some(person{Name: "Alice", Age: 25})},
		{"null struct", `null`, option.Option[person]{}},
		{"valid time", `"2024-01-15T10:30:00Z"`, option.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))},
		{"null time", `null`, option.Option[time.Time]{}},
		{"valid slice", `[4,5,6]`, option.Some([]int{4, 5, 6})},
		{"null slice", `null`, option.Option[[]int]{}},
		{"valid pointer", `10`, option.Some(&i)},
		{"null pointer", `null`, option.Option[*int]{}},
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
		{"string into int", `"abc"`, new(option.Option[int])},
		{"int into string", `123`, new(option.Option[string])},
		{"string into bool", `"true"`, new(option.Option[bool])},
		{"invalid struct field", `{"name":"Alice","age":"twenty-five"}`, new(option.Option[person])},
		{"invalid time", `"not-a-time"`, new(option.Option[time.Time])},
		{"truncated json", `{`, new(option.Option[person])},
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
		{"invalid int", option.Option[int]{}, ``},
		{"invalid string", option.Option[string]{}, ``},
		{"invalid time", option.Option[time.Time]{}, ``},
		{"string", option.Some("salusa"), `salusa`},
		{"empty string", option.Some(""), ``},
		{"time", option.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), `2024-01-15T10:30:00Z`},
		{"text marshaler", option.Some(upper("salusa")), `SALUSA`},
		{"int", option.Some(123), `123`},
		{"false", option.Some(false), `false`},
		{"struct", option.Some(person{Name: "Bob", Age: 30}), `{"name":"Bob","age":30}`},
		{"slice", option.Some([]int{1, 2}), `[1,2]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}

	t.Run("unsupported value", func(t *testing.T) {
		_, err := option.Some(make(chan int)).MarshalText()
		assert.Error(t, err)
	})
}

func TestNull_UnmarshalText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want any
	}{
		{"empty int", ``, option.Option[int]{}},
		{"empty string", ``, option.Option[string]{}},
		{"empty time", ``, option.Option[time.Time]{}},
		{"int", `456`, option.Some(456)},
		{"string", `world`, option.Some("world")},
		{"true", `true`, option.Some(true)},
		{"time", `2024-01-15T10:30:00Z`, option.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))},
		{"text unmarshaler", `SaLuSa`, option.Some(upper("salusa"))},
		{"struct", `{"name":"Bob","age":30}`, option.Some(person{Name: "Bob", Age: 30})},
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
		var n option.Option[undecodable]
		require.ErrorIs(t, n.UnmarshalText([]byte("nope")), errBadText)
		assert.Equal(t, option.Option[undecodable]{}, n)
	})

	t.Run("json fallback error", func(t *testing.T) {
		var n option.Option[int]
		assert.Error(t, n.UnmarshalText([]byte(`"abc"`)))
	})
}

func TestNull_Scan(t *testing.T) {
	ts := time.Date(2025, 4, 22, 10, 0, 0, 0, time.UTC)

	t.Run("int from int64", func(t *testing.T) {
		var n option.Option[int]
		require.NoError(t, n.Scan(int64(987)))
		assert.Equal(t, option.Some(987), n)
	})

	t.Run("string from string", func(t *testing.T) {
		var n option.Option[string]
		require.NoError(t, n.Scan("db string"))
		assert.Equal(t, option.Some("db string"), n)
	})

	t.Run("string from bytes", func(t *testing.T) {
		var n option.Option[string]
		require.NoError(t, n.Scan([]byte("db bytes")))
		assert.Equal(t, option.Some("db bytes"), n)
	})

	t.Run("bool from bool", func(t *testing.T) {
		var n option.Option[bool]
		require.NoError(t, n.Scan(true))
		assert.Equal(t, option.Some(true), n)
	})

	t.Run("time from time", func(t *testing.T) {
		var n option.Option[time.Time]
		require.NoError(t, n.Scan(ts))
		assert.Equal(t, option.Some(ts), n)
	})

	t.Run("nil invalidates", func(t *testing.T) {
		n := option.Some(100)
		require.NoError(t, n.Scan(nil))
		assert.Equal(t, option.Option[int]{}, n)

		s := option.Some("initial")
		require.NoError(t, s.Scan(nil))
		assert.Equal(t, option.Option[string]{}, s)
	})

	t.Run("unsupported value", func(t *testing.T) {
		var n option.Option[int]
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
		{"valid int", option.Some(55), int64(55)},
		{"invalid int", option.Option[int]{}, nil},
		{"valid string", option.Some("sql value"), "sql value"},
		{"invalid string", option.Option[string]{}, nil},
		{"valid bool", option.Some(true), true},
		{"valid time", option.Some(ts), ts},
		{"invalid time", option.Option[time.Time]{}, nil},
		{"invalid struct", option.Option[person]{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("valid struct has no database representation", func(t *testing.T) {
		_, err := option.Some(person{Name: "Bob"}).Value()
		assert.Error(t, err)
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("Null int", func(t *testing.T) {
		got, ok := option.Unwrap(reflect.TypeFor[option.Option[int]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[int](), got)
	})

	t.Run("Null struct", func(t *testing.T) {
		got, ok := option.Unwrap(reflect.TypeFor[option.Option[person]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[person](), got)
	})

	t.Run("nested Null", func(t *testing.T) {
		got, ok := option.Unwrap(reflect.TypeFor[option.Option[option.Option[string]]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[option.Option[string]](), got)
	})

	t.Run("non Null struct", func(t *testing.T) {
		got, ok := option.Unwrap(reflect.TypeFor[person]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("non struct", func(t *testing.T) {
		got, ok := option.Unwrap(reflect.TypeFor[string]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("pointer to Null", func(t *testing.T) {
		got, ok := option.Unwrap(reflect.TypeFor[*option.Option[int]]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("nil type", func(t *testing.T) {
		got, ok := option.Unwrap(nil)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestNull_OrElse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert.Equal(t, 42, option.Some(42).OrElse(7))
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, 7, option.Option[int]{}.OrElse(7))
	})

	t.Run("valid zero value is not the fallback", func(t *testing.T) {
		assert.Equal(t, "", option.Some("").OrElse("fallback"))
		assert.Equal(t, 0, option.Some(0).OrElse(7))
	})
}

func TestNull_Map(t *testing.T) {
	double := func(i int) int { return i * 2 }
	describe := func(i int) string { return fmt.Sprintf("%d", i) }

	t.Run("valid", func(t *testing.T) {
		assert.Equal(t, option.Some(42), option.Some(21).Map(double))
	})

	t.Run("changes type", func(t *testing.T) {
		assert.Equal(t, option.Some("21"), option.Some(21).Map(describe))
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, option.Option[string]{}, option.Option[int]{}.Map(describe))
	})

	t.Run("fn is not called when invalid", func(t *testing.T) {
		var called bool
		got := option.Option[int]{}.Map(func(i int) string {
			called = true
			return "mapped"
		})
		assert.False(t, called)
		assert.Equal(t, option.Option[string]{}, got)
	})
}

func TestNone(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, option.Option[int]{}, option.None[int]())
		assert.False(t, option.None[int]().Valid())
	})

	t.Run("string", func(t *testing.T) {
		assert.Equal(t, option.Option[string]{}, option.None[string]())
		assert.False(t, option.None[string]().Valid())
	})

	t.Run("struct", func(t *testing.T) {
		assert.Equal(t, option.Option[person]{}, option.None[person]())
		assert.False(t, option.None[person]().Valid())
	})
}

func TestOfNull(t *testing.T) {
	t.Run("non-nil pointer", func(t *testing.T) {
		val := 42
		assert.Equal(t, option.Some(42), option.OfNull(&val))
	})

	t.Run("nil pointer", func(t *testing.T) {
		var p *int
		assert.Equal(t, option.None[int](), option.OfNull(p))
	})

	t.Run("non-nil string pointer", func(t *testing.T) {
		s := "hello"
		assert.Equal(t, option.Some("hello"), option.OfNull(&s))
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var s *string
		assert.Equal(t, option.None[string](), option.OfNull(s))
	})
}

func TestOptional_Ok(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		val, ok := option.Some(42).Ok()
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("valid zero value", func(t *testing.T) {
		val, ok := option.Some(0).Ok()
		assert.True(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("invalid int", func(t *testing.T) {
		val, ok := option.None[int]().Ok()
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("invalid string", func(t *testing.T) {
		val, ok := option.None[string]().Ok()
		assert.False(t, ok)
		assert.Equal(t, "", val)
	})
}

func TestOptional_Or(t *testing.T) {
	t.Run("valid does not invoke fn", func(t *testing.T) {
		var called bool
		got := option.Some(42).Or(func() option.Option[int] {
			called = true
			return option.Some(99)
		})
		assert.False(t, called)
		assert.Equal(t, option.Some(42), got)
	})

	t.Run("invalid invokes fn returning valid", func(t *testing.T) {
		var called bool
		got := option.None[int]().Or(func() option.Option[int] {
			called = true
			return option.Some(99)
		})
		assert.True(t, called)
		assert.Equal(t, option.Some(99), got)
	})

	t.Run("invalid invokes fn returning invalid", func(t *testing.T) {
		var called bool
		got := option.None[int]().Or(func() option.Option[int] {
			called = true
			return option.None[int]()
		})
		assert.True(t, called)
		assert.Equal(t, option.None[int](), got)
	})
}

func TestOptional_All(t *testing.T) {
	t.Run("valid iterates once", func(t *testing.T) {
		var values []int
		for v := range option.Some(42).All() {
			values = append(values, v)
		}
		assert.Equal(t, []int{42}, values)
	})

	t.Run("invalid does not iterate", func(t *testing.T) {
		var values []int
		for v := range option.None[int]().All() {
			values = append(values, v)
		}
		assert.Empty(t, values)
	})

	t.Run("early break terminates iteration", func(t *testing.T) {
		var count int
		for range option.Some("test").All() {
			count++
			break
		}
		assert.Equal(t, 1, count)
	})
}
