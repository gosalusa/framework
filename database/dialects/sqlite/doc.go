// Package sqlite implements the dialects.Dialect interface for SQLite.
//
// Importing the package registers the dialect with dialects.Register under both
// the "sqlite3" and "sqlite" driver names. New returns a dialect backed by
// SQLiteCore, and Config describes how to connect to a SQLite database file.
package sqlite