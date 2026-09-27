// Package nulls provides a generic Null type for values that may be null,
// suitable for use as JSON fields, query parameters, and database columns.
//
// Unlike the single field types in database/sql, Null is generic over the type
// it wraps, so a nullable string, time, or struct needs no bespoke type. The
// zero value is an invalid null, New builds a valid one, and OrElse falls back
// to a value when the null is invalid:
//
//	name := nulls.New("Salusa").OrElse("unknown")
//	age := nulls.Null[int]{}.OrElse(0)
//
// A Null marshals to the JSON null literal when invalid, and to the JSON
// encoding of the wrapped value when valid. The same type implements
// [encoding.TextMarshaler] and [encoding.TextUnmarshaler], where empty text
// decodes to an invalid null, which is what makes a Null bind cleanly from a
// missing or blank query parameter.
package optional
