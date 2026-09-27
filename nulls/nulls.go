package nulls

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"reflect"
)

var nullBytes = []byte("null")

var emptyText = []byte{}

type Null[T any] sql.Null[T]

var _ json.Marshaler = Null[int]{}
var _ json.Unmarshaler = (*Null[int])(nil)
var _ encoding.TextMarshaler = Null[int]{}
var _ encoding.TextUnmarshaler = (*Null[int])(nil)
var _ sql.Scanner = (*Null[int])(nil)
var _ driver.Valuer = Null[int]{}

func New[T any](n T) Null[T] {
	return Null[T]{
		V:     n,
		Valid: true,
	}
}

// MarshalJSON implements json.Marshaler.
func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.V)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Null[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, nullBytes) {
		var zero T
		n.V = zero
		n.Valid = false
		return nil
	}

	n.Valid = true
	return json.Unmarshal(b, &n.V)
}

// MarshalText implements encoding.TextMarshaler.
func (n Null[T]) MarshalText() ([]byte, error) {
	if !n.Valid {
		return emptyText, nil
	}

	if m, ok := any(n.V).(encoding.TextMarshaler); ok {
		return m.MarshalText()
	}

	v := reflect.ValueOf(n.V)
	if v.Kind() == reflect.String {
		return []byte(v.String()), nil
	}

	return json.Marshal(n.V)
}

// UnmarshalText implements encoding.TextUnmarshaler. An empty text decodes to
// an invalid null.
func (n *Null[T]) UnmarshalText(text []byte) error {
	var zero T
	if len(text) == 0 {
		n.V = zero
		n.Valid = false
		return nil
	}

	if u, ok := any(&n.V).(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText(text); err != nil {
			n.V = zero
			n.Valid = false
			return err
		}
		n.Valid = true
		return nil
	}

	v := reflect.ValueOf(&n.V).Elem()
	if v.Kind() == reflect.String {
		v.SetString(string(text))
		n.Valid = true
		return nil
	}

	n.Valid = true
	return json.Unmarshal(text, &n.V)
}

func (n *Null[T]) Scan(value any) error {
	return (*sql.Null[T])(n).Scan(value)
}

func (n Null[T]) Value() (driver.Value, error) {
	return sql.Null[T](n).Value()
}

// nullable is implemented by Null[T] so that the wrapped type can be recovered
// from a reflect.Type.
type nullable interface {
	wrappedType() reflect.Type
}

var _ nullable = Null[int]{}

func (n Null[T]) wrappedType() reflect.Type {
	return reflect.TypeFor[T]()
}

// Unwrap returns the type wrapped by the null of type t, and reports whether t
// is a Null. It allows reflection based code, such as migration generation, to
// treat a Null[T] as a nullable T.
func Unwrap(t reflect.Type) (reflect.Type, bool) {
	if t == nil || t.Kind() != reflect.Struct || !t.Implements(reflect.TypeFor[nullable]()) {
		return nil, false
	}

	return reflect.New(t).Elem().Interface().(nullable).wrappedType(), true
}
