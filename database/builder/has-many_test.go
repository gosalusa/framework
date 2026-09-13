package builder_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/test"
)

func TestHasMany_Load(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		foos := []*test.Foo{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		}
		for _, f := range foos {
			assert.NoError(t, model.Save(tx, f))
		}
		bars := []*test.Bar{
			{ID: 2, FooID: 1},
			{ID: 3, FooID: 1},
			{ID: 4, FooID: 2},
			{ID: 5, FooID: 2},
			{ID: 6, FooID: 3},
			{ID: 7, FooID: 3},
		}
		for _, b := range bars {
			assert.NoError(t, model.Save(tx, b))
		}

		err := builder.Load(tx, foos, "Bars")
		assert.NoError(t, err)

		assertJsonEqual(t, `[
			{
			  "id": 1,
			  "name": "",
			  "bar": null,
			  "bars": [
				{ "id": 2, "foo_id": 1, "foo": null },
				{ "id": 3, "foo_id": 1, "foo": null }
			  ]
			},
			{
			  "id": 2,
			  "name": "",
			  "bar": null,
			  "bars": [
				{ "id": 4, "foo_id": 2, "foo": null },
				{ "id": 5, "foo_id": 2, "foo": null }
			  ]
			},
			{
			  "id": 3,
			  "name": "",
			  "bar": null,
			  "bars": [
				{ "id": 6, "foo_id": 3, "foo": null },
				{ "id": 7, "foo_id": 3, "foo": null }
			  ]
			}
		  ]`, foos)
		for _, foo := range foos {
			assert.True(t, foo.Bars.Loaded())
		}
	})
}

func BenchmarkHasManyLoad(b *testing.B) {
	test.RunBenchmark(b, "", func(t *testing.B, tx *sqlx.Tx) {
		fooCount := 100
		barPerFoo := 100
		foos := make([]*test.Foo, fooCount)
		for i := 0; i < fooCount; i++ {
			f := &test.Foo{}
			foos[i] = f
			MustSave(tx, f)
			for j := 0; j < barPerFoo; j++ {
				MustSave(tx, &test.Bar{FooID: f.ID})
			}
		}

		var err error
		for i := 0; i < b.N; i++ {
			err = builder.Load(tx, foos, "Bars")
			assert.NoError(t, err)
		}

	})
}
