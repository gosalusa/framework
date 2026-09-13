package builder_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/builder"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/model"
	"gosalusa.com/database/model/mixins"
	"gosalusa.com/internal/test"
)

type errorDB struct{}

func (errorDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, errors.New("exec error")
}
func (errorDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, errors.New("query error")
}
func (errorDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
func (errorDB) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, errors.New("query error")
}
func (errorDB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	return nil
}
func (errorDB) DriverName() string {
	return "sqlite3"
}

func TestBuilderNew(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "new",
			Builder:          builder.New[*test.Foo](),
			ExpectedSQLite:   "SELECT *",
			ExpectedBindings: []any{},
		},
	})
}

func TestOrWhereMethods(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "or where column",
			Builder:          NewTestBuilder().WhereColumn("a", "=", "b").OrWhereColumn("c", "=", "d"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" = \"b\" OR \"c\" = \"d\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "or where in",
			Builder:          NewTestBuilder().WhereIn("a", []any{1, 2}).OrWhereIn("b", []any{3, 4}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"a\" in (?, ?) OR \"b\" in (?, ?)",
			ExpectedBindings: []any{1, 2, 3, 4},
		},
		{
			Name:             "or where exists",
			Builder:          NewTestBuilder().WhereExists(NewTestBuilder().Select("a")).OrWhereExists(NewTestBuilder().Select("b")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE EXISTS (SELECT \"a\" FROM \"foos\") OR EXISTS (SELECT \"b\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "where not exists",
			Builder:          NewTestBuilder().WhereNotExists(NewTestBuilder().Select("a")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE NOT EXISTS (SELECT \"a\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "or where not exists",
			Builder:          NewTestBuilder().WhereNotExists(NewTestBuilder().Select("a")).OrWhereNotExists(NewTestBuilder().Select("b")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE NOT EXISTS (SELECT \"a\" FROM \"foos\") OR NOT EXISTS (SELECT \"b\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "or where subquery",
			Builder:          NewTestBuilder().Where("x", "=", 1).OrWhereSubquery(NewTestBuilder().Select("a"), "=", "b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE \"x\" = ? OR (SELECT \"a\" FROM \"foos\") = ?",
			ExpectedBindings: []any{1, "b"},
		},
		{
			Name: "or where has",
			Builder: NewTestBuilder().Where("x", "=", 1).OrWhereHas("Bar", func(q *builder.Builder) *builder.Builder {
				return q.Where("id", "=", "b")
			}),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" WHERE "x" = ? OR EXISTS (SELECT "bars".* FROM "bars" WHERE "foo_id" = "foos"."id" AND "id" = ?)`,
			ExpectedBindings: []any{1, "b"},
		},
		{
			Name:             "or where raw",
			Builder:          NewTestBuilder().WhereRaw("a = ?", 1).OrWhereRaw("b = ?", 2),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" WHERE a = ? OR b = ?",
			ExpectedBindings: []any{1, 2},
		},
	})
}

func TestHavingOrMethods(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "having column",
			Builder:          NewTestBuilder().GroupBy("a").HavingColumn("a", "=", "b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING \"a\" = \"b\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "or having column",
			Builder:          NewTestBuilder().GroupBy("a").HavingColumn("a", "=", "b").OrHavingColumn("c", "=", "d"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING \"a\" = \"b\" OR \"c\" = \"d\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "having in",
			Builder:          NewTestBuilder().GroupBy("a").HavingIn("a", []any{1, 2}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING \"a\" in (?, ?)",
			ExpectedBindings: []any{1, 2},
		},
		{
			Name:             "or having in",
			Builder:          NewTestBuilder().GroupBy("a").HavingIn("a", []any{1, 2}).OrHavingIn("b", []any{3}),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING \"a\" in (?, ?) OR \"b\" in (?)",
			ExpectedBindings: []any{1, 2, 3},
		},
		{
			Name:             "having exists",
			Builder:          NewTestBuilder().GroupBy("a").HavingExists(NewTestBuilder().Select("a")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING EXISTS (SELECT \"a\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "or having exists",
			Builder:          NewTestBuilder().GroupBy("a").HavingExists(NewTestBuilder().Select("a")).OrHavingExists(NewTestBuilder().Select("b")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING EXISTS (SELECT \"a\" FROM \"foos\") OR EXISTS (SELECT \"b\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "having not exists",
			Builder:          NewTestBuilder().GroupBy("a").HavingNotExists(NewTestBuilder().Select("a")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING NOT EXISTS (SELECT \"a\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "or having not exists",
			Builder:          NewTestBuilder().GroupBy("a").HavingNotExists(NewTestBuilder().Select("a")).OrHavingNotExists(NewTestBuilder().Select("b")),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING NOT EXISTS (SELECT \"a\" FROM \"foos\") OR NOT EXISTS (SELECT \"b\" FROM \"foos\")",
			ExpectedBindings: []any{},
		},
		{
			Name:             "having subquery",
			Builder:          NewTestBuilder().GroupBy("a").HavingSubquery(NewTestBuilder().Select("a"), "=", "b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING (SELECT \"a\" FROM \"foos\") = ?",
			ExpectedBindings: []any{"b"},
		},
		{
			Name:             "or having subquery",
			Builder:          NewTestBuilder().GroupBy("a").Having("x", "=", 1).OrHavingSubquery(NewTestBuilder().Select("a"), "=", "b"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING \"x\" = ? OR (SELECT \"a\" FROM \"foos\") = ?",
			ExpectedBindings: []any{1, "b"},
		},
		{
			Name: "having has",
			Builder: NewTestBuilder().GroupBy("a").HavingHas("Bar", func(q *builder.Builder) *builder.Builder {
				return q.Where("id", "=", "b")
			}),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" GROUP BY "a" HAVING EXISTS (SELECT "bars".* FROM "bars" WHERE "foo_id" = "foos"."id" AND "id" = ?)`,
			ExpectedBindings: []any{"b"},
		},
		{
			Name: "or having has",
			Builder: NewTestBuilder().GroupBy("a").Having("x", "=", 1).OrHavingHas("Bar", func(q *builder.Builder) *builder.Builder {
				return q.Where("id", "=", "b")
			}),
			ExpectedSQLite:   `SELECT "foos".* FROM "foos" GROUP BY "a" HAVING "x" = ? OR EXISTS (SELECT "bars".* FROM "bars" WHERE "foo_id" = "foos"."id" AND "id" = ?)`,
			ExpectedBindings: []any{1, "b"},
		},
		{
			Name:             "having raw",
			Builder:          NewTestBuilder().GroupBy("a").HavingRaw("count(*) > ?", 1),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING count(*) > ?",
			ExpectedBindings: []any{1},
		},
		{
			Name:             "or having raw",
			Builder:          NewTestBuilder().GroupBy("a").HavingRaw("a = ?", 1).OrHavingRaw("b = ?", 2),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\" HAVING a = ? OR b = ?",
			ExpectedBindings: []any{1, 2},
		},
	})
}

func TestAddGroupBy(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "add group by",
			Builder:          NewTestBuilder().GroupBy("a", "b").AddGroupBy("c"),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\" GROUP BY \"a\", \"b\", \"c\"",
			ExpectedBindings: []any{},
		},
	})
}

func TestUnordered(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "unordered",
			Builder:          NewTestBuilder().OrderBy("a").OrderByDesc("b").Unordered(),
			ExpectedSQLite:   "SELECT \"foos\".* FROM \"foos\"",
			ExpectedBindings: []any{},
		},
	})
}

func TestAddSelect(t *testing.T) {
	test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
		{
			Name:             "add select",
			Builder:          NewTestBuilder().Select("a").AddSelect("b"),
			ExpectedSQLite:   "SELECT \"a\", \"b\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "add select subquery",
			Builder:          NewTestBuilder().AddSelectSubquery(NewTestBuilder().Select("a"), "test"),
			ExpectedSQLite:   "SELECT \"foos\".*, (SELECT \"a\" FROM \"foos\") AS \"test\" FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "select function",
			Builder:          NewTestBuilder().SelectFunction("count", "*"),
			ExpectedSQLite:   "SELECT count(*) FROM \"foos\"",
			ExpectedBindings: []any{},
		},
		{
			Name:             "add select function",
			Builder:          NewTestBuilder().Select("a").AddSelectFunction("max", "b"),
			ExpectedSQLite:   "SELECT \"a\", max(\"b\") FROM \"foos\"",
			ExpectedBindings: []any{},
		},
	})
}

func TestConditionsBuild(t *testing.T) {
	c := builder.NewConditionBuilder().Where("a", "=", "b")
	assert.Len(t, c.Build(), 1)
}

func TestForeignKeyEqual(t *testing.T) {
	fk := &builder.ForeignKey{LocalKey: "a", RelatedTable: "t", RelatedKey: "b"}
	assert.True(t, fk.Equal(&builder.ForeignKey{LocalKey: "a", RelatedTable: "t", RelatedKey: "b"}))
	assert.False(t, fk.Equal(&builder.ForeignKey{LocalKey: "x", RelatedTable: "t", RelatedKey: "b"}))
}

func TestActiveScopes(t *testing.T) {
	assert.Empty(t, NewTestBuilder().ActiveScopes())
	assert.Empty(t, builder.NewBuilder().ActiveScopes())
}

func TestDump(t *testing.T) {
	assert.NotNil(t, NewTestBuilder().Dump())
}

func TestCount(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1, Name: "foo1"})
		MustSave(tx, &test.Foo{ID: 2, Name: "foo2"})

		count, err := NewTestBuilder().Count(tx)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestLoadOne(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1, Name: "foo1"})

		var foo test.Foo
		err := NewTestBuilder().LoadOne(tx, &foo)
		assert.NoError(t, err)
		assert.Equal(t, 1, foo.ID)
	})
}

func TestFind(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1, Name: "foo1"})
		MustSave(tx, &test.Foo{ID: 2, Name: "foo2"})

		foo, err := builder.From[*test.Foo]().Find(tx, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, foo.ID)
	})
}

type compoundKeyModel struct {
	model.BaseModel
	ID  int `db:"id,primary"`
	ID2 int `db:"id2,primary"`
}

func (compoundKeyModel) Table() string { return "compound_key_models" }

func TestFind_error_compound_key(t *testing.T) {
	_, err := builder.From[*compoundKeyModel]().Find(nil, 1)
	assert.Error(t, err)
}

func TestQueryError(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		_, err := builder.From[*test.Foo]().WhereRaw("not valid sql").Get(tx)
		assert.Error(t, err)
		var qe *builder.QueryError
		assert.True(t, errors.As(err, &qe))
		assert.NotEmpty(t, qe.Error())
		assert.NotNil(t, errors.Unwrap(err))
	})
}

func TestLoadMissing(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		foo := MustSave(tx, &test.Foo{ID: 1})
		MustSave(tx, &test.Bar{ID: 2, FooID: 1})

		foos := []*test.Foo{foo}
		err := builder.LoadMissing(tx, foos, "Bar")
		assert.NoError(t, err)
		assert.True(t, foo.Bar.Loaded())

		err = builder.LoadMissing(tx, foos, "Bar")
		assert.NoError(t, err)
	})
}

func TestRelationshipQuery(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		foo := MustSave(tx, &test.Foo{ID: 5})

		test.QueryTest(t, []test.Case[dialects.QueryBuilder]{
			{
				Name:             "has one query",
				Builder:          foo.Bar.Query(),
				ExpectedSQLite:   "SELECT \"bars\".* FROM \"bars\" WHERE \"foo_id\" = ?",
				ExpectedBindings: []any{5},
			},
		})
	})
}

func TestEachError(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1})
		MustSave(tx, &test.Foo{ID: 2})

		err := NewTestBuilder().Each(tx, func(v *test.Foo) error {
			return errors.New("boom")
		})
		assert.Error(t, err)
	})
}

func TestSoftDeleteDelete(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.FooSoftDelete{ID: 1, Name: "a"})
		MustSave(tx, &test.FooSoftDelete{ID: 2, Name: "b"})

		err := builder.From[*test.FooSoftDelete]().Where("id", "=", 1).Delete(tx)
		assert.NoError(t, err)

		foos, err := builder.From[*test.FooSoftDelete]().Get(tx)
		assert.NoError(t, err)
		assert.Len(t, foos, 1)
		assert.Equal(t, 2, foos[0].ID)
	})
}

func TestActiveScopesGlobal(t *testing.T) {
	foos := builder.From[*test.FooSoftDelete]()
	assert.Len(t, foos.ActiveScopes(), 1)
	assert.Empty(t, foos.WithoutGlobalScope(mixins.SoftDeleteScope).ActiveScopes())
}

func TestModelBuilderLoad(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		MustSave(tx, &test.Foo{ID: 1, Name: "foo1"})

		var foos []*test.Foo
		err := NewTestBuilder().Load(tx, &foos)
		assert.NoError(t, err)
		assert.Len(t, foos, 1)
	})
}

func TestFirstError(t *testing.T) {
	test.Run(t, "", func(t *testing.T, tx *sqlx.Tx) {
		_, err := builder.From[*test.Foo]().WhereRaw("not valid sql").First(tx)
		assert.Error(t, err)
	})
}

func TestUpdateEmpty(t *testing.T) {
	err := builder.From[*test.Foo]().Update(nil, builder.Updates{})
	assert.NoError(t, err)
}

func TestUpdateError(t *testing.T) {
	err := builder.From[*test.Foo]().Update(errorDB{}, builder.Updates{"id": 1})
	assert.Error(t, err)
}

func TestDeleteError(t *testing.T) {
	err := builder.From[*test.Foo]().Where("id", "=", 1).Delete(errorDB{})
	assert.Error(t, err)
}
