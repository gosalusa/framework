package builder

import (
	"errors"

	"gosalusa.com/database"
	"gosalusa.com/database/dialects"
)

var (
	// ErrNoUpdates is returned when an update statement is built with no
	// columns to update.
	ErrNoUpdates = errors.New("no updates found")
)

// Updates maps column names to the values they should be set to.
type Updates map[string]any

// Update updates the records matched by the query with the given columns.
func (b *ModelBuilder[T]) Update(tx database.DB, updates Updates) error {
	return b.builder.Update(tx, updates)
}

// Update updates the records matched by the query with the given columns.
func (b *Builder) Update(tx database.DB, updates Updates) error {
	if len(updates) == 0 {
		return nil
	}
	d, err := dialects.New(tx.DriverName())
	if err != nil {
		return err
	}
	r, err := d.EncodeUpdateQuery(b.UpdateQuery(updates))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(b.ctx, r.SQL, r.Bindings...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateReturning updates the records matched by the query with the given
// columns and returns the updated records that were matched.
func (b *ModelBuilder[T]) UpdateReturning(tx database.DB, updates Updates) ([]T, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	d, err := dialects.New(tx.DriverName())
	if err != nil {
		return nil, err
	}

	// b.Query().Select.Columns
	q := b.UpdateQuery(updates)
	q.Returning = []dialects.Column{
		{Column: "*"},
	}

	r, err := d.EncodeUpdateQuery(q)
	if err != nil {
		return nil, err
	}

	result := []T{}

	err = load(b.Context(), tx, r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateQuery returns the dialects.UpdateQuery that Update will execute.
func (b *ModelBuilder[T]) UpdateQuery(updates Updates) *dialects.UpdateQuery {
	return b.builder.UpdateQuery(updates)
}

// UpdateQuery returns the dialects.UpdateQuery that Update will execute.
func (b *Builder) UpdateQuery(updates Updates) *dialects.UpdateQuery {
	return &dialects.UpdateQuery{
		Table:  b.GetTable(),
		Values: updates,
		Wheres: b.wheres.conditions,
	}
}
