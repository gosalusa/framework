package generic

import (
	"errors"

	"gosalusa.com/database/dialects"
)

// ErrUnkownExprType is returned when a value has a type that cannot be
// encoded as SQL.
var ErrUnkownExprType = errors.New("unknown dialects.Expr type")

// Core provides the database-specific behavior used by Generic to render SQL:
// identifier quoting, data type names, current-time literals, auto-increment
// clauses, literal escaping, binding placeholders, and capability flags.
type Core interface {
	Identifier(string) string
	DataType(dialects.DataType) string
	CurrentTime() string
	AutoIncrement() string
	Escape(v any) string
	Binding() string
	Features() dialects.Features
}

// Generic encodes dialects query ASTs into RawQueries, using a Core for all
// database-specific syntax.
type Generic struct {
	core Core
}

var _ dialects.Dialect = (*Generic)(nil)

// New returns a Generic dialect backed by core.
func New(c Core) *Generic {
	return &Generic{core: c}
}

// Features implements [dialects.Dialect].
func (g *Generic) Features() dialects.Features {
	return g.core.Features()
}
