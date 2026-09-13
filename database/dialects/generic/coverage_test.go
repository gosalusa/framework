package generic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/generic"
)

func TestGeneric_EncodeDeleteQuery(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeDeleteQuery(&dialects.DeleteQuery{
		Table:  "foos",
		Wheres: []dialects.Condition{{Column: dialects.Column{Column: "id"}, Operator: "=", Value: 5}},
	})
	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM `foos` WHERE `id` = ?", r.SQL)
	assert.Equal(t, []any{5}, r.Bindings)
}

func TestGeneric_EncodeDropTableQuery(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeDropTableQuery(&dialects.DropTableQuery{Table: "foos"})
	require.NoError(t, err)
	assert.Equal(t, "DROP TABLE `foos`", r.SQL)

	r, err = g.EncodeDropTableQuery(&dialects.DropTableQuery{Table: "foos", IfExists: true})
	require.NoError(t, err)
	assert.Equal(t, "DROP TABLE IF EXISTS `foos`", r.SQL)
}

func TestGeneric_EncodeForeignKey(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeForeignKey(&dialects.ForeignKey{
		Name:           "fk",
		Columns:        []string{"foo_id"},
		ForeignTable:   "foos",
		ForeignColumns: []string{"id"},
	})
	require.NoError(t, err)
	assert.Equal(t, "CONSTRAINT `fk` FOREIGN KEY (`foo_id`) REFERENCES `foos` (`id`)", r.SQL)
}

func TestGeneric_EncodeIndex(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeIndex(&dialects.Index{
		Name:    "idx",
		Table:   "foos",
		Columns: []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, "CREATE INDEX IF NOT EXISTS `idx` ON `foos` (`a`, `b`)", r.SQL)

	r, err = g.EncodeIndex(&dialects.Index{
		Name:    "idx",
		Table:   "foos",
		Columns: []string{"a"},
		Unique:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, "CREATE UNIQUE INDEX IF NOT EXISTS `idx` ON `foos` (`a`)", r.SQL)
}

func TestGeneric_EncodeAlterTableQuery(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeAlterTableQuery(&dialects.AlterTableQuery{
		Table:         "foos",
		DropColumns:   []string{"old"},
		ModifyColumns: []dialects.ColumnDefinition{{Name: "a", Datatype: dialects.DataTypeInt32}},
		AddColumns:    []dialects.ColumnDefinition{{Name: "b", Datatype: dialects.DataTypeString}},
		ForeignKeys:   []dialects.ForeignKey{{Name: "fk", Columns: []string{"foo_id"}, ForeignTable: "bars", ForeignColumns: []string{"id"}}},
		Indexes:       []dialects.Index{{Name: "idx", Table: "foos", Columns: []string{"a"}}},
	})
	require.NoError(t, err)
	assert.Equal(t,
		"ALTER TABLE `foos` DROP COLUMN `old`;"+
			" ALTER TABLE `foos` MODIFY COLUMN `a` int32 NOT NULL;"+
			" ALTER TABLE `foos` ADD `b` string NOT NULL;"+
			" ALTER TABLE `foos` ADD CONSTRAINT `fk` FOREIGN KEY (`foo_id`) REFERENCES `bars` (`id`);"+
			" CREATE INDEX IF NOT EXISTS `idx` ON `foos` (`a`);",
		r.SQL)
}

func TestGeneric_EncodeAny(t *testing.T) {
	g := generic.New(&testCore{})
	t.Run("query builder", func(t *testing.T) {
		r, err := g.EncodeAny(&selectQuery{&dialects.SelectQuery{
			Select: dialects.Select{Columns: []dialects.Column{{Column: "id"}}},
			From:   "foos",
		}})
		require.NoError(t, err)
		assert.Equal(t, "(SELECT `id` FROM `foos`)", r.SQL)
	})
	t.Run("conditions", func(t *testing.T) {
		r, err := g.EncodeAny([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Operator: "=", Value: 1},
		})
		require.NoError(t, err)
		assert.Equal(t, "(`a` = ?)", r.SQL)
		assert.Equal(t, []any{1}, r.Bindings)
	})
	t.Run("column", func(t *testing.T) {
		r, err := g.EncodeAny(dialects.Column{Column: "col"})
		require.NoError(t, err)
		assert.Equal(t, "`col`", r.SQL)
	})
	t.Run("raw query", func(t *testing.T) {
		r, err := g.EncodeAny(dialects.Raw("RAW", 1))
		require.NoError(t, err)
		assert.Equal(t, "RAW", r.SQL)
		assert.Equal(t, []any{1}, r.Bindings)
	})
	t.Run("raw string", func(t *testing.T) {
		r, err := g.EncodeAny(dialects.RawString("rawstring"))
		require.NoError(t, err)
		assert.Equal(t, "rawstring", r.SQL)
	})
	t.Run("literal", func(t *testing.T) {
		r, err := g.EncodeAny(5)
		require.NoError(t, err)
		assert.Equal(t, "?", r.SQL)
		assert.Equal(t, []any{5}, r.Bindings)
	})
}

type selectQuery struct {
	q *dialects.SelectQuery
}

func (s *selectQuery) Query() *dialects.SelectQuery { return s.q }

func TestGeneric_EncodeConditionsErrors(t *testing.T) {
	g := generic.New(&testCore{})
	t.Run("no operator with column", func(t *testing.T) {
		_, err := g.EncodeConditions([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Value: 1},
		})
		assert.Error(t, err)
	})
	t.Run("nil with bad operator", func(t *testing.T) {
		_, err := g.EncodeConditions([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Operator: ">", Value: nil},
		})
		assert.Error(t, err)
	})
	t.Run("nil IS NULL", func(t *testing.T) {
		r, err := g.EncodeConditions([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Operator: "=", Value: nil},
		})
		require.NoError(t, err)
		assert.Equal(t, "`a` IS NULL", r.SQL)
	})
	t.Run("nil IS NOT NULL", func(t *testing.T) {
		r, err := g.EncodeConditions([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Operator: "!=", Value: nil},
		})
		require.NoError(t, err)
		assert.Equal(t, "`a` IS NOT NULL", r.SQL)
	})
	t.Run("in list", func(t *testing.T) {
		r, err := g.EncodeConditions([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Operator: "in", Value: []any{1, 2}},
		})
		require.NoError(t, err)
		assert.Equal(t, "`a` in (?, ?)", r.SQL)
		assert.Equal(t, []any{1, 2}, r.Bindings)
	})
	t.Run("or", func(t *testing.T) {
		r, err := g.EncodeConditions([]dialects.Condition{
			{Column: dialects.Column{Column: "a"}, Operator: "=", Value: 1},
			{Column: dialects.Column{Column: "b"}, Operator: "=", Value: 2, Or: true},
		})
		require.NoError(t, err)
		assert.Equal(t, "`a` = ? OR `b` = ?", r.SQL)
	})
	t.Run("columnless condition", func(t *testing.T) {
		r, err := g.EncodeConditions([]dialects.Condition{
			{Operator: "EXISTS", Value: &selectQuery{&dialects.SelectQuery{
				Select: dialects.Select{Columns: []dialects.Column{{Column: "id"}}},
				From:   "foos",
			}}},
		})
		require.NoError(t, err)
		assert.Equal(t, "EXISTS (SELECT `id` FROM `foos`)", r.SQL)
	})
	t.Run("raw where", func(t *testing.T) {
		r, err := g.EncodeConditions([]dialects.Condition{
			{Value: dialects.Raw("foo = ?", 3)},
		})
		require.NoError(t, err)
		assert.Equal(t, "foo = ?", r.SQL)
		assert.Equal(t, []any{3}, r.Bindings)
	})
}

func TestGeneric_EncodeSelectsDistinct(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeSelects(&dialects.Select{
		Distinct: true,
		Columns:  []dialects.Column{{Column: "a"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "DISTINCT `a`", r.SQL)
}

func TestGeneric_EncodeSelectsEmpty(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeSelects(&dialects.Select{})
	require.NoError(t, err)
	assert.Equal(t, "", r.SQL)
}

func TestGeneric_EncodeOrderByDesc(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeOrderBy([]dialects.OrderColumn{
		{Column: "a"},
		{Column: "b", Descending: true},
	})
	require.NoError(t, err)
	assert.Equal(t, "ORDER BY `a`, `b` DESC", r.SQL)
}

func TestGeneric_EncodeCreateTableQueryFull(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeCreateTableQuery(&dialects.CreateTableQuery{
		IfNotExists: true,
		Table:       "foos",
		Columns: []dialects.ColumnDefinition{
			{Name: "id", Datatype: dialects.DataTypeInt32, Primary: true},
			{Name: "name", Datatype: dialects.DataTypeString, Unique: true},
			{Name: "auto", Datatype: dialects.DataTypeInt32, AutoIncrement: true},
			{Name: "def", Datatype: dialects.DataTypeString, DefaultValue: "v"},
			{Name: "now", Datatype: dialects.DataTypeDateTime, DefaultCurrentTime: true},
		},
		PrimaryKeys: []string{"id"},
		ForeignKeys: []dialects.ForeignKey{
			{Name: "fk", Columns: []string{"foo_id"}, ForeignTable: "bars", ForeignColumns: []string{"id"}},
		},
		Indexes: []dialects.Index{{Name: "idx", Table: "foos", Columns: []string{"name"}}},
	})
	require.NoError(t, err)
	assert.Equal(t,
		"CREATE TABLE IF NOT EXISTS `foos` ("+
			"`id` int32 PRIMARY KEY NOT NULL,"+
			" `name` string NOT NULL UNIQUE,"+
			" `auto` int32 PRIMARY KEY AUTO_INCREMENT NOT NULL,"+
			" `def` string NOT NULL DEFAULT,"+
			" `now` date-time NOT NULL DEFAULT CURRENT_TIMESTAMP,"+
			" PRIMARY KEY (`id`),"+
			" CONSTRAINT `fk` FOREIGN KEY (`foo_id`) REFERENCES `bars` (`id`));"+
			" CREATE INDEX IF NOT EXISTS `idx` ON `foos` (`name`);",
		r.SQL)
}

func TestGeneric_EncodeColumnDefinitionDefaultCurrentTime(t *testing.T) {
	g := generic.New(&testCore{})
	r, err := g.EncodeColumnDefinition(&dialects.ColumnDefinition{
		Name:               "t",
		Datatype:           dialects.DataTypeDateTime,
		DefaultCurrentTime: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "`t` date-time NOT NULL DEFAULT CURRENT_TIMESTAMP", r.SQL)
}
