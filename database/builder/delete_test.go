package builder_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestDeleter(t *testing.T) {
	test.DeleteQueryTest(t, []test.Case[dialects.DeleteQueryBuilder]{
		{
			Name:             "delete all",
			Builder:          NewTestBuilder(),
			ExpectedSQLite:   "DELETE FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "delete where",
			Builder:          NewTestBuilder().Where("id", "=", 5),
			ExpectedSQLite:   "DELETE FROM \"foos\" WHERE \"id\" = ?",
			ExpectedBindings: []any{5},
		},
	})
}

func TestDelete(t *testing.T) {
	test.Run(t, "delete", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1, Name: "test1"})
		MustSave(tx, &test.Foo{ID: 2, Name: "test2"})

		err := builder.From[*test.Foo]().Where("id", "=", 1).Delete(tx)
		assert.NoError(t, err)

		foos, err := builder.From[*test.Foo]().Get(tx)
		assert.NoError(t, err)
		assert.Len(t, foos, 1)
		assert.Equal(t, 2, foos[0].ID)
	})
}
