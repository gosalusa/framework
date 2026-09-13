package dialects

import (
	"errors"
	"fmt"
)

// ErrNotRegistered is returned by New when the given driver name has not been
// registered with Register.
var ErrNotRegistered = errors.New("no dialect registered")

// Features describes the SQL capabilities supported by a dialect.
type Features struct {
	// Returning indicates the dialect supports the RETURNING clause on INSERT,
	// UPDATE, and DELETE statements.
	Returning bool
}

// Dialect encodes dialect query ASTs into RawQueries.
type Dialect interface {
	EncodeSelectQuery(q *SelectQuery) (RawQuery, error)
	EncodeInsertQuery(q *InsertQuery) (RawQuery, error)
	EncodeUpdateQuery(q *UpdateQuery) (RawQuery, error)
	EncodeDeleteQuery(q *DeleteQuery) (RawQuery, error)
	EncodeCreateTableQuery(q *CreateTableQuery) (RawQuery, error)
	EncodeDropTableQuery(q *DropTableQuery) (RawQuery, error)
	EncodeAlterTableQuery(q *AlterTableQuery) (RawQuery, error)
	Features() Features
}

var dialects = map[string]func() Dialect{}

// Register associates a dialect factory with a database driver name. Concrete
// dialect packages call Register (typically from their init function) so that
// New can construct them by name.
func Register(driver string, factory func() Dialect) {
	dialects[driver] = factory
}

// New returns the dialect registered for driverName. It returns
// ErrNotRegistered wrapped if no dialect is registered for the name.
func New(driverName string) (Dialect, error) {
	f, ok := dialects[driverName]
	if !ok {
		return nil, fmt.Errorf("%w for %s", ErrNotRegistered, driverName)
	}
	return f(), nil
}
