package sqlite

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
)

// SQLiteCore implements generic.Core for SQLite. It quotes identifiers with
// double quotes, uses ? bindings, and supports RETURNING.
type SQLiteCore struct{}

// New returns a dialect that renders SQL for SQLite.
func New() dialects.Dialect {
	return generic.New(&SQLiteCore{})
}

// Identifier returns s quoted as a SQLite identifier. A "*" segment is left
// unquoted.
func (*SQLiteCore) Identifier(s string) string {
	if s == "*" {
		return s
	}
	parts := strings.Split(s, ".")
	for i, p := range parts {
		if p == "*" {
			continue
		}
		parts[i] = `"` + p + `"`
	}
	return strings.Join(parts, ".")
}

// DataType maps a dialects.DataType to a SQLite column type.
func (*SQLiteCore) DataType(t dialects.DataType) string {
	switch t.Name {
	case dialects.DataTypeString.Name, dialects.DataTypeText.Name, dialects.DataTypeJSON.Name:
		return "TEXT"
	case dialects.DataTypeDate.Name, dialects.DataTypeDateTime.Name:
		return "TIMESTAMP"
	case dialects.DataTypeInt32.Name, dialects.DataTypeUInt32.Name, dialects.DataTypeBoolean.Name:
		return "INTEGER"
	case dialects.DataTypeFloat32.Name:
		return "FLOAT"
	}
	return t.Name
}

// CurrentTime returns the SQLite expression for the current timestamp.
func (*SQLiteCore) CurrentTime() string {
	return "CURRENT_TIMESTAMP"
}

// AutoIncrement returns the clause that makes a column an auto-incrementing
// primary key in SQLite.
func (*SQLiteCore) AutoIncrement() string {
	return "PRIMARY KEY AUTOINCREMENT"
}

// Escape renders v as a SQL literal, doubling embedded single quotes in
// strings and rendering values that implement encoding.TextMarshaler as their
// marshaled text.
func (s *SQLiteCore) Escape(v any) string {
	if marshaler, ok := v.(encoding.TextMarshaler); ok {
		str, err := marshaler.MarshalText()
		if err != nil {
			panic(fmt.Errorf("failed to escape value: %w", err))
		}
		return s.Escape(string(str))
	}
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.String {
		return "'" + strings.ReplaceAll(val.String(), "'", "''") + "'"
	}
	if val.CanInt() || val.CanUint() || val.CanFloat() {
		return fmt.Sprint(v)
	}
	if val.Kind() == reflect.Bool {
		if val.Bool() {
			return "1"
		} else {
			return "0"
		}
	}

	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("failed to escape value: %w", err))
	}
	return s.Escape(string(b))
}

// Binding returns the SQLite binding placeholder.
func (*SQLiteCore) Binding() string {
	return "?"
}

// Features reports the capabilities supported by SQLite, which include
// RETURNING.
func (s *SQLiteCore) Features() dialects.Features {
	return dialects.Features{
		Returning: true,
	}
}

// UseSQLite registers the SQLite dialect under the "sqlite3" and "sqlite"
// driver names.
func UseSQLite() {
	dialects.Register("sqlite3", New)
	dialects.Register("sqlite", New)
}
func init() {
	UseSQLite()
}
