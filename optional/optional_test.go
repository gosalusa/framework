package optional_test

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

	"gosalusa.com/optional"
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
		o := optional.Some(42)
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("string", func(t *testing.T) {
		o := optional.Some("hello")
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, "hello", val)
	})

	t.Run("struct", func(t *testing.T) {
		want := person{Name: "Salusa", Age: 3}
		o := optional.Some(want)
		assert.True(t, o.Valid())
		val, ok := o.Ok()
		assert.True(t, ok)
		assert.Equal(t, want, val)
	})

	t.Run("zero value is still valid", func(t *testing.T) {
		o := optional.Some(0)
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
		{"valid int", optional.Some(123), `123`},
		{"invalid int", optional.Optional[int]{}, `null`},
		{"valid string", optional.Some("hello world"), `"hello world"`},
		{"invalid string", optional.Optional[string]{}, `null`},
		{"valid true", optional.Some(true), `true`},
		{"valid false", optional.Some(false), `false`},
		{"invalid bool", optional.Optional[bool]{}, `null`},
		{"valid float", optional.Some(3.14), `3.14`},
		{"invalid float", optional.Optional[float64]{}, `null`},
		{"valid struct", optional.Some(person{Name: "Bob", Age: 30}), `{"name":"Bob","age":30}`},
		{"invalid struct", optional.Optional[person]{}, `null`},
		{"valid time", optional.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), `"2024-01-15T10:30:00Z"`},
		{"invalid time", optional.Optional[time.Time]{}, `null`},
		{"valid slice", optional.Some([]int{1, 2, 3}), `[1,2,3]`},
		{"invalid slice", optional.Optional[[]int]{}, `null`},
		{"valid pointer", optional.Some(&i), `5`},
		{"invalid pointer", optional.Optional[*int]{}, `null`},
		{"valid nil pointer", optional.Some[*int](nil), `null`},
		{"valid struct field", struct {
			Name optional.Optional[string] `json:"name"`
		}{optional.Some("Salusa")}, `{"name":"Salusa"}`},
		{"invalid struct field", struct {
			Name optional.Optional[string] `json:"name"`
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
		_, err := json.Marshal(optional.Some(make(chan int)))
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
		{"valid int", `456`, optional.Some(456)},
		{"null int", `null`, optional.Optional[int]{}},
		{"valid string", `"world"`, optional.Some("world")},
		{"null string", `null`, optional.Optional[string]{}},
		{"empty string", `""`, optional.Some("")},
		{"valid true", `true`, optional.Some(true)},
		{"valid false", `false`, optional.Some(false)},
		{"null bool", `null`, optional.Optional[bool]{}},
		{"valid float", `9.81`, optional.Some(9.81)},
		{"null float", `null`, optional.Optional[float64]{}},
		{"valid struct", `{"name":"Alice","age":25}`, optional.Some(person{Name: "Alice", Age: 25})},
		{"null struct", `null`, optional.Optional[person]{}},
		{"valid time", `"2024-01-15T10:30:00Z"`, optional.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))},
		{"null time", `null`, optional.Optional[time.Time]{}},
		{"valid slice", `[4,5,6]`, optional.Some([]int{4, 5, 6})},
		{"null slice", `null`, optional.Optional[[]int]{}},
		{"valid pointer", `10`, optional.Some(&i)},
		{"null pointer", `null`, optional.Optional[*int]{}},
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
		{"string into int", `"abc"`, new(optional.Optional[int])},
		{"int into string", `123`, new(optional.Optional[string])},
		{"string into bool", `"true"`, new(optional.Optional[bool])},
		{"invalid struct field", `{"name":"Alice","age":"twenty-five"}`, new(optional.Optional[person])},
		{"invalid time", `"not-a-time"`, new(optional.Optional[time.Time])},
		{"truncated json", `{`, new(optional.Optional[person])},
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
		{"invalid int", optional.Optional[int]{}, ``},
		{"invalid string", optional.Optional[string]{}, ``},
		{"invalid time", optional.Optional[time.Time]{}, ``},
		{"string", optional.Some("salusa"), `salusa`},
		{"empty string", optional.Some(""), ``},
		{"time", optional.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), `2024-01-15T10:30:00Z`},
		{"text marshaler", optional.Some(upper("salusa")), `SALUSA`},
		{"int", optional.Some(123), `123`},
		{"false", optional.Some(false), `false`},
		{"struct", optional.Some(person{Name: "Bob", Age: 30}), `{"name":"Bob","age":30}`},
		{"slice", optional.Some([]int{1, 2}), `[1,2]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}

	t.Run("unsupported value", func(t *testing.T) {
		_, err := optional.Some(make(chan int)).MarshalText()
		assert.Error(t, err)
	})
}

func TestNull_UnmarshalText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want any
	}{
		{"empty int", ``, optional.Optional[int]{}},
		{"empty string", ``, optional.Optional[string]{}},
		{"empty time", ``, optional.Optional[time.Time]{}},
		{"int", `456`, optional.Some(456)},
		{"string", `world`, optional.Some("world")},
		{"true", `true`, optional.Some(true)},
		{"time", `2024-01-15T10:30:00Z`, optional.Some(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))},
		{"text unmarshaler", `SaLuSa`, optional.Some(upper("salusa"))},
		{"struct", `{"name":"Bob","age":30}`, optional.Some(person{Name: "Bob", Age: 30})},
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
		var n optional.Optional[undecodable]
		require.ErrorIs(t, n.UnmarshalText([]byte("nope")), errBadText)
		assert.Equal(t, optional.Optional[undecodable]{}, n)
	})

	t.Run("json fallback error", func(t *testing.T) {
		var n optional.Optional[int]
		assert.Error(t, n.UnmarshalText([]byte(`"abc"`)))
	})
}

func TestNull_Scan(t *testing.T) {
	ts := time.Date(2025, 4, 22, 10, 0, 0, 0, time.UTC)

	t.Run("int from int64", func(t *testing.T) {
		var n optional.Optional[int]
		require.NoError(t, n.Scan(int64(987)))
		assert.Equal(t, optional.Some(987), n)
	})

	t.Run("string from string", func(t *testing.T) {
		var n optional.Optional[string]
		require.NoError(t, n.Scan("db string"))
		assert.Equal(t, optional.Some("db string"), n)
	})

	t.Run("string from bytes", func(t *testing.T) {
		var n optional.Optional[string]
		require.NoError(t, n.Scan([]byte("db bytes")))
		assert.Equal(t, optional.Some("db bytes"), n)
	})

	t.Run("bool from bool", func(t *testing.T) {
		var n optional.Optional[bool]
		require.NoError(t, n.Scan(true))
		assert.Equal(t, optional.Some(true), n)
	})

	t.Run("time from time", func(t *testing.T) {
		var n optional.Optional[time.Time]
		require.NoError(t, n.Scan(ts))
		assert.Equal(t, optional.Some(ts), n)
	})

	t.Run("nil invalidates", func(t *testing.T) {
		n := optional.Some(100)
		require.NoError(t, n.Scan(nil))
		assert.Equal(t, optional.Optional[int]{}, n)

		s := optional.Some("initial")
		require.NoError(t, s.Scan(nil))
		assert.Equal(t, optional.Optional[string]{}, s)
	})

	t.Run("unsupported value", func(t *testing.T) {
		var n optional.Optional[int]
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
		{"valid int", optional.Some(55), int64(55)},
		{"invalid int", optional.Optional[int]{}, nil},
		{"valid string", optional.Some("sql value"), "sql value"},
		{"invalid string", optional.Optional[string]{}, nil},
		{"valid bool", optional.Some(true), true},
		{"valid time", optional.Some(ts), ts},
		{"invalid time", optional.Optional[time.Time]{}, nil},
		{"invalid struct", optional.Optional[person]{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("valid struct has no database representation", func(t *testing.T) {
		_, err := optional.Some(person{Name: "Bob"}).Value()
		assert.Error(t, err)
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("Null int", func(t *testing.T) {
		got, ok := optional.Unwrap(reflect.TypeFor[optional.Optional[int]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[int](), got)
	})

	t.Run("Null struct", func(t *testing.T) {
		got, ok := optional.Unwrap(reflect.TypeFor[optional.Optional[person]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[person](), got)
	})

	t.Run("nested Null", func(t *testing.T) {
		got, ok := optional.Unwrap(reflect.TypeFor[optional.Optional[optional.Optional[string]]]())
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeFor[optional.Optional[string]](), got)
	})

	t.Run("non Null struct", func(t *testing.T) {
		got, ok := optional.Unwrap(reflect.TypeFor[person]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("non struct", func(t *testing.T) {
		got, ok := optional.Unwrap(reflect.TypeFor[string]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("pointer to Null", func(t *testing.T) {
		got, ok := optional.Unwrap(reflect.TypeFor[*optional.Optional[int]]())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("nil type", func(t *testing.T) {
		got, ok := optional.Unwrap(nil)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestNull_OrElse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert.Equal(t, 42, optional.Some(42).OrElse(7))
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, 7, optional.Optional[int]{}.OrElse(7))
	})

	t.Run("valid zero value is not the fallback", func(t *testing.T) {
		assert.Equal(t, "", optional.Some("").OrElse("fallback"))
		assert.Equal(t, 0, optional.Some(0).OrElse(7))
	})
}

func TestNull_Map(t *testing.T) {
	double := func(i int) int { return i * 2 }
	describe := func(i int) string { return fmt.Sprintf("%d", i) }

	t.Run("valid", func(t *testing.T) {
		assert.Equal(t, optional.Some(42), optional.Some(21).Map(double))
	})

	t.Run("changes type", func(t *testing.T) {
		assert.Equal(t, optional.Some("21"), optional.Some(21).Map(describe))
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, optional.Optional[string]{}, optional.Optional[int]{}.Map(describe))
	})

	t.Run("fn is not called when invalid", func(t *testing.T) {
		var called bool
		got := optional.Optional[int]{}.Map(func(i int) string {
			called = true
			return "mapped"
		})
		assert.False(t, called)
		assert.Equal(t, optional.Optional[string]{}, got)
	})
}

func TestNone(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, optional.Optional[int]{}, optional.None[int]())
		assert.False(t, optional.None[int]().Valid())
	})

	t.Run("string", func(t *testing.T) {
		assert.Equal(t, optional.Optional[string]{}, optional.None[string]())
		assert.False(t, optional.None[string]().Valid())
	})

	t.Run("struct", func(t *testing.T) {
		assert.Equal(t, optional.Optional[person]{}, optional.None[person]())
		assert.False(t, optional.None[person]().Valid())
	})
}

func TestOfNull(t *testing.T) {
	t.Run("non-nil pointer", func(t *testing.T) {
		val := 42
		assert.Equal(t, optional.Some(42), optional.OfNull(&val))
	})

	t.Run("nil pointer", func(t *testing.T) {
		var p *int
		assert.Equal(t, optional.None[int](), optional.OfNull(p))
	})

	t.Run("non-nil string pointer", func(t *testing.T) {
		s := "hello"
		assert.Equal(t, optional.Some("hello"), optional.OfNull(&s))
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var s *string
		assert.Equal(t, optional.None[string](), optional.OfNull(s))
	})
}

func TestOptional_Ok(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		val, ok := optional.Some(42).Ok()
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("valid zero value", func(t *testing.T) {
		val, ok := optional.Some(0).Ok()
		assert.True(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("invalid int", func(t *testing.T) {
		val, ok := optional.None[int]().Ok()
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("invalid string", func(t *testing.T) {
		val, ok := optional.None[string]().Ok()
		assert.False(t, ok)
		assert.Equal(t, "", val)
	})
}

func TestOptional_Or(t *testing.T) {
	t.Run("valid does not invoke fn", func(t *testing.T) {
		var called bool
		got := optional.Some(42).Or(func() optional.Optional[int] {
			called = true
			return optional.Some(99)
		})
		assert.False(t, called)
		assert.Equal(t, optional.Some(42), got)
	})

	t.Run("invalid invokes fn returning valid", func(t *testing.T) {
		var called bool
		got := optional.None[int]().Or(func() optional.Optional[int] {
			called = true
			return optional.Some(99)
		})
		assert.True(t, called)
		assert.Equal(t, optional.Some(99), got)
	})

	t.Run("invalid invokes fn returning invalid", func(t *testing.T) {
		var called bool
		got := optional.None[int]().Or(func() optional.Optional[int] {
			called = true
			return optional.None[int]()
		})
		assert.True(t, called)
		assert.Equal(t, optional.None[int](), got)
	})
}

func TestOptional_All(t *testing.T) {
	t.Run("valid iterates once", func(t *testing.T) {
		var values []int
		for v := range optional.Some(42).All() {
			values = append(values, v)
		}
		assert.Equal(t, []int{42}, values)
	})

	t.Run("invalid does not iterate", func(t *testing.T) {
		var values []int
		for v := range optional.None[int]().All() {
			values = append(values, v)
		}
		assert.Empty(t, values)
	})

	t.Run("early break terminates iteration", func(t *testing.T) {
		var count int
		for range optional.Some("test").All() {
			count++
			break
		}
		assert.Equal(t, 1, count)
	})
}
