package builder

import "gosalusa.com/database/dialects"

// ForUpdate adds a FOR UPDATE clause to the query, locking the selected rows
// until the transaction is committed.
func (b *Builder) ForUpdate() *Builder {
	b.query.ForUpdate = dialects.ForUpdateDefault
	return b
}

// ForUpdateSkipLocked adds a FOR UPDATE SKIP LOCKED clause to the query,
// locking the selected rows while skipping any rows that are already locked.
func (b *Builder) ForUpdateSkipLocked() *Builder {
	b.query.ForUpdate = dialects.ForUpdateSkipLocked
	return b
}
