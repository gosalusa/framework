package mysql

import (
	"encoding/json"
	"fmt"
	"strings"

	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
)

// MySQLCore implements generic.Core for MySQL. It quotes identifiers with
// backticks and uses ? bindings.
type MySQLCore struct{}

// Identifier returns s quoted as a MySQL identifier. A "*" segment is left
// unquoted.
func (*MySQLCore) Identifier(s string) string {
	if s == "*" {
		return s
	}
	parts := strings.Split(s, ".")
	for i, p := range parts {
		if p == "*" {
			continue
		}
		parts[i] = "`" + p + "`"
	}
	return strings.Join(parts, ".")
}

// DataType maps a dialects.DataType to a MySQL column type. String defaults to
// VARCHAR(255) when Size is zero.
func (*MySQLCore) DataType(t dialects.DataType) string {
	switch t.Name {
	case dialects.DataTypeString.Name:
		s := t.Size
		if s == 0 {
			s = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", s)
	case dialects.DataTypeText.Name, dialects.DataTypeJSON.Name:
		return "MEDIUMTEXT"
	case dialects.DataTypeInt8.Name:
		return "TINYINT"
	case dialects.DataTypeInt16.Name:
		return "SMALLINT"
	case dialects.DataTypeInt32.Name:
		return "INT"
	case dialects.DataTypeInt64.Name:
		return "BIGINT"
	case dialects.DataTypeUInt8.Name:
		return "TINYINT UNSIGNED"
	case dialects.DataTypeUInt16.Name:
		return "SMALLINT UNSIGNED"
	case dialects.DataTypeUInt32.Name:
		return "INT UNSIGNED"
	case dialects.DataTypeUInt64.Name:
		return "BIGINT UNSIGNED"
	case dialects.DataTypeBoolean.Name:
		return "BOOLEAN"
	case dialects.DataTypeFloat32.Name:
		return "FLOAT"
	case dialects.DataTypeFloat64.Name:
		return "DOUBLE"
	case dialects.DataTypeDate.Name:
		return "DATE"
	case dialects.DataTypeDateTime.Name:
		return "DATETIME"
	}

	return t.Name
}

// CurrentTime returns the MySQL expression for the current timestamp.
func (*MySQLCore) CurrentTime() string {
	return "CURRENT_TIMESTAMP"
}

// AutoIncrement returns the clause that makes a column an auto-incrementing
// primary key in MySQL.
func (*MySQLCore) AutoIncrement() string {
	return "PRIMARY KEY AUTO_INCREMENT"
}

// Escape renders v as a SQL literal, doubling embedded single quotes in
// strings and JSON-marshaling unknown types.
func (s *MySQLCore) Escape(v any) string {
	switch v := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return fmt.Sprint(v)
	case bool:
		if v {
			return "1"
		} else {
			return "0"
		}
	default:
		b, err := json.Marshal(v)
		if err != nil {
			panic(fmt.Errorf("failed to escape value: %w", err))
		}
		return s.Escape(string(b))
	}
}

// Binding returns the MySQL binding placeholder.
func (*MySQLCore) Binding() string {
	return "?"
}

// Features reports the capabilities supported by MySQL. MySQL does not support
// RETURNING.
func (s *MySQLCore) Features() dialects.Features {
	return dialects.Features{}
}

// New returns a dialect that renders SQL for MySQL.
func New() dialects.Dialect {
	return generic.New(&MySQLCore{})
}

// UseMySql registers the MySQL dialect under the "mysql" driver name.
func UseMySql() {
	dialects.Register("mysql", New)
}
func init() {
	UseMySql()
}
