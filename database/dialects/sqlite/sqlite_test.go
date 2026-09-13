package sqlite_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/database/dialects/sqlite"
)

func TestSQLiteCoreIdentifier(t *testing.T) {
	c := &sqlite.SQLiteCore{}
	assert.Equal(t, "*", c.Identifier("*"))
	assert.Equal(t, `"foo"`, c.Identifier("foo"))
	assert.Equal(t, `"foo".*`, c.Identifier("foo.*"))
	assert.Equal(t, `"a"."b"`, c.Identifier("a.b"))
}

func TestSQLiteCoreDataType(t *testing.T) {
	c := &sqlite.SQLiteCore{}
	cases := map[dialects.DataType]string{
		dialects.DataTypeString:   "TEXT",
		dialects.DataTypeText:     "TEXT",
		dialects.DataTypeJSON:     "TEXT",
		dialects.DataTypeDate:     "TIMESTAMP",
		dialects.DataTypeDateTime: "TIMESTAMP",
		dialects.DataTypeInt32:    "INTEGER",
		dialects.DataTypeUInt32:   "INTEGER",
		dialects.DataTypeBoolean:  "INTEGER",
		dialects.DataTypeFloat32:  "FLOAT",
	}
	for dt, expected := range cases {
		assert.Equal(t, expected, c.DataType(dt), dt.Name)
	}
	assert.Equal(t, "int64", c.DataType(dialects.DataTypeInt64))
}

func TestSQLiteCoreMisc(t *testing.T) {
	c := &sqlite.SQLiteCore{}
	assert.Equal(t, "CURRENT_TIMESTAMP", c.CurrentTime())
	assert.Equal(t, "PRIMARY KEY AUTOINCREMENT", c.AutoIncrement())
	assert.Equal(t, "?", c.Binding())
}

type marshaler string

func (m marshaler) MarshalText() ([]byte, error) {
	return []byte("marshaled-" + string(m)), nil
}

type errMarshaler string

func (m errMarshaler) MarshalText() ([]byte, error) {
	return nil, errors.New("marshal error")
}

func TestSQLiteCoreEscape(t *testing.T) {
	c := &sqlite.SQLiteCore{}
	assert.Equal(t, "'foo'", c.Escape("foo"))
	assert.Equal(t, "'bar''s'", c.Escape("bar's"))
	assert.Equal(t, "42", c.Escape(42))
	assert.Equal(t, "42", c.Escape(uint(42)))
	assert.Equal(t, "4.5", c.Escape(4.5))
	assert.Equal(t, "1", c.Escape(true))
	assert.Equal(t, "0", c.Escape(false))
	assert.Equal(t, `'{"a":1}'`, c.Escape(map[string]int{"a": 1}))
	assert.Equal(t, "'marshaled-v'", c.Escape(marshaler("v")))
	assert.Panics(t, func() {
		c.Escape(errMarshaler("x"))
	})
	assert.Panics(t, func() {
		c.Escape(make(chan int))
	})
}

func TestSQLiteNew(t *testing.T) {
	d := sqlite.New()
	_, ok := d.(*generic.Generic)
	assert.True(t, ok)
	require.NotNil(t, d)
}

func TestUseSQLite(t *testing.T) {
	sqlite.UseSQLite()
	d, err := dialects.New("sqlite3")
	if err != nil {
		assert.FailNow(t, "no dialect registered")
	}
	_, ok := d.(*generic.Generic)
	assert.True(t, ok)
}
