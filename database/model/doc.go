// Package model defines the types and helpers for persisting and loading
// database records through Go structs called models.
//
// # Defining a model
//
// A model is a struct that embeds BaseModel and uses the db struct tag to
// describe its columns. The tag's first value is the column name, and the
// remaining values are flags or modifiers:
//
//	type Foo struct {
//		model.BaseModel
//
//		ID         int       `db:"id,primary,autoincrement"`
//		Name       string    `db:"name,size:100"`
//		Slug       string    `db:"slug,unique"`
//		Score      int       `db:"score,index"`
//		Note       *string   `db:"note,nullable"`
//		ComputedAt time.Time `db:"computed_at,readonly"`
//		Password   string    `db:"-"`
//	}
//
// The supported tag values are:
//
//	primary        the column is part of the primary key
//	autoincrement  the primary key is populated by the database on insert
//	nullable       the column may store NULL
//	readonly       the column is excluded from INSERT and UPDATE statements
//	               (used for values computed by the database)
//	index          create an index on the column
//	unique         create a unique index on the column
//	type:XXX       override the column's SQL type (defaults are inferred from
//	               the Go field type)
//	size:NNN       set the column size, e.g. size:100 for a VARCHAR(100)
//	               (only used when generating migrations)
//	-              skip the field entirely: keep it in memory only
//
// If no db tag is present, the Go field name is used verbatim as the column
// name. Anonymous (embedded) structs are walked recursively, which is what
// lets mixins like mixins.Timestamps contribute their columns, and why
// BaseModel contributes nothing. Fields whose type implements
// relationship.Relationship are skipped during column collection and are
// managed separately by the builder package.
//
// The table name is the kebab of the struct name plus a trailing "s" (Foo
// becomes "foos"), unless the model implements Table() string.
//
// # Saving models
//
// Save inserts a new row or updates an existing one depending on whether the
// model has already been loaded from (or saved to) the database:
//
//	foo := &Foo{Name: "test"}
//	err := model.Save(tx, foo)
//
// Save, SaveContext, MustSave, and MustSaveContext persist a single model.
// InsertMany and InsertManyContext persist a slice of models in one statement
// and populate autoincrement primary keys on the caller's copies.
//
// Hooks run around saves: a model may implement hooks.BeforeSaver or
// hooks.AfterSaver, and embedded mixins can contribute them (for example
// mixins.Timestamps sets created_at and updated_at before every save).
//
// # Loading from HTTP requests
//
// modeldi.Register[*Foo] registers a dependency-injection provider that loads
// a Foo by ID from a request. The ID is resolved from the URL query, a
// gorilla/mux path variable, or an r.PathValue, in that order. The generated
// template wires this up in init so handlers can inject the model directly:
//
//	type fooRequest struct {
//		Foo *models.Foo `inject:"id"`
//	}
//
// # Generating migrations
//
// The migration generator is driven by the same db tags. Place this comment
// above a model struct and run go generate:
//
//	//go:generate spice generate:migration
//
// spice generates a migration that keeps the model's table in sync with the
// columns described by its db tags.
package model
