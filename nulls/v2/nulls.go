package nulls

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type Null[T any] sql.Null[T]

var _ json.Marshaler = Null[int]{}
var _ json.Unmarshaler = (*Null[int])(nil)
var _ sql.Scanner = (*Null[int])(nil)

func (n *Null[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		var zero T
		n.V = zero
		n.Valid = false
		return nil
	}
	n.Valid = true
	return json.Unmarshal(b, &n.V)
}

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.V)
}

func (n *Null[T]) Scan(src any) error {
	return (*sql.Null[T])(n).Scan(src)
}

func (n *Null[T]) Value() (driver.Value, error) {
	return (*sql.Null[T])(n).Value()
}
