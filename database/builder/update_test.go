package builder_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestUpdater(t *testing.T) {
	test.UpdateQueryTest(t, []test.Case[*dialects.UpdateQuery]{
		{
			Name:             "Update all",
			Builder:          NewTestBuilder().UpdateQuery(builder.Updates{"id": 1}),
			ExpectedSQLite:   `UPDATE "foos" SET "id" = ?`,
			ExpectedBindings: []any{1},
		},
		{
			Name:             "Update all multi",
			Builder:          NewTestBuilder().UpdateQuery(builder.Updates{"id": 1, "foo": "bar"}),
			ExpectedSQLite:   `UPDATE "foos" SET "foo" = ?, "id" = ?`,
			ExpectedBindings: []any{"bar", 1},
		},
		{
			Name:             "Update where",
			Builder:          NewTestBuilder().Where("id", "=", 5).UpdateQuery(builder.Updates{"id": 1}),
			ExpectedSQLite:   `UPDATE "foos" SET "id" = ? WHERE "id" = ?`,
			ExpectedBindings: []any{1, 5},
		},
		{
			Name:               "Update where multi",
			Builder:            NewTestBuilder().Where("id", "=", 5).UpdateQuery(builder.Updates{"id": 1, "foo": "bar"}),
			ExpectedSQLite:     `UPDATE "foos" SET "foo" = ?, "id" = ? WHERE "id" = ?`,
			ExpectedMySQL:      "UPDATE `foos` SET `foo` = ?, `id` = ? WHERE `id` = ?",
			ExpectedPostgreSQL: `UPDATE "foos" SET "foo" = $1, "id" = $2 WHERE "id" = $3`,
			ExpectedBindings:   []any{"bar", 1, 5},
		},
	})
}

func TestUpdate(t *testing.T) {
	test.Run(t, "update", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1, Name: "test1"})
		MustSave(tx, &test.Foo{ID: 2, Name: "test2"})

		err := builder.From[*test.Foo]().Where("id", "=", 1).Update(tx, builder.Updates{"name": "new test1"})
		assert.NoError(t, err)

		foos, err := builder.From[*test.Foo]().OrderBy("id").Get(tx)
		assert.NoError(t, err)
		assert.Len(t, foos, 2)
		assert.Equal(t, 1, foos[0].ID)
		assert.Equal(t, "new test1", foos[0].Name)
	})
}
