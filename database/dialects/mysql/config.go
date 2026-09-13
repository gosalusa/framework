package mysql

import (
	"github.com/go-sql-driver/mysql"
)

// SimpleConfig builds a MySQL data source from plain connection settings.
type SimpleConfig struct {
	Username string
	Password string
	Host     string
	Database string
}

// DriverName returns the database driver name, "mysql".
func (c *SimpleConfig) DriverName() string {
	return "mysql"
}

// DataSourceName returns the DSN for the connection settings. It enables
// multi-statement execution and time parsing.
func (c *SimpleConfig) DataSourceName() string {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = c.Username
	mysqlCfg.Passwd = c.Password
	mysqlCfg.Addr = c.Host
	mysqlCfg.DBName = c.Database
	mysqlCfg.MultiStatements = true
	mysqlCfg.ParseTime = true
	return mysqlCfg.FormatDSN()
}

// Config wraps an upstream go-sql-driver/mysql Config.
type Config struct {
	cfg *mysql.Config
}

// NewMySQLConfig returns a Config that wraps cfg.
func NewMySQLConfig(cfg *mysql.Config) *Config {
	return &Config{
		cfg: cfg,
	}
}

// DriverName returns the database driver name, "mysql".
func (c *Config) DriverName() string {
	return "mysql"
}

// DataSourceName formats the wrapped config as a DSN.
func (c *Config) DataSourceName() string {
	return c.cfg.FormatDSN()
}
