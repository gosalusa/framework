package mixins

import (
	"context"
	"time"

	"gosalusa.com/database"
	"gosalusa.com/database/hooks"
)

// Timestamps adds created_at and updated_at columns to a model and keeps them
// populated on every save. Embed it in a model to get its BeforeSave behavior
// automatically.
type Timestamps struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

var _ hooks.BeforeSaver = (*Timestamps)(nil)

// BeforeSave implements hooks.BeforeSaver. It sets CreatedAt to the current
// time if it is zero and always sets UpdatedAt to the current time.
func (t *Timestamps) BeforeSave(ctx context.Context, tx database.DB) error {
	now := time.Now()
	if (t.CreatedAt.Equal(time.Time{})) {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	return nil
}
