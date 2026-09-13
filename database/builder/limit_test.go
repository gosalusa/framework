package builder_test

import (
	"testing"

	"gosalusa.com/database/dialects"
	"gosalusa.com/internal/test"
)

func TestLimit(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "limit",
			Builder:          NewTestBuilder().Limit(1),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" LIMIT ?",
			ExpectedBindings: []any{1},
		},
		{
			Name:             "offset",
			Builder:          NewTestBuilder().Offset(1),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" OFFSET ?",
			ExpectedBindings: []any{1},
		},
		{
			Name:             "limit and offset",
			Builder:          NewTestBuilder().Limit(1).Offset(2),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" LIMIT ? OFFSET ?",
			ExpectedBindings: []any{1, 2},
		},
	})
}
