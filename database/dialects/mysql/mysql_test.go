package mysql_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
	"gosalusa.com/database/dialects/mysql"
)

func TestMySQLCoreIdentifier(t *testing.T) {
	c := &mysql.MySQLCore{}
	assert.Equal(t, "*", c.Identifier("*"))
	assert.Equal(t, "`foo`", c.Identifier("foo"))
	assert.Equal(t, "`foo`.*", c.Identifier("foo.*"))
	assert.Equal(t, "`a`.`b`", c.Identifier("a.b"))
}

func TestMySQLCoreDataType(t *testing.T) {
	c := &mysql.MySQLCore{}
	cases := map[dialects.DataType]string{
		dialects.DataTypeString:   "VARCHAR(255)",
		dialects.DataTypeText:     "MEDIUMTEXT",
		dialects.DataTypeJSON:     "MEDIUMTEXT",
		dialects.DataTypeInt8:     "TINYINT",
		dialects.DataTypeInt16:    "SMALLINT",
		dialects.DataTypeInt32:    "INT",
		dialects.DataTypeInt64:    "BIGINT",
		dialects.DataTypeUInt8:    "TINYINT UNSIGNED",
		dialects.DataTypeUInt16:   "SMALLINT UNSIGNED",
		dialects.DataTypeUInt32:   "INT UNSIGNED",
		dialects.DataTypeUInt64:   "BIGINT UNSIGNED",
		dialects.DataTypeBoolean:  "BOOLEAN",
		dialects.DataTypeFloat32:  "FLOAT",
		dialects.DataTypeFloat64:  "DOUBLE",
		dialects.DataTypeDate:     "DATE",
		dialects.DataTypeDateTime: "DATETIME",
	}
	for dt, expected := range cases {
		assert.Equal(t, expected, c.DataType(dt), dt.Name)
	}
	assert.Equal(t, "blob", c.DataType(dialects.DataTypeBlob))
}

func TestMySQLCoreMisc(t *testing.T) {
	c := &mysql.MySQLCore{}
	assert.Equal(t, "CURRENT_TIMESTAMP", c.CurrentTime())
	assert.Equal(t, "PRIMARY KEY AUTO_INCREMENT", c.AutoIncrement())
	assert.Equal(t, "?", c.Binding())
}

func TestMySQLCoreEscape(t *testing.T) {
	c := &mysql.MySQLCore{}
	assert.Equal(t, "'foo'", c.Escape("foo"))
	assert.Equal(t, "'bar''s'", c.Escape("bar's"))
	assert.Equal(t, "42", c.Escape(42))
	assert.Equal(t, "4.5", c.Escape(4.5))
	assert.Equal(t, "1", c.Escape(true))
	assert.Equal(t, "0", c.Escape(false))
	assert.Equal(t, `'{"a":1}'`, c.Escape(map[string]int{"a": 1}))
	assert.Panics(t, func() {
		c.Escape(make(chan int))
	})
}

func TestMySQLNew(t *testing.T) {
	d := mysql.New()
	_, ok := d.(*generic.Generic)
	assert.True(t, ok)
	require.NotNil(t, d)
}

func TestUseMySql(t *testing.T) {
	mysql.UseMySql()
	d, err := dialects.New("mysql")
	if err != nil {
		assert.FailNow(t, "no dialect registered")
	}
	_, ok := d.(*generic.Generic)
	assert.True(t, ok)
}
