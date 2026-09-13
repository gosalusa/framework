package builder_test

import (
	"context"
	"testing"

	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/model/mixins"
	"gosalusa.com/internal/test"
)

type ScopeFoo struct {
	test.Foo
}

func (f *ScopeFoo) Scopes() []*builder.Scope {
	return []*builder.Scope{
		mixins.SoftDeleteScope,
	}
}

type ScopeBar struct {
	test.Bar
	ScopeFoo *builder.BelongsTo[*ScopeFoo] `db:"-" json:"foo"`
}

func TestScope(t *testing.T) {
	scopeA := &builder.Scope{
		Name: "with-a",
		Query: func(b *builder.Builder) *builder.Builder {
			return b.Where("a", "=", "b")
		},
	}
	scopeCtx := &builder.Scope{
		Name: "ctx",
		Query: func(b *builder.Builder) *builder.Builder {
			foo := b.Context().Value("foo")
			return b.Where("a", "=", foo)
		},
	}
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "scope",
			Builder:          NewTestBuilder().WithScope(scopeA),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = ?",
			ExpectedBindings: []any{"b"},
		},
		{
			Name:             "without scope",
			Builder:          NewTestBuilder().WithScope(scopeA).WithoutScope(scopeA),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "global scope",
			Builder:          builder.From[*ScopeFoo](),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"foos\".\"deleted_at\" IS NULL",
			ExpectedBindings: []any{},
		},
		{
			Name:             "without global scope",
			Builder:          builder.From[*ScopeFoo]().WithoutGlobalScope(mixins.SoftDeleteScope),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name: "global scope whereHas",
			Builder: builder.From[*ScopeBar]().WhereHas("ScopeFoo", func(q *builder.Builder) *builder.Builder {
				return q
			}),
			ExpectedSQLite:   `SELECT "bars".* FROM "bars" WHERE EXISTS (SELECT "foos".* FROM "foos" WHERE "id" = "bars"."foo_id" AND "foos"."deleted_at" IS NULL)`,
			ExpectedBindings: []any{},
		},
		{
			Name:             "access-context",
			Builder:          NewTestBuilder().WithScope(scopeCtx).WithContext(context.WithValue(context.Background(), "foo", "bar")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = ?",
			ExpectedBindings: []any{"bar"},
		},
	})
}
