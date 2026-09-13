package dialects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/dialects/sqlite"
)

func TestJoinQueries(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		r := dialects.JoinQueries([]dialects.RawQuery{})
		assert.Equal(t, "", r.SQL)
		assert.Empty(t, r.Bindings)
	})
	t.Run("one", func(t *testing.T) {
		r := dialects.JoinQueries([]dialects.RawQuery{
			{SQL: "SELECT 1", Bindings: []any{1}},
		})
		assert.Equal(t, "SELECT 1;", r.SQL)
		assert.Equal(t, []any{1}, r.Bindings)
	})
	t.Run("multiple", func(t *testing.T) {
		r := dialects.JoinQueries([]dialects.RawQuery{
			{SQL: "a", Bindings: []any{1}},
			{SQL: "b", Bindings: []any{2, 3}},
		})
		assert.Equal(t, "a; b;", r.SQL)
		assert.Equal(t, []any{1, 2, 3}, r.Bindings)
	})
}

func TestRaw(t *testing.T) {
	r := dialects.Raw("SELECT ?", 1)
	assert.Equal(t, "SELECT ?", r.SQL)
	assert.Equal(t, []any{1}, r.Bindings)
}

func TestDataTypeIsValid(t *testing.T) {
	assert.True(t, dialects.DataTypeString.IsValid())
	assert.True(t, dialects.DataTypeBlob.IsValid())
	assert.True(t, dialects.DataTypeText.IsValid())
	assert.True(t, dialects.DataTypeEnum.IsValid())
	assert.True(t, dialects.DataTypeBoolean.IsValid())
	assert.True(t, dialects.DataTypeDate.IsValid())
	assert.True(t, dialects.DataTypeDateTime.IsValid())
	assert.True(t, dialects.DataTypeFloat32.IsValid())
	assert.True(t, dialects.DataTypeFloat64.IsValid())
	assert.True(t, dialects.DataTypeInt8.IsValid())
	assert.True(t, dialects.DataTypeInt16.IsValid())
	assert.True(t, dialects.DataTypeInt32.IsValid())
	assert.True(t, dialects.DataTypeInt64.IsValid())
	assert.True(t, dialects.DataTypeUInt8.IsValid())
	assert.True(t, dialects.DataTypeUInt16.IsValid())
	assert.True(t, dialects.DataTypeUInt32.IsValid())
	assert.True(t, dialects.DataTypeUInt64.IsValid())
	assert.True(t, dialects.DataTypeJSON.IsValid())
	assert.False(t, dialects.DataType(dialects.DataType{Name: "invalid"}).IsValid())
}

type selectQueryBuilder struct {
	q *dialects.SelectQuery
}

func (b *selectQueryBuilder) Query() *dialects.SelectQuery { return b.q }

type insertQueryBuilder struct {
	q *dialects.InsertQuery
}

func (b *insertQueryBuilder) InsertQuery() *dialects.InsertQuery { return b.q }

type updateQueryBuilder struct {
	q *dialects.UpdateQuery
}

func (b *updateQueryBuilder) UpdateQuery() *dialects.UpdateQuery { return b.q }

type deleteQueryBuilder struct {
	q *dialects.DeleteQuery
}

func (b *deleteQueryBuilder) DeleteQuery() *dialects.DeleteQuery { return b.q }

type createTableQueryBuilder struct {
	q *dialects.CreateTableQuery
}

func (b *createTableQueryBuilder) CreateTableQuery() *dialects.CreateTableQuery { return b.q }

type dropTableQueryBuilder struct {
	q *dialects.DropTableQuery
}

func (b *dropTableQueryBuilder) DropTableQuery() *dialects.DropTableQuery { return b.q }

type alterTableQueryBuilder struct {
	q *dialects.AlterTableQuery
}

func (b *alterTableQueryBuilder) AlterTableQuery() *dialects.AlterTableQuery { return b.q }

func TestSelectQuery(t *testing.T) {
	q := dialects.NewSelectQuery()
	q.From = "foos"
	q.Select.Columns = append(q.Select.Columns, dialects.Column{Column: "id"})

	s := dialects.NewSelect()
	assert.Empty(t, s.Columns)
	assert.False(t, s.Distinct)
}

func TestMultiQueryBuilder(t *testing.T) {
	d := sqlite.New()
	b := dialects.NewMultiQueryBuilder()
	assert.NotNil(t, b)

	b.AddAlterTableQuery(&dialects.AlterTableQuery{Table: "a", AddColumns: []dialects.ColumnDefinition{{Name: "x", Datatype: dialects.DataTypeString}}})
	b.AddCreateTableQuery(&dialects.CreateTableQuery{Table: "c", Columns: []dialects.ColumnDefinition{{Name: "id", Datatype: dialects.DataTypeInt32}}})
	b.AddDeleteQuery(&dialects.DeleteQuery{Table: "d"})
	b.AddDropTableQuery(&dialects.DropTableQuery{Table: "e"})
	b.AddInsertQuery(&dialects.InsertQuery{Table: "f", Values: []map[string]any{{"id": 1}}})
	b.AddSelectQuery(&dialects.SelectQuery{Select: dialects.NewSelect(), From: "h"})
	b.AddUpdateQuery(&dialects.UpdateQuery{Table: "g", Values: map[string]any{"id": 1}})

	queries := b.MultiQuery(nil)
	assert.Len(t, queries, 7)

	results, err := dialects.EncodeMultiQuery(d, queries)
	require.NoError(t, err)
	assert.Contains(t, results.SQL, "ALTER TABLE")
	assert.Contains(t, results.SQL, "CREATE TABLE")
	assert.Contains(t, results.SQL, "DELETE FROM")
	assert.Contains(t, results.SQL, "DROP TABLE")
	assert.Contains(t, results.SQL, "INSERT INTO")
	assert.Contains(t, results.SQL, "UPDATE")
}

func TestEncodeMultiQueryEmpty(t *testing.T) {
	d := sqlite.New()
	results, err := dialects.EncodeMultiQuery(d, []dialects.MultiQuery{})
	require.NoError(t, err)
	assert.Equal(t, "", results.SQL)
}

func TestEncodeMultiQueryUnknown(t *testing.T) {
	d := sqlite.New()
	results, err := dialects.EncodeMultiQuery(d, []dialects.MultiQuery{{}})
	require.NoError(t, err)
	assert.Equal(t, ";", results.SQL)
}
