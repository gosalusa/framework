// Package generic provides a database-agnostic implementation of the
// dialects.Dialect interface.
//
// A Generic encoder turns dialects query ASTs into RawQueries by assembling the
// SQL that is common to all databases and delegating the database-specific
// details to a Core: identifier quoting, data type names, literal escaping,
// binding placeholders, and capability flags.
//
// The concrete dialect packages (sqlite, mysql, and postgres) each supply a
// Core and wrap it in a Generic with New.
package generic