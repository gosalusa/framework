package option

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"iter"
	"reflect"
)

var nullBytes = []byte("null")

var emptyText = []byte{}

// Option is a value of type T that may be null.
//
// The zero value is an invalid null holding the zero value of T. Use [Some] to
// build a valid null, and [Option.OrElse] to fall back to a value when the null
// is invalid.
//
// Option implements [json.Marshaler], [json.Unmarshaler],
// [encoding.TextMarshaler], [encoding.TextUnmarshaler], [sql.Scanner], and
// [driver.Valuer], so it can be used directly as a JSON field, a query
// parameter, or a database column. When Option is used as a struct field with a
// database tag, migration generation treats the column as nullable, and the
// wrapped type is used to infer the column type.
type Option[T any] struct {
	null sql.Null[T]
}

var _ json.Marshaler = Option[int]{}
var _ json.Unmarshaler = (*Option[int])(nil)
var _ encoding.TextMarshaler = Option[int]{}
var _ encoding.TextUnmarshaler = (*Option[int])(nil)
var _ sql.Scanner = (*Option[int])(nil)
var _ driver.Valuer = Option[int]{}

// Some returns a valid [Option] holding n.
func Some[T any](n T) Option[T] {
	return Option[T]{
		null: sql.Null[T]{
			V:     n,
			Valid: true,
		},
	}
}

// OfNull returns a valid [Option] holding the value pointed to by n, or an
// empty [Option] if n is nil.
func OfNull[T any](n *T) Option[T] {
	if n == nil {
		return None[T]()
	}
	return Some(*n)
}

// None returns an empty [Option].
func None[T any]() Option[T] {
	return Option[T]{}
}

// MarshalJSON implements [json.Marshaler]. An invalid null marshals to the JSON
// null literal, and a valid null marshals to the JSON encoding of the wrapped
// value.
func (n Option[T]) MarshalJSON() ([]byte, error) {
	if !n.null.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.null.V)
}

// UnmarshalJSON implements [json.Unmarshaler]. The JSON null literal sets the
// value to the zero value of T and marks the null invalid, and any other input
// is unmarshalled into the wrapped value and marks the null valid.
//
// A decoding error is returned with the null already marked valid, so a Null
// that is reused across decodes should be reset to its zero value first.
func (n *Option[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, nullBytes) {
		var zero T
		n.null.V = zero
		n.null.Valid = false
		return nil
	}

	n.null.Valid = true
	return json.Unmarshal(b, &n.null.V)
}

// MarshalText implements [encoding.TextMarshaler]. An invalid null marshals to
// empty text, and a valid null marshals to the text encoding of the wrapped
// value.
//
// If the wrapped value implements [encoding.TextMarshaler] its text encoding is
// used, and a wrapped string is used verbatim. Every other wrapped value falls
// back to its JSON encoding, so the text may not be a faithful representation
// for types that do not have a text form.
func (n Option[T]) MarshalText() ([]byte, error) {
	if !n.null.Valid {
		return emptyText, nil
	}

	if m, ok := any(n.null.V).(encoding.TextMarshaler); ok {
		return m.MarshalText()
	}

	v := reflect.ValueOf(n.null.V)
	if v.Kind() == reflect.String {
		return []byte(v.String()), nil
	}

	return json.Marshal(n.null.V)
}

// UnmarshalText implements [encoding.TextUnmarshaler]. Empty text decodes to an
// invalid null holding the zero value of T, and any other text is unmarshalled
// into the wrapped value and marks the null valid.
//
// If the wrapped value implements [encoding.TextUnmarshaler] the text is
// decoded by it, and a wrapped string is set to the text directly. Every other
// wrapped value falls back to JSON, so the text must be a JSON encoding.
//
// A decoding error is returned with the null left invalid holding the zero
// value of T.
//
// Decoding an empty text to an invalid null is what makes a Null usable as a
// query parameter: a parameter that is absent or blank decodes to an invalid
// null rather than to the zero value of T.
func (n *Option[T]) UnmarshalText(text []byte) error {
	var zero T
	if len(text) == 0 {
		n.null.V = zero
		n.null.Valid = false
		return nil
	}

	if u, ok := any(&n.null.V).(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText(text); err != nil {
			n.null.V = zero
			n.null.Valid = false
			return err
		}
		n.null.Valid = true
		return nil
	}

	v := reflect.ValueOf(&n.null.V).Elem()
	if v.Kind() == reflect.String {
		v.SetString(string(text))
		n.null.Valid = true
		return nil
	}

	n.null.Valid = true
	return json.Unmarshal(text, &n.null.V)
}

// Scan implements [sql.Scanner], reading a database value into the null. A nil
// value sets the value to the zero value of T and marks the null invalid, and
// any other value is converted to T and marks the null valid.
//
// An error is returned when the value cannot be converted to T.
func (n *Option[T]) Scan(value any) error {
	return n.null.Scan(value)
}

// Value implements [driver.Valuer], returning the value to store in a database
// column. An invalid null returns a nil value, and a valid null returns the
// wrapped value.
//
// An error is returned when the wrapped value has no database representation.
func (n Option[T]) Value() (driver.Value, error) {
	return n.null.Value()
}

// nullable is implemented by Null[T] so that the wrapped type can be recovered
// from a reflect.Type.
type nullable interface {
	wrappedType() reflect.Type
}

var _ nullable = Option[int]{}

func (n Option[T]) wrappedType() reflect.Type {
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

// OrElse returns the wrapped value if the null is valid, and fallback
// otherwise.
func (n Option[T]) OrElse(fallback T) T {
	if !n.null.Valid {
		return fallback
	}
	return n.null.V
}

// Map applies fn to the wrapped value and returns the result as a valid
// [Option]. An invalid null is returned unchanged and fn is not called, so fn
// does not need to handle the zero value of T.
func (n Option[T]) Map[U any](fn func(T) U) Option[U] {
	if !n.null.Valid {
		return Option[U]{}
	}
	return Some(fn(n.null.V))
}

// Ok returns the wrapped value and true if the [Option] is valid, or the zero
// value of T and false otherwise.
func (n Option[T]) Ok() (T, bool) {
	return n.null.V, n.null.Valid
}

// Or returns n if it is valid, and otherwise calls fn and returns its result.
// fn is only evaluated if n is invalid, allowing lazy fallback resolution.
func (n Option[T]) Or(fn func() Option[T]) Option[T] {
	if n.null.Valid {
		return n
	}
	return fn()
}

// All returns an iterator over the wrapped value. If the [Option] is valid,
// it yields the value once; otherwise, it yields nothing.
func (n Option[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if !n.null.Valid {
			return
		}
		yield(n.null.V)
	}
}

// All returns an iterator over the wrapped value. If the [Option] is valid,
// it yields the value once; otherwise, it yields nothing.
func (n Option[T]) Valid() bool {
	return n.null.Valid
}
