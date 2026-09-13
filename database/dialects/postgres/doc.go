// Package postgres implements the dialects.Dialect interface for PostgreSQL.
//
// Importing the package registers the dialect with dialects.Register under the
// "postgres" driver name. New returns a dialect backed by PosgtgresCore, and
// Config describes how to connect to a PostgreSQL database.
package postgres