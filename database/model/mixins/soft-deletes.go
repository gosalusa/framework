package mixins

import (
	"time"

	"gosalusa.com/database"
	"gosalusa.com/database/builder"
)

// SoftDelete adds a deleted_at column to a model and soft deletes records
// instead of removing them. Embed it in a model to hide soft-deleted rows from
// every query and to turn Delete into an update of deleted_at.
type SoftDelete struct {
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

// Scopes returns the model's global scopes, which applies SoftDeleteScope to
// every query.
func (f *SoftDelete) Scopes() []*builder.Scope {
	return []*builder.Scope{
		SoftDeleteScope,
	}
}

// SoftDeleteScope is the global scope applied by SoftDelete. It filters out
// rows where deleted_at is set and converts Delete queries into an update of
// deleted_at to the current time.
var SoftDeleteScope = &builder.Scope{
	Name: "soft-deletes",
	Query: func(b *builder.Builder) *builder.Builder {
		return b.Where(b.GetTable()+".deleted_at", "=", nil)
	},
	Delete: func(next func(q *builder.Builder, tx database.DB) error) func(q *builder.Builder, tx database.DB) error {
		return func(q *builder.Builder, tx database.DB) error {
			return q.Update(tx, builder.Updates{
				"deleted_at": time.Now(),
			})
		}
	},
}
