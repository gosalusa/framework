package model

import (
	"context"

	"gosalusa.com/database"
)

// Contexter is implemented by models that carry the context they were loaded
// or saved with.
type Contexter interface {
	Context() context.Context
}

// Model is implemented by anything that can be persisted to the database.
type Model interface {
	// InDatabase reports whether the model has already been loaded from or
	// saved to the database. It controls whether Save inserts a new row or
	// updates an existing one.
	InDatabase() bool
}

// BaseModel provides the default state and hook implementations for models.
// Embed it as the first field in a model struct to satisfy Model and to track
// whether the model is in the database and the context it was operated with.
type BaseModel struct {
	inDatabase bool
	ctx        context.Context
}

var _ Model = &BaseModel{}

// InDatabase reports whether the model has been loaded from or saved to the
// database.
func (m *BaseModel) InDatabase() bool {
	return m.inDatabase
}

// Context returns the context the model was last loaded or saved with, or nil
// if it has never been used with a context.
func (m *BaseModel) Context() context.Context {
	return m.ctx
}

// AfterLoad implements hooks.AfterLoader by marking the model as in the
// database and storing the context it was loaded with.
func (m *BaseModel) AfterLoad(ctx context.Context, tx database.DB) error {
	m.inDatabase = true
	m.ctx = ctx
	return nil
}

// AfterSave implements hooks.AfterSaver by marking the model as in the
// database and storing the context it was saved with.
func (m *BaseModel) AfterSave(ctx context.Context, tx database.DB) error {
	m.inDatabase = true
	m.ctx = ctx
	return nil
}
