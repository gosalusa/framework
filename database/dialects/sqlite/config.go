package sqlite

import (
	// _ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

// Config describes how to connect to a SQLite database. Path is the path to
// the database file.
type Config struct {
	Path string
}

// NewConfig returns a Config for the database file at path.
func NewConfig(path string) *Config {
	return &Config{
		Path: path,
	}
}

// DriverName returns the database driver name, "sqlite3".
func (c *Config) DriverName() string {
	return "sqlite3"
}

// DataSourceName returns the path to the SQLite database file.
func (c *Config) DataSourceName() string {
	return c.Path
}
