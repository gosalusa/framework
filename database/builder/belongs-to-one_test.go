package builder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects/sqlite"
	"gosalusa.com/database/migrate"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/test"
)

func ExampleBelongsTo() {
	sqlite.UseSQLite()
	type Bar struct {
		model.BaseModel
		ID   int    `db:"id,autoincrement,primary"`
		Name string `db:"name"`
	}

	type Foo struct {
		model.BaseModel
		ID    int `db:"id,autoincrement,primary"`
		BarID int `db:"bar_id"`
		Bar   *builder.BelongsTo[*Bar]
	}

	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer db.Close()

	createFoo, err := migrate.CreateFromModel(&Foo{})
	check(err)
	err = createFoo.Run(context.Background(), db)
	check(err)
	createBar, err := migrate.CreateFromModel(&Bar{})
	check(err)
	err = createBar.Run(context.Background(), db)
	check(err)

	foo := &Foo{BarID: 1}
	err = model.Save(db, foo)
	check(err)
	bar := &Bar{ID: 1, Name: "bar name"}
	err = model.Save(db, bar)
	check(err)

	err = builder.Load(db, foo, "Bar")
	check(err)
	relatedBar, _ := foo.Bar.Value()

	fmt.Println(relatedBar.Name)

	// Output: bar name
}

func TestBelongsToLoad(t *testing.T) {
	test.Run(t, "ints", func(t *testing.T, tx *sqlx.Tx) {
		foos := []*test.Foo{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		}
		for _, f := range foos {
			model.MustSave(tx, f)
		}
		bars := []*test.Bar{
			{ID: 4, FooID: 1},
			{ID: 5, FooID: 2},
			{ID: 6, FooID: 3},
		}
		for _, b := range bars {
			model.MustSave(tx, b)
		}

		err := builder.Load(tx, bars, "Foo")
		if !assert.NoError(t, err) {
			return
		}

		for _, bar := range bars {
			assert.True(t, bar.Foo.Loaded())
			foo, ok := bar.Foo.Value()
			assert.True(t, ok)
			assert.Equal(t, bar.FooID, foo.ID)
		}
	})

	// test.Run(t, "uuids", func(t *testing.T, tx *sqlx.Tx) {
	// 	type Foo struct {
	// 		model.BaseModel
	// 		ID   int       `json:"id" db:"id,primary,autoincrement"`
	// 		Name uuid.UUID `json:"name" db:"name"`
	// 	}
	// 	type Bar struct {
	// 		model.BaseModel
	// 		FooName *uuid.UUID               `json:"foo_id" db:"foo_id"`
	// 		Foo     *builder.BelongsTo[*Foo] `json:"foo"    db:"-" foreign:"foo_id" owner:"name"`
	// 	}

	// 	bars := []*Bar{
	// 		{FooName: newUUID()},
	// 		{FooName: newUUID()},
	// 		{FooName: nil},
	// 		{FooName: nil},
	// 	}
	// 	for _, b := range bars {
	// 		if b.FooName != nil {
	// 			MustSave(tx, &Foo{Name: *b.FooName})
	// 		}
	// 	}
	// 	err := builder.Load(tx, bars, "Foo")
	// 	assert.NoError(t, err)

	// 	for i, b := range bars {
	// 		f, ok := b.Foo.Value()
	// 		assert.True(t, ok)
	// 		if i < 2 {
	// 			assert.NotNil(t, f)
	// 		} else {
	// 			assert.Nil(t, f)
	// 		}
	// 	}
	// })
}

func newUUID() *uuid.UUID {
	id := uuid.New()
	return &id
}
