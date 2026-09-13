// Package mysql implements the dialects.Dialect interface for MySQL.
//
// Importing the package registers the dialect with dialects.Register under the
// "mysql" driver name. New returns a dialect backed by MySQLCore, and the
// config types describe how to connect to a MySQL database.
package mysql