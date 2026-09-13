package schema_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/database"
	"gosalusa.com/database/dialects"
	"gosalusa.com/database/schema"
	"gosalusa.com/internal/test"
)

func TestBlueprintColumnTypes(t *testing.T) {
	b := schema.NewBlueprint("types")
	require.Equal(t, b, b.GetBlueprint())
	assert.Equal(t, "types", b.TableName())

	assert.Equal(t, dialects.DataTypeText, b.Text("text").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeBoolean, b.Bool("bool").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeInt32, b.Int("int").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeInt8, b.Int8("int8").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeInt16, b.Int16("int16").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeInt64, b.Int64("int64").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeUInt32, b.UInt("uint").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeUInt8, b.UInt8("uint8").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeUInt16, b.UInt16("uint16").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeUInt32, b.UInt32("uint32").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeUInt64, b.UInt64("uint64").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeFloat32, b.Float("float").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeFloat32, b.Float32("float32").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeFloat64, b.Float64("float64").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeJSON, b.JSON("json").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeDate, b.Date("date").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeDateTime, b.DateTime("datetime").ColumnDefinition().Datatype)
	assert.Equal(t, dialects.DataTypeBlob, b.Blob("blob").ColumnDefinition().Datatype)
}

func TestBlueprintOfTypeAddColumn(t *testing.T) {
	b := schema.NewBlueprint("types")
	c := b.OfType(dialects.DataTypeInt32, "id")
	assert.Equal(t, "id", c.Name())
	assert.Same(t, b, b.AddColumn(c))
}

func TestBlueprintGoString(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		b := schema.NewBlueprint("foo")
		b.Int("id").Primary().AutoIncrement()
		b.String("name").Default("bob").Unique().Index()
		b.DropColumn("old")
		b.ForeignKey("foo_id", "bars", "id")
		src := b.GoString()
		assert.Contains(t, src, "table.Int(\"id\").Primary().AutoIncrement()")
		assert.Contains(t, src, "table.String(\"name\").Default(\"bob\").Unique()")
		assert.Contains(t, src, "table.DropColumn(\"old\")")
		assert.Contains(t, src, "table.ForeignKey(\"foo_id\", \"bars\", \"id\")")
		assert.NotContains(t, src, "PrimaryKey")
	})
	t.Run("composite primary key", func(t *testing.T) {
		b := schema.NewBlueprint("foo")
		b.PrimaryKey("a", "b")
		src := b.GoString()
		assert.Contains(t, src, "table.PrimaryKey(\"a\", \"b\")")
	})
}

func TestBlueprintMerge(t *testing.T) {
	t.Run("different name no-op", func(t *testing.T) {
		old := schema.NewBlueprint("old")
		old.String("a")
		old.Merge(schema.NewBlueprint("new"))
		assert.Equal(t, "old", old.TableName())
	})
	t.Run("add and change", func(t *testing.T) {
		b := schema.NewBlueprint("foo")
		b.String("name")

		nb := schema.NewBlueprint("foo")
		nb.String("name").Change().Nullable()
		nb.String("added")

		b.Merge(nb)
		src := b.GoString()
		assert.Contains(t, src, `String("added")`)
		assert.Contains(t, src, ".Nullable()")
	})
	t.Run("drop columns and primary keys", func(t *testing.T) {
		b := schema.NewBlueprint("foo")
		b.String("a")
		b.String("b")

		nb := schema.NewBlueprint("foo")
		nb.DropColumn("a")
		nb.PrimaryKey("b")

		b.Merge(nb)
		src := b.GoString()
		assert.Contains(t, src, `"b"`)
		assert.NotContains(t, src, `"a"`)
	})
}

func TestBlueprintUpdate(t *testing.T) {
	t.Run("add column", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		old.String("name")

		newBP := schema.NewBlueprint("foo")
		newBP.String("name")
		newBP.String("added")

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)

		assert.True(t, hasChanges)
		assert.Contains(t, update.GoString(), `String("added")`)
	})
	t.Run("no changes", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		old.String("name")

		newBP := schema.NewBlueprint("foo")
		newBP.String("name")

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)
		assert.False(t, hasChanges)
	})
	t.Run("change column", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		old.String("name")

		newBP := schema.NewBlueprint("foo")
		newBP.String("name").Change().Nullable()

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)
		assert.True(t, hasChanges)
		assert.Contains(t, update.GoString(), ".Nullable()")
	})
	t.Run("drop column", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		old.String("name")
		old.String("gone")

		newBP := schema.NewBlueprint("foo")
		newBP.String("name")

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)
		assert.True(t, hasChanges)
		assert.Contains(t, update.GoString(), `DropColumn("gone")`)
	})
	t.Run("add foreign key", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		newBP := schema.NewBlueprint("foo")
		newBP.ForeignKey("foo_id", "bars", "id")

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)
		assert.True(t, hasChanges)
		assert.Contains(t, update.GoString(), `ForeignKey("foo_id", "bars", "id")`)
	})
	t.Run("add index", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		newBP := schema.NewBlueprint("foo")
		newBP.Index("idx").AddColumn("name")

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)
		assert.True(t, hasChanges)
		assert.Contains(t, update.GoString(), `Index("idx")`)
	})
	t.Run("existing foreign key and index skipped", func(t *testing.T) {
		old := schema.NewBlueprint("foo")
		old.ForeignKey("foo_id", "bars", "id")
		old.Index("idx").AddColumn("name")
		old.String("name")

		newBP := schema.NewBlueprint("foo")
		newBP.ForeignKey("foo_id", "bars", "id")
		newBP.Index("idx").AddColumn("name")
		newBP.String("name")

		update := schema.NewBlueprint("foo")
		hasChanges := update.Update(old, newBP)
		assert.False(t, hasChanges)
	})
}

func TestColumnBuilderMethods(t *testing.T) {
	b := schema.NewColumn("foo", dialects.DataTypeString)
	assert.Equal(t, "foo", b.Name())
	assert.Equal(t, b, b.Nullable())
	assert.Equal(t, b, b.NotNullable())
	assert.Equal(t, b, b.Primary())
	assert.Equal(t, b, b.AutoIncrement())
	assert.Equal(t, b, b.After("other"))
	assert.Equal(t, b, b.Change())
	assert.Equal(t, b, b.Default(1))
	assert.Equal(t, b, b.Type(dialects.DataTypeInt32))
	assert.Equal(t, b, b.Unique())
	assert.Equal(t, b, b.DefaultCurrentTime())
	assert.Equal(t, b, b.Index())

	assert.True(t, schema.NewColumn("foo", dialects.DataTypeInt32).Equals(schema.NewColumn("foo", dialects.DataTypeInt32)))
	assert.False(t, schema.NewColumn("foo", dialects.DataTypeInt32).Equals(schema.NewColumn("bar", dialects.DataTypeInt32)))
	assert.False(t, schema.NewColumn("foo", dialects.DataTypeInt32).Equals(schema.NewColumn("foo", dialects.DataTypeString)))

	src := schema.NewColumn("foo", dialects.DataTypeString).
		Primary().
		AutoIncrement().
		Nullable().
		Default(1).
		DefaultCurrentTime().
		Unique().
		Index().
		Change().
		GoString()
	assert.Contains(t, src, ".Primary()")
	assert.Contains(t, src, ".AutoIncrement()")
	assert.Contains(t, src, ".Nullable()")
	assert.Contains(t, src, ".Default(1)")
	assert.Contains(t, src, ".Unique()")
	assert.Contains(t, src, ".DefaultCurrentTime()")
	assert.Contains(t, src, ".Index()")
	assert.Contains(t, src, ".Change()")

	plain := schema.NewColumn("foo", dialects.DataTypeString).GoString()
	assert.Equal(t, "", plain)
}

func TestCreateTableBuilder(t *testing.T) {
	b := schema.Create("foo", func(t *schema.Blueprint) {})
	assert.NotNil(t, b.GetBlueprint())
	assert.Equal(t, schema.BlueprintTypeCreate, b.Type())
	assert.Contains(t, b.GoString(), "schema.Create(\"foo\"")
	assert.Same(t, b, b.IfNotExists())

	c := schema.NewColumn("a", dialects.DataTypeString)
	assert.Same(t, b, b.Columns(c))
	assert.Same(t, b, b.AddColumns(c))
}

func TestUpdateTableBuilder(t *testing.T) {
	b := schema.Table("foo_new", func(t *schema.Blueprint) {})
	assert.NotNil(t, b.GetBlueprint())
	assert.Equal(t, schema.BlueprintTypeUpdate, b.Type())
	assert.Contains(t, b.GoString(), "schema.Table(")
	assert.Contains(t, b.GoString(), "foo")

	test.Run(t, "run", func(t *testing.T, tx *sqlx.Tx) {
		err := schema.Create("foo_new", func(t *schema.Blueprint) {
			t.Int("id")
		}).Run(context.Background(), tx)
		require.NoError(t, err)

		ub := schema.Table("foo_new", func(t *schema.Blueprint) {
			t.String("name")
		})
		err = ub.Run(context.Background(), tx)
		require.NoError(t, err)
	})
}

func TestDropTable(t *testing.T) {
	test.Run(t, "drop", func(t *testing.T, tx *sqlx.Tx) {
		err := schema.Create("dropme", func(t *schema.Blueprint) {
			t.Int("id")
		}).Run(context.Background(), tx)
		require.NoError(t, err)

		err = schema.Drop("dropme").Run(context.Background(), tx)
		assert.NoError(t, err)
	})
	test.Run(t, "drop if exists", func(t *testing.T, tx *sqlx.Tx) {
		err := schema.Create("dropme2", func(t *schema.Blueprint) {
			t.Int("id")
		}).Run(context.Background(), tx)
		require.NoError(t, err)

		err = schema.DropIfExists("dropme2").Run(context.Background(), tx)
		assert.NoError(t, err)

		err = schema.DropIfExists("dropme2").Run(context.Background(), tx)
		assert.NoError(t, err)
	})
}

func TestRunnerFunc(t *testing.T) {
	r := schema.Run(func(ctx context.Context, tx database.DB) error {
		return nil
	})
	err := r.Run(context.Background(), nil)
	assert.NoError(t, err)

	err = schema.Run(func(ctx context.Context, tx database.DB) error {
		return assert.AnError
	}).Run(context.Background(), nil)
	assert.Error(t, err)
}

func TestIndexBuilderUnique(t *testing.T) {
	b := schema.NewBlueprint("table")
	idx := b.Index("idx")
	idx.AddColumn("a")
	idx.AddColumn("b")
	idx.Unique()
	index := idx.Index()
	assert.Equal(t, "table", index.Table)
	assert.Equal(t, []string{"a", "b"}, index.Columns)
	assert.True(t, index.Unique)
	assert.Contains(t, idx.GoString(), `.AddColumn("a")`)
	assert.Contains(t, idx.GoString(), `.AddColumn("b")`)
	assert.Contains(t, idx.GoString(), ".Unique()")

	plain := b.Index("plain").GoString()
	assert.Equal(t, "", plain)
}
