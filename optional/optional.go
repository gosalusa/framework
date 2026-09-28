package optional

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

// Optional is a value of type T that may be null.
//
// The zero value is an invalid null holding the zero value of T. Use [Some] to
// build a valid null, and [Optional.OrElse] to fall back to a value when the null
// is invalid.
//
// Optional implements [json.Marshaler], [json.Unmarshaler],
// [encoding.TextMarshaler], [encoding.TextUnmarshaler], [sql.Scanner], and
// [driver.Valuer], so it can be used directly as a JSON field, a query
// parameter, or a database column. When Optional is used as a struct field with a
// database tag, migration generation treats the column as nullable, and the
// wrapped type is used to infer the column type.
type Optional[T any] sql.Null[T]

var _ json.Marshaler = Optional[int]{}
var _ json.Unmarshaler = (*Optional[int])(nil)
var _ encoding.TextMarshaler = Optional[int]{}
var _ encoding.TextUnmarshaler = (*Optional[int])(nil)
var _ sql.Scanner = (*Optional[int])(nil)
var _ driver.Valuer = Optional[int]{}

// Some returns a valid [Optional] holding n.
func Some[T any](n T) Optional[T] {
	return Optional[T]{
		V:     n,
		Valid: true,
	}
}

// Of returns a valid [Optional] holding n.
func OfNull[T any](n *T) Optional[T] {
	if n == nil {
		return None[T]()
	}
	return Some(*n)
}

// Some returns an empty [Optional].
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// MarshalJSON implements [json.Marshaler]. An invalid null marshals to the JSON
// null literal, and a valid null marshals to the JSON encoding of the wrapped
// value.
func (n Optional[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.V)
}

// UnmarshalJSON implements [json.Unmarshaler]. The JSON null literal sets the
// value to the zero value of T and marks the null invalid, and any other input
// is unmarshalled into the wrapped value and marks the null valid.
//
// A decoding error is returned with the null already marked valid, so a Null
// that is reused across decodes should be reset to its zero value first.
func (n *Optional[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, nullBytes) {
		var zero T
		n.V = zero
		n.Valid = false
		return nil
	}

	n.Valid = true
	return json.Unmarshal(b, &n.V)
}

// MarshalText implements [encoding.TextMarshaler]. An invalid null marshals to
// empty text, and a valid null marshals to the text encoding of the wrapped
// value.
//
// If the wrapped value implements [encoding.TextMarshaler] its text encoding is
// used, and a wrapped string is used verbatim. Every other wrapped value falls
// back to its JSON encoding, so the text may not be a faithful representation
// for types that do not have a text form.
func (n Optional[T]) MarshalText() ([]byte, error) {
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
func (n *Optional[T]) UnmarshalText(text []byte) error {
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

// Scan implements [sql.Scanner], reading a database value into the null. A nil
// value sets the value to the zero value of T and marks the null invalid, and
// any other value is converted to T and marks the null valid.
//
// An error is returned when the value cannot be converted to T.
func (n *Optional[T]) Scan(value any) error {
	return (*sql.Null[T])(n).Scan(value)
}

// Value implements [driver.Valuer], returning the value to store in a database
// column. An invalid null returns a nil value, and a valid null returns the
// wrapped value.
//
// An error is returned when the wrapped value has no database representation.
func (n Optional[T]) Value() (driver.Value, error) {
	return sql.Null[T](n).Value()
}

// nullable is implemented by Null[T] so that the wrapped type can be recovered
// from a reflect.Type.
type nullable interface {
	wrappedType() reflect.Type
}

var _ nullable = Optional[int]{}

func (n Optional[T]) wrappedType() reflect.Type {
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
func (n Optional[T]) OrElse(fallback T) T {
	if !n.Valid {
		return fallback
	}
	return n.V
}

// Map applies fn to the wrapped value and returns the result as a valid
// [Optional]. An invalid null is returned unchanged and fn is not called, so fn
// does not need to handle the zero value of T.
func (n Optional[T]) Map[U any](fn func(T) U) Optional[U] {
	if !n.Valid {
		return Optional[U]{}
	}
	return Some(fn(n.V))
}

func (n Optional[T]) IfPresent(fn func(T)) {
	if n.Valid {
		fn(n.V)
	}
}

func (n Optional[T]) Ok() (T, bool) {
	return n.V, n.Valid
}
