package postgres

import (
	"fmt"

	_ "github.com/lib/pq"
)

// Config describes how to connect to a PostgreSQL database. DisableSSL opts
// out of SSL.
type Config struct {
	Username   string
	Password   string
	Host       string
	Database   string
	DisableSSL bool
}

// DriverName returns the database driver name, "postgres".
func (c *Config) DriverName() string {
	return "postgres"
}

// DataSourceName returns the key=value DSN used to connect to the database.
func (c *Config) DataSourceName() string {
	ssl := ""
	if c.DisableSSL {
		ssl = "sslmode=disable"
	}
	return fmt.Sprintf("host=%s dbname=%s user=%s password=%s %s", c.Host, c.Database, c.Username, c.Password, ssl)
}
