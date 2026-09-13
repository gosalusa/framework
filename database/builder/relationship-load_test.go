package builder_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/internal/test"
)

func TestMulti(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		foos := []*test.Foo{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		}
		for _, f := range foos {
			MustSave(tx, f)
			MustSave(tx, &test.Bar{ID: f.ID + 3, FooID: f.ID})
		}

		err := builder.Load(tx, foos, "Bar")
		assert.NoError(t, err)

		err = builder.Load(tx, foos, "Bars")
		assert.NoError(t, err)

		for _, foo := range foos {
			assert.True(t, foo.Bar.Loaded(), "bar not loaded")

			assert.True(t, foo.Bars.Loaded(), "bars not loaded")
		}
	})
}

func TestLoad_deep(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		foo := &test.Foo{ID: 1}
		MustSave(tx, foo)
		MustSave(tx, &test.Bar{ID: 2, FooID: 1})

		err := builder.Load(tx, foo, "Bar.Foo.Bar")
		assert.NoError(t, err)

		bar, ok := foo.Bar.Value()
		assert.True(t, ok, "bar not loaded")
		foo2, ok := bar.Foo.Value()
		assert.True(t, ok, "bar.foo not loaded")
		bar2, ok := foo2.Bar.Value()
		assert.True(t, ok, "bar.foo.bar not loaded")

		assert.Equal(t, bar.ID, bar2.ID)
		assert.Equal(t, foo.ID, foo2.ID)

	})
}
