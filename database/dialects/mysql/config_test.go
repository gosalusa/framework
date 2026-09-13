package mysql_test

import (
	"testing"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects/mysql"
)

func TestSimpleConfig(t *testing.T) {
	c := &mysql.SimpleConfig{
		Username: "user",
		Password: "pass",
		Host:     "localhost:3306",
		Database: "db",
	}
	assert.Equal(t, "mysql", c.DriverName())
	assert.Equal(t, "user:pass@/db?multiStatements=true&parseTime=true", c.DataSourceName())
}

func TestNewMySQLConfig(t *testing.T) {
	inner := gomysql.NewConfig()
	inner.Net = "tcp"
	inner.User = "u"
	inner.Passwd = "p"
	inner.Addr = "h"
	inner.DBName = "d"
	c := mysql.NewMySQLConfig(inner)
	require.NotNil(t, c)
	assert.Equal(t, "mysql", c.DriverName())
	assert.Contains(t, c.DataSourceName(), "u:p@tcp(h)/d")
}

func TestConfig(t *testing.T) {
	c := mysql.NewMySQLConfig(&gomysql.Config{
		Net:    "tcp",
		User:   "user",
		Passwd: "pass",
		Addr:   "localhost:3306",
		DBName: "db",
	})
	assert.Equal(t, "mysql", c.DriverName())
	assert.Contains(t, c.DataSourceName(), "user:pass@tcp(localhost:3306)/db")
}
