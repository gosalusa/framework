// Package dialects defines the query ASTs that describe SQL statements and
// the Dialect interface that renders them into RawQueries.
//
// The query structs (SelectQuery, InsertQuery, UpdateQuery, DeleteQuery,
// CreateTableQuery, DropTableQuery, and AlterTableQuery) are plain data with
// no database-specific knowledge. They are produced by the higher-level
// builder, schema, migrate, and model packages and encoded by a Dialect into a
// RawQuery, which pairs a SQL string with its positional bind values.
//
// The generic package provides a database-agnostic encoder built on a small
// Core interface, and the sqlite, mysql, and postgres packages implement that
// interface for their respective databases. Each registers itself with Register
// so New can construct the right dialect for a database driver name:
//
//	d, err := dialects.New("sqlite3")
//	query, err := d.EncodeSelectQuery(q)
//
// MultiQuery and EncodeMultiQuery batch several queries together so they can be
// sent to the database as one multi-statement SQL string.
package dialects